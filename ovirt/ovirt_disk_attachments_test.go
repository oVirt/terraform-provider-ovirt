package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

func TestDiskAttachmentsResource(t *testing.T) {
	t.Parallel()

	p := newProvider(ovirtclientlog.NewTestLogger(t))
	storageDomainID := p.getTestHelper().GetStorageDomainID()
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(
						`
provider "ovirt" {
	mock = true
}

resource "ovirt_disk" "test1" {
	storagedomain_id = "%s"
	format           = "raw"
    size             = 512
    alias            = "test"
    sparse           = true
}

resource "ovirt_disk" "test2" {
	storagedomain_id = "%s"
	format           = "raw"
    size             = 512
    alias            = "test"
    sparse           = true
}

resource "ovirt_vm" "test" {
	cluster_id  = "%s"
	template_id = "%s"
}

resource "ovirt_disk_attachments" "test" {
	vm_id          = ovirt_vm.test.id

	attachment {
		disk_id        = ovirt_disk.test1.id
		disk_interface = "virtio_scsi"
	}
	attachment {
		disk_id        = ovirt_disk.test2.id
		disk_interface = "virtio_scsi"
	}
}
`,
						storageDomainID,
						storageDomainID,
						clusterID,
						templateID,
					),
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_disk_attachments.test",
							"attachment.#",
							regexp.MustCompile("^2$"),
						),
					),
				},
				{
					Config: fmt.Sprintf(
						`
provider "ovirt" {
	mock = true
}

resource "ovirt_disk" "test1" {
	storagedomain_id = "%s"
	format           = "raw"
    size             = 512
    alias            = "test"
    sparse           = true
}

resource "ovirt_disk" "test2" {
	storagedomain_id = "%s"
	format           = "raw"
    size             = 512
    alias            = "test"
    sparse           = true
}

resource "ovirt_vm" "test" {
	cluster_id  = "%s"
	template_id = "%s"
}

resource "ovirt_disk_attachments" "test" {
	vm_id          = ovirt_vm.test.id

	attachment {
		disk_id        = ovirt_disk.test1.id
		disk_interface = "virtio_scsi"
	}
}
`,
						storageDomainID,
						storageDomainID,
						clusterID,
						templateID,
					),
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_disk_attachments.test",
							"attachment.#",
							regexp.MustCompile("^1$"),
						),
					),
				},
			},
		},
	)
}

func TestDiskAttachmentsResourceImport(t *testing.T) {
	t.Parallel()

	p := newProvider(ovirtclientlog.NewTestLogger(t))
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
	storagedomain_id = "%s"
	format           = "raw"
    size             = 512
    alias            = "test"
    sparse           = true
}

resource "ovirt_vm" "test" {
	cluster_id  = "%s"
	template_id = "%s"
}
`,
		storageDomainID,
		clusterID,
		templateID,
	)

	configPart2 := fmt.Sprintf(`
%s

resource "ovirt_disk_attachments" "test" {
	vm_id          = ovirt_vm.test.id
	attachment {
		disk_id        = ovirt_disk.test.id
		disk_interface = "virtio_scsi"
	}
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
					ResourceName: "ovirt_disk_attachments.test",
					ImportStateIdFunc: func(state *terraform.State) (string, error) {
						diskID := state.RootModule().Resources["ovirt_disk.test"].Primary.Attributes["id"]
						vmID := state.RootModule().Resources["ovirt_vm.test"].Primary.Attributes["id"]

						_, err := client.CreateDiskAttachment(
							vmID,
							diskID,
							ovirtclient.DiskInterfaceVirtIOSCSI,
							nil,
						)
						if err != nil {
							return "", fmt.Errorf("failed to create test disk attachment (%w)", err)
						}
						return vmID, nil
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_disk_attachments.test",
							"id",
							regexp.MustCompile("^.+$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_disk_attachments.test",
							"attachment.#",
							regexp.MustCompile("^1$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_disk_attachments.test",
							"vm_id",
							regexp.MustCompile("^.+$"),
						),
					),
				},
			},
		},
	)
}
