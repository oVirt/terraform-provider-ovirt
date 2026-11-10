package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestClusterDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()

	cluster, err := helper.GetClient().GetCluster(helper.GetClusterID())
	if err != nil {
		t.Fatalf("failed to get cluster from test helper: %v", err)
	}

	configByName := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_cluster" "test_by_name" {
			name = "%s"
		}

		output "cluster_id" {
			value = data.ovirt_cluster.test_by_name.id
		}
	`, cluster.Name())

	configByID := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_cluster" "test_by_id" {
			id = "%s"
		}

		output "cluster_name" {
			value = data.ovirt_cluster.test_by_id.name
		}
	`, cluster.ID())

	configNotFound := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_cluster" "test_not_found" {
			name = "non-existent-cluster"
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
							outputID := s.RootModule().Outputs["cluster_id"].Value.(string)
							if outputID != string(cluster.ID()) {
								return fmt.Errorf("expected cluster ID %q, got %q", cluster.ID(), outputID)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_cluster.test_by_name", "id", string(cluster.ID())),
						resource.TestCheckResourceAttr("data.ovirt_cluster.test_by_name", "name", cluster.Name()),
					),
				},
				{
					Config: configByID,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputName := s.RootModule().Outputs["cluster_name"].Value.(string)
							if outputName != cluster.Name() {
								return fmt.Errorf("expected cluster name %q, got %q", cluster.Name(), outputName)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_cluster.test_by_id", "id", string(cluster.ID())),
						resource.TestCheckResourceAttr("data.ovirt_cluster.test_by_id", "name", cluster.Name()),
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
