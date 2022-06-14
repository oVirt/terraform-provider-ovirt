package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func TestDiskAttachmentResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	storageDomainID := p.getTestHelper().GetStorageDomainID()
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()

	baseConfig := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		resource "ovirt_disk" "test" {
			storage_domain_id = "%s"
			format           = "raw"
			size             = 1048576
			alias            = "test"
			sparse           = true
		}

		resource "ovirt_vm" "test" {
			cluster_id  = "%s"
			template_id = "%s"
			name        = "test"
		}`,
		storageDomainID,
		clusterID,
		templateID,
	)

	testcases := []struct {
		name             string
		inputBootable    string
		inputActive      string
		expectedBootable bool
		expectedActive   bool
	}{
		{
			name:             "all set to true",
			inputBootable:    "true",
			inputActive:      "true",
			expectedBootable: true,
			expectedActive:   true,
		},
		{
			name:             "all set to false",
			inputBootable:    "false",
			inputActive:      "false",
			expectedBootable: false,
			expectedActive:   false,
		},
		{
			name:             "using defaults",
			inputBootable:    "null",
			inputActive:      "null",
			expectedBootable: false,
			expectedActive:   false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			resource.UnitTest(
				t, resource.TestCase{
					ProviderFactories: p.getProviderFactories(),
					Steps: []resource.TestStep{
						{
							Config: fmt.Sprintf(`
								%s

								resource "ovirt_disk_attachment" "test" {
									vm_id          = ovirt_vm.test.id
									disk_id        = ovirt_disk.test.id
									disk_interface = "virtio_scsi"
									bootable       = %s
									active         = %s
								}`,
								baseConfig,
								testcase.inputBootable,
								testcase.inputActive,
							),
							Check: resource.ComposeTestCheckFunc(
								resource.TestMatchResourceAttr(
									"ovirt_disk_attachment.test",
									"id",
									regexp.MustCompile("^.+$"),
								),
								func(s *terraform.State) error {
									VMID := s.RootModule().Resources["ovirt_vm.test"].Primary.ID
									diskAttachmentID := s.RootModule().Resources["ovirt_disk_attachment.test"].Primary.ID
									diskAttachment, err := p.getTestHelper().GetClient().GetDiskAttachment(ovirtclient.VMID(VMID), ovirtclient.DiskAttachmentID(diskAttachmentID))
									if err != nil {
										return err
									}
									if diskAttachment.DiskInterface() != "virtio_scsi" {
										return fmt.Errorf("Expected disk_interface 'virtio_scsi', but got '%s'", diskAttachment.DiskInterface())
									}
									if diskAttachment.Bootable() != testcase.expectedActive {
										return fmt.Errorf("Expected bootable to be %t, but got %t", testcase.expectedBootable, diskAttachment.Bootable())
									}
									if diskAttachment.Active() != testcase.expectedActive {
										return fmt.Errorf("Expected active to be %t, but got %t", testcase.expectedActive, diskAttachment.Active())
									}
									return nil
								},
							),
						},
					},
				},
			)
		})
	}
}

func TestDiskAttachmentResourceImport(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	storageDomainID := p.getTestHelper().GetStorageDomainID()
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()
	client := p.getTestHelper().GetClient()

	configPart1 := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

resource "ovirt_disk" "test" {
	storage_domain_id = "%s"
	format           = "raw"
    size             = 1048576
    alias            = "test"
    sparse           = true
}

resource "ovirt_vm" "test" {
	cluster_id  = "%s"
	template_id = "%s"
    name        = "test"
}
`,
		storageDomainID,
		clusterID,
		templateID,
	)

	configPart2 := fmt.Sprintf(`
%s

resource "ovirt_disk_attachment" "test" {
	vm_id          = ovirt_vm.test.id
	disk_id        = ovirt_disk.test.id
	disk_interface = "virtio_scsi"
}
`, configPart1)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: configPart1,
				},
				{
					Config:       configPart2,
					ImportState:  true,
					ResourceName: "ovirt_disk_attachment.test",
					ImportStateIdFunc: func(state *terraform.State) (string, error) {
						diskID := state.RootModule().Resources["ovirt_disk.test"].Primary.Attributes["id"]
						vmID := state.RootModule().Resources["ovirt_vm.test"].Primary.Attributes["id"]

						diskAttachment, err := client.CreateDiskAttachment(
							ovirtclient.VMID(vmID),
							ovirtclient.DiskID(diskID),
							ovirtclient.DiskInterfaceVirtIOSCSI,
							nil,
						)
						if err != nil {
							return "", fmt.Errorf("failed to create test disk attachment (%w)", err)
						}
						return fmt.Sprintf("%s/%s", vmID, diskAttachment.ID()), nil
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_disk_attachment.test",
							"id",
							regexp.MustCompile("^.+$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_disk_attachment.test",
							"disk_id",
							regexp.MustCompile("^.+$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_disk_attachment.test",
							"disk_interface",
							regexp.MustCompile("^.+$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_disk_attachment.test",
							"vm_id",
							regexp.MustCompile("^.+$"),
						),
					),
				},
			},
		},
	)
}
