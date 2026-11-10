package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestDatacenterDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()

	dcs, err := helper.GetClient().ListDatacenters()
	if err != nil {
		t.Fatalf("failed to list datacenters: %v", err)
	}
	if len(dcs) == 0 {
		t.Fatalf("no datacenters found in test environment")
	}
	dc := dcs[0]

	configByName := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_datacenter" "test_by_name" {
			name = "%s"
		}

		output "datacenter_id" {
			value = data.ovirt_datacenter.test_by_name.id
		}
	`, dc.Name())

	configByID := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_datacenter" "test_by_id" {
			id = "%s"
		}

		output "datacenter_name" {
			value = data.ovirt_datacenter.test_by_id.name
		}
	`, dc.ID())

	configNotFound := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_datacenter" "test_not_found" {
			name = "non-existent-datacenter-xyz-123"
		}
	`

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: configByName,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputID := s.RootModule().Outputs["datacenter_id"].Value.(string)
							if outputID != string(dc.ID()) {
								return fmt.Errorf("expected datacenter ID %q, got %q", dc.ID(), outputID)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_datacenter.test_by_name", "id", string(dc.ID())),
						resource.TestCheckResourceAttr("data.ovirt_datacenter.test_by_name", "name", dc.Name()),
					),
				},
				{
					Config: configByID,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputName := s.RootModule().Outputs["datacenter_name"].Value.(string)
							if outputName != dc.Name() {
								return fmt.Errorf("expected datacenter name %q, got %q", dc.Name(), outputName)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_datacenter.test_by_id", "id", string(dc.ID())),
						resource.TestCheckResourceAttr("data.ovirt_datacenter.test_by_id", "name", dc.Name()),
					),
				},
				{
					Config:      configNotFound,
					ExpectError: regexp.MustCompile("not found"),
				},
			},
		},
	)
}
