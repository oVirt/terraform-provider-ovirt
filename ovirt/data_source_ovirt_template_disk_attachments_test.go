package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestTemplateDiskAttachmentsDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))

	configBaseVM := fmt.Sprintf(`
		resource "ovirt_disk" "test1" {
			storagedomain_id = "%s"
			format           = "raw"
			size             = 1048576
			alias            = "test"
			sparse           = true
		}
		
		resource "ovirt_disk" "test2" {
			storagedomain_id = "%s"
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

			depends_on = [ovirt_vm.test, ovirt_disk.test1, ovirt_disk.test2]
		}
	`, p.getTestHelper().GetStorageDomainID(),
		p.getTestHelper().GetStorageDomainID(),
		p.getTestHelper().GetClusterID(),
		p.getTestHelper().GetBlankTemplateID())

	configTemplate := `
		resource "ovirt_template" "blueprint" {
			vm_id = ovirt_vm.test.id
			name = "blueprint"
			
			depends_on = [ovirt_disk_attachments.test]
		}`

	config :=
		fmt.Sprintf(`
			provider "ovirt" {
				mock = true
			}

			%s
			%s

			data "ovirt_template_disks" "list" {
				template_id = ovirt_template.blueprint.id

				depends_on = [ovirt_template.blueprint]
			}

			output "attachment_list" {
				value = data.ovirt_template_disks.list
			}`,
			configBaseVM, configTemplate)

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

						diskAttachments, ok := v["disk_attachments"].([]interface{})
						if !ok {
							return fmt.Errorf("missing key 'disk_attachments' in output")
						}

						if len(diskAttachments) != 2 {
							return fmt.Errorf("expected 2 disk attachments, but got only %d", len(diskAttachments))
						}

						return nil
					},
				),
			},
		},
	})
}
