package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestDiskAttachmentsResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	storageDomainID := p.getTestHelper().GetStorageDomainID()
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
						provider "ovirt" {
							mock = true
						}

						resource "ovirt_disk" "test1" {
							storage_domain_id = "%s"
							format           = "raw"
							size             = 1048576
							alias            = "test"
							sparse           = true
						}

						resource "ovirt_disk" "test2" {
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

						resource "ovirt_disk_attachments" "test" {
							vm_id          = ovirt_vm.test.id

							attachment {
								disk_id        = ovirt_disk.test1.id
								disk_interface = "virtio_scsi"
								bootable       = true
								active         = true
							}
							attachment {
								disk_id        = ovirt_disk.test2.id
								disk_interface = "virtio_scsi"
								bootable       = null
								active         = null
							}
						}`,
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
						func(s *terraform.State) error {
							VMID := s.RootModule().Resources["ovirt_vm.test"].Primary.ID
							disk1ID := s.RootModule().Resources["ovirt_disk.test1"].Primary.ID
							disk2ID := s.RootModule().Resources["ovirt_disk.test2"].Primary.ID

							vm, err := p.getTestHelper().GetClient().GetVM(ovirtclient.VMID(VMID))
							if err != nil {
								return err
							}

							attachments, err := vm.ListDiskAttachments()
							if err != nil {
								return err
							}

							for _, attachment := range attachments {
								if string(attachment.DiskID()) == disk1ID {
									if !attachment.Active() || !attachment.Bootable() {
										return fmt.Errorf("Attachment for disk %s: Expected (bootable:%t,active:%t), but got (bootable:%t,active:%t)",
											disk1ID, true, true, attachments[0].Active(), attachments[0].Bootable())
									}
									continue
								}
								if string(attachment.DiskID()) == disk2ID {
									if attachment.Active() || attachment.Bootable() {
										return fmt.Errorf("Attachment for disk %s: Expected (bootable:%t,active:%t), but got (bootable:%t,active:%t)", disk2ID, false, false, attachments[1].Active(), attachments[1].Bootable())
									}
									continue
								}
								return fmt.Errorf("Unknown attachment found: %v", attachment)
							}

							return nil
						},
					),
				},
				{
					Config: fmt.Sprintf(`
						provider "ovirt" {
							mock = true
						}

						resource "ovirt_disk" "test1" {
							storage_domain_id = "%s"
							format           = "raw"
							size             = 1048576
							alias            = "test"
							sparse           = true
						}

						resource "ovirt_disk" "test2" {
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

						resource "ovirt_disk_attachments" "test" {
							vm_id          = ovirt_vm.test.id

							attachment {
								disk_id        = ovirt_disk.test1.id
								disk_interface = "virtio_scsi"
							}
						}`,
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
						func(s *terraform.State) error {
							VMID := s.RootModule().Resources["ovirt_vm.test"].Primary.ID

							vm, err := p.getTestHelper().GetClient().GetVM(ovirtclient.VMID(VMID))
							if err != nil {
								return err
							}

							attachments, err := vm.ListDiskAttachments()
							if err != nil {
								return err
							}

							if attachments[0].Active() || attachments[0].Bootable() {
								return fmt.Errorf("Expected (bootable:%t,active:%t), but got (bootable:%t,active:%t)", false, false, attachments[0].Active(), attachments[0].Bootable())
							}

							return nil
						},
					),
				},
			},
		},
	)
}

func TestDiskAttachmentsResourceImport(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	storageDomainID := p.getTestHelper().GetStorageDomainID()
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()
	client := p.getTestHelper().GetClient()

	configPart1 := fmt.Sprintf(`
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

	configPart2 := fmt.Sprintf(`
		%s

		resource "ovirt_disk_attachments" "test" {
			vm_id          = ovirt_vm.test.id
			attachment {
				disk_id        = ovirt_disk.test.id
				disk_interface = "virtio_scsi"
			}
		}`, configPart1)

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
							ovirtclient.VMID(vmID),
							ovirtclient.DiskID(diskID),
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
