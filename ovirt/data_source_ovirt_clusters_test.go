package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtClustersDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtClustersDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_clusters.name_regex_filtered_cluster"),
					resource.TestCheckResourceAttr("data.ovirt_clusters.name_regex_filtered_cluster", "clusters.#", "2"),
					resource.TestMatchResourceAttr("data.ovirt_clusters.name_regex_filtered_cluster", "clusters.0.name", regexp.MustCompile("^Default*")),
					resource.TestMatchResourceAttr("data.ovirt_clusters.name_regex_filtered_cluster", "clusters.1.name", regexp.MustCompile("^Default*")),
				),
			},
		},
	})
}

func TestAccOvirtClustersDataSource_searchFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtClustersDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_clusters.search_filtered_cluster"),
					resource.TestCheckResourceAttr("data.ovirt_clusters.search_filtered_cluster", "clusters.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_clusters.search_filtered_cluster", "clusters.0.name", "Default"),
				),
			},
		},
	})

}

var testAccCheckOvirtClustersDataSourceNameRegexConfig = `
data "ovirt_clusters" "name_regex_filtered_cluster" {
	name_regex = "^Default*"
  }
`

var testAccCheckOvirtClustersDataSourceSearchConfig = `
data "ovirt_clusters" "search_filtered_cluster" {
	search = {
	  criteria       = "name = Default"
	  max            = 1
	  case_sensitive = false
	}
}
`
