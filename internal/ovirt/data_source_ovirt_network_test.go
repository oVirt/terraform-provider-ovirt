package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestNetworkDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()

	networks, err := helper.GetClient().ListNetworks()
	if err != nil {
		t.Fatalf("failed to list networks from test helper: %v", err)
	}
	if len(networks) == 0 {
		t.Fatalf("no networks found in test environment")
	}
	network := networks[0]

	configByName := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_network" "test_by_name" {
			name = "%s"
		}

		output "network_id" {
			value = data.ovirt_network.test_by_name.id
		}

		output "network_datacenter_id" {
			value = data.ovirt_network.test_by_name.datacenter_id
		}
	`, network.Name())

	configByNameAndDC := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_network" "test_by_name_dc" {
			name          = "%s"
			datacenter_id = "%s"
		}

		output "network_id_dc" {
			value = data.ovirt_network.test_by_name_dc.id
		}
	`, network.Name(), network.DatacenterID())

	configByID := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_network" "test_by_id" {
			id = "%s"
		}

		output "network_name" {
			value = data.ovirt_network.test_by_id.name
		}
	`, network.ID())

	configNotFound := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_network" "test_not_found" {
			name = "non-existent-network"
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
							outputID := s.RootModule().Outputs["network_id"].Value.(string)
							if outputID != string(network.ID()) {
								return fmt.Errorf("expected network ID %q, got %q", network.ID(), outputID)
							}
							outputDC := s.RootModule().Outputs["network_datacenter_id"].Value.(string)
							if outputDC != string(network.DatacenterID()) {
								return fmt.Errorf("expected datacenter ID %q, got %q", network.DatacenterID(), outputDC)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_network.test_by_name", "id", string(network.ID())),
						resource.TestCheckResourceAttr("data.ovirt_network.test_by_name", "name", network.Name()),
					),
				},
				{
					Config: configByNameAndDC,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputID := s.RootModule().Outputs["network_id_dc"].Value.(string)
							if outputID != string(network.ID()) {
								return fmt.Errorf("expected network ID %q, got %q", network.ID(), outputID)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_network.test_by_name_dc", "id", string(network.ID())),
					),
				},
				{
					Config: configByID,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputName := s.RootModule().Outputs["network_name"].Value.(string)
							if outputName != network.Name() {
								return fmt.Errorf("expected network name %q, got %q", network.Name(), outputName)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_network.test_by_id", "id", string(network.ID())),
						resource.TestCheckResourceAttr("data.ovirt_network.test_by_id", "name", network.Name()),
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
