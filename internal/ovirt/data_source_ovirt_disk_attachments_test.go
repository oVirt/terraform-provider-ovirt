package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestDiskAttachmentsDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))

	config :=
		fmt.Sprintf(`
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
				attachment {
					disk_id        = ovirt_disk.test2.id
					disk_interface = "virtio_scsi"
				}
			}
			
			data "ovirt_disk_attachments" "list" {
				vm_id = ovirt_vm.test.id
				depends_on = [
					ovirt_disk_attachments.test
				]
			}

			output "attachment_list" {
				value = data.ovirt_disk_attachments.list
			}`,
			p.getTestHelper().GetStorageDomainID(),
			p.getTestHelper().GetStorageDomainID(),
			p.getTestHelper().GetClusterID(),
			p.getTestHelper().GetBlankTemplateID())

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"ovirt_disk_attachments.test",
						"attachment.#",
						regexp.MustCompile("^2$"),
					),
					func(s *terraform.State) error {
						v := s.RootModule().Outputs["attachment_list"].Value.(map[string]interface{})
						attachments, ok := v["attachments"]
						if !ok {
							return fmt.Errorf("missing key 'attachments' in output")
						}
						attachmentSize := len(attachments.([]interface{}))
						if attachmentSize != 2 {
							return fmt.Errorf("expected 2 attachments, but got only %d", attachmentSize)
						}
						return nil
					},
				),
			},
		},
	})
}
