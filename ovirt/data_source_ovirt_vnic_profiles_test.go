package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtVNicProfilesDataSource_nameRegexFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtVNicProfilesDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_vnic_profiles.name_regex_filtered_cluster"),
					resource.TestCheckResourceAttr("data.ovirt_vnic_profiles.name_regex_filtered_cluster", "vnic_profiles.#", "2"),
					resource.TestMatchResourceAttr("data.ovirt_vnic_profiles.name_regex_filtered_cluster", "vnic_profiles.0.name", regexp.MustCompile("^(no_)?mirror*")),
					resource.TestMatchResourceAttr("data.ovirt_vnic_profiles.name_regex_filtered_cluster", "vnic_profiles.1.name", regexp.MustCompile("^(no_)?mirror*")),
				),
			},
		},
	})
}

var testAccCheckOvirtVNicProfilesDataSourceNameRegexConfig = `
data "ovirt_networks" "search_filtered_network" {
  search = {
    criteria       = "datacenter = Default and name = ovirtmgmt-test"
    max            = 1
    case_sensitive = false
  }
}

data "ovirt_vnic_profiles" "name_regex_filtered_cluster" {
  name_regex = ".*mirror$"
  network_id = data.ovirt_networks.search_filtered_network.networks.0.id
}
`
