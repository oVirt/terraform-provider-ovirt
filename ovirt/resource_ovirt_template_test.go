package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestTemplateResource(t *testing.T) {
	p := newProvider(newTestLogger(t))

	config :=
		fmt.Sprintf(`
			provider "ovirt" {
				mock = true
			}

			resource "ovirt_vm" "test" {
				cluster_id  = "%s"
				template_id = "%s"
				name        = "test"
			}

			resource "ovirt_disk" "test1" {
				storagedomain_id = "%s"
				format           = "raw"
				size             = 1048576
				alias            = "test1"
				sparse           = true
			}

			resource "ovirt_disk_attachments" "test" {
				vm_id          = ovirt_vm.test.id
			
				attachment {
					disk_id        = ovirt_disk.test1.id
					disk_interface = "virtio_scsi"
				}
			}
			
			resource "ovirt_template" "blueprint" {
				vm_id	= ovirt_vm.test.id
				name	= "blueprint1"
			}`,
			p.getTestHelper().GetClusterID(),
			p.getTestHelper().GetBlankTemplateID(),
			p.getTestHelper().GetStorageDomainID())

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						VMID := s.RootModule().Resources["ovirt_vm.test"].Primary.Attributes["id"]
						templateVMID := s.RootModule().Resources["ovirt_template.blueprint"].Primary.Attributes["vm_id"]
						if VMID != templateVMID {
							return fmt.Errorf("Template vm_id %s doesn't match the base VMs ID %s", templateVMID, VMID)
						}
						return nil
					},
				),
			},
		},
	})
}
