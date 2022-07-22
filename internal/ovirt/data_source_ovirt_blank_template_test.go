package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v2"
)

func TestBlankTemplateID(t *testing.T) {
	p := newProvider(newTestLogger(t))

	config :=
		`
			provider "ovirt" {
				mock = true
			}

			data "ovirt_blank_template" "blank" {
				
			}

			output "blank_template_id" {
				value = data.ovirt_blank_template.blank.id
			}`

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							v := s.RootModule().Outputs["blank_template_id"].Value.(string)
							tpl, err := p.getTestHelper().GetClient().GetTemplate(ovirtclient.TemplateID(v))
							if err != nil {
								return err
							}
							isBlank, err := tpl.IsBlank()
							if err != nil {
								return err
							}
							if !isBlank {
								return fmt.Errorf("template %s is not a blank template", tpl.ID())
							}
							return nil
						},
					),
				},
			},
		},
	)
}
