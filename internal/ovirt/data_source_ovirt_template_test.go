package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestTemplateDataSource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	blankTemplateID := p.getTestHelper().GetBlankTemplateID()

	config := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_template" "by_name" {
			name = "Blank"
		}

		data "ovirt_template" "by_id" {
			id = "%s"
		}

		output "template_id_by_name" {
			value = data.ovirt_template.by_name.id
		}

		output "template_name_by_id" {
			value = data.ovirt_template.by_id.name
		}
	`, blankTemplateID)

	configNotFound := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_template" "not_found" {
			name = "non-existent-template-xyz-12345"
		}
	`

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputID := s.RootModule().Outputs["template_id_by_name"].Value.(string)
							outputName := s.RootModule().Outputs["template_name_by_id"].Value.(string)
							if outputID != string(blankTemplateID) {
								return fmt.Errorf("expected template ID %q, got %q", blankTemplateID, outputID)
							}
							if outputName != "Blank" {
								return fmt.Errorf("expected template name 'Blank', got %q", outputName)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_template.by_name", "id", string(blankTemplateID)),
						resource.TestCheckResourceAttr("data.ovirt_template.by_name", "name", "Blank"),
						resource.TestCheckResourceAttr("data.ovirt_template.by_id", "id", string(blankTemplateID)),
						resource.TestCheckResourceAttr("data.ovirt_template.by_id", "name", "Blank"),
					),
				},
				{
					Config:      configNotFound,
					ExpectError: regexp.MustCompile("not found|Failed to get template"),
				},
			},
		},
	)
}
