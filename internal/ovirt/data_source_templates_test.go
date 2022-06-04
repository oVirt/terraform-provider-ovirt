package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestTemplatesDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))

	config :=
		fmt.Sprintf(
			`
			provider "ovirt" {
				mock = true
			}
			
			resource "ovirt_vm" "test" {
				cluster_id  = "%s"
				template_id = "%s"
				name        = "test"
			}

            resource "ovirt_template" "test" {
                name  = "test"
                vm_id = ovirt_vm.test.id 
            }
			
			data "ovirt_templates" "list" {
				name          = ovirt_template.test.name
                fail_on_empty = true
			}

			output "template_ids" {
				value = data.ovirt_templates.list.templates.*.id
			}`,
			p.getTestHelper().GetClusterID(),
			p.getTestHelper().GetBlankTemplateID(),
		)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: func(s *terraform.State) error {
						templateID := s.RootModule().Resources["ovirt_template.test"].Primary.ID
						templateIDs := s.RootModule().Outputs["template_ids"].Value.([]interface{})
						for _, checkTemplateID := range templateIDs {
							if checkTemplateID == templateID {
								return nil
							}
						}
						return fmt.Errorf("failed to find previously created template using the ovirt_templates resource")
					},
				},
			},
		},
	)
}
