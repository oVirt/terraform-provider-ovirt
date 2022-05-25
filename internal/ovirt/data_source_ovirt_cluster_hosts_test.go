package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestClusterHostsDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))

	config :=
		fmt.Sprintf(`
			provider "ovirt" {
				mock = true
			}

			data "ovirt_cluster_hosts" "list" {
				cluster_id = "%s"
			}

			output "hosts_list" {
				value = data.ovirt_cluster_hosts.list
			}`,
			p.getTestHelper().GetClusterID())

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						v := s.RootModule().Outputs["hosts_list"].Value.(map[string]interface{})
						hosts, ok := v["hosts"]
						if !ok {
							return fmt.Errorf("missing key 'hosts' in output")
						}
						hostSize := len(hosts.([]interface{}))
						if hostSize != 1 {
							return fmt.Errorf("expected 1 hosts, but got only %d", hostSize)
						}
						return nil
					},
				),
			},
		},
	})
}
