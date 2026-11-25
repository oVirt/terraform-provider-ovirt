package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestTagDataSource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	tagName := fmt.Sprintf("%s-%s", t.Name(), p.getTestHelper().GenerateRandomID(5))
	description := "Test tag description"

	config := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		resource "ovirt_tag" "test" {
			name        = "%s"
			description = "%s"
		}

		data "ovirt_tag" "by_name" {
			name = ovirt_tag.test.name
		}

		data "ovirt_tag" "by_id" {
			id = ovirt_tag.test.id
		}

		output "tag_id_by_name" {
			value = data.ovirt_tag.by_name.id
		}

		output "tag_name_by_id" {
			value = data.ovirt_tag.by_id.name
		}
	`, tagName, description)

	configNotFound := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_tag" "test_not_found" {
			name = "non-existent-tag-xyz"
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
							outputID := s.RootModule().Outputs["tag_id_by_name"].Value.(string)
							outputName := s.RootModule().Outputs["tag_name_by_id"].Value.(string)
							if outputID == "" {
								return fmt.Errorf("expected non-empty tag ID")
							}
							if outputName != tagName {
								return fmt.Errorf("expected tag name %q, got %q", tagName, outputName)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_tag.by_name", "name", tagName),
						resource.TestCheckResourceAttr("data.ovirt_tag.by_name", "description", description),
						resource.TestCheckResourceAttr("data.ovirt_tag.by_id", "name", tagName),
						resource.TestCheckResourceAttr("data.ovirt_tag.by_id", "description", description),
					),
				},
				{
					Config:      configNotFound,
					ExpectError: regexp.MustCompile("not found|Failed to get tag"),
				},
			},
		},
	)
}
