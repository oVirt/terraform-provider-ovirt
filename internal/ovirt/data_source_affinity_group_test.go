package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAffinityGroupDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()

	clusterID := helper.GetClusterID()
	affinityGroup, err := helper.GetClient().CreateAffinityGroup(clusterID, p.getTestHelper().GenerateTestResourceName(t), nil)
	if err != nil {
		t.Fatalf("Failed to create initial affinity group: %v", err)
	}

	config :=
		fmt.Sprintf(
			`
			provider "ovirt" {
				mock = true
			}
			
            data "ovirt_affinity_group" "main" {
				cluster_id = "%s"
				name = "%s"
			}
			
			output "main_affinity_id" {
				value = data.ovirt_affinity_group.main.id
			}`,
			clusterID,
			affinityGroup.Name(),
		)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: func(s *terraform.State) error {
						v := s.RootModule().Outputs["main_affinity_id"].Value.(string)
						if string(affinityGroup.ID()) != v {
							return fmt.Errorf("Expected affinity group ID %s, but got %s", affinityGroup.ID(), v)
						}

						return nil
					},
				},
			},
		},
	)
}
