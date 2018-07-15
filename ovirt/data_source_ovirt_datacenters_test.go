package ovirt

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtDataCentersDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtDataCentersDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_datacenters.name_regex_filtered_datacenter"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.name_regex_filtered_datacenter", "datacenters.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.name_regex_filtered_datacenter", "datacenters.0.name", "Default"),
				),
			},
		},
	})
}

func TestAccOvirtDataCentersDataSource_searchFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtDataCentersDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_datacenters.search_filtered_datacenter"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.search_filtered_datacenter", "datacenters.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.search_filtered_datacenter", "datacenters.0.name", "Default"),
				),
			},
		},
	})
}

var testAccCheckOvirtDataCentersDataSourceNameRegexConfig = `
data "ovirt_datacenters" "name_regex_filtered_datacenter" {
	name_regex = "^Defa*"
  }
`

var testAccCheckOvirtDataCentersDataSourceSearchConfig = `
data "ovirt_datacenters" "search_filtered_datacenter" {
	search = {
	  criteria       = "name = Default and status = up and Storage.name = data"
	  max            = 2
	  case_sensitive = false
	}
  }
`
