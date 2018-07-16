package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtNetworksDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtNetworksDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_networks.name_regex_filtered_network"),
					resource.TestCheckResourceAttr("data.ovirt_networks.name_regex_filtered_network", "networks.#", "1"),
					resource.TestMatchResourceAttr("data.ovirt_networks.name_regex_filtered_network", "networks.0.name", regexp.MustCompile("^ovirtmgmt-*")),
				),
			},
		},
	})
}

func TestAccOvirtNetworksDataSource_searchFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtNetworksDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_networks.search_filtered_network"),
					resource.TestCheckResourceAttr("data.ovirt_networks.search_filtered_network", "networks.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_networks.search_filtered_network", "networks.0.name", "ovirtmgmt-test"),
				),
			},
		},
	})

}

var testAccCheckOvirtNetworksDataSourceNameRegexConfig = `
data "ovirt_networks" "name_regex_filtered_network" {
	name_regex = "^ovirtmgmt-t*"
  }
`

var testAccCheckOvirtNetworksDataSourceSearchConfig = `
data "ovirt_networks" "search_filtered_network" {
	search = {
	  criteria       = "datacenter = Default and name = ovirtmgmt-test"
	  max            = 1
	  case_sensitive = false
	}
}
`
