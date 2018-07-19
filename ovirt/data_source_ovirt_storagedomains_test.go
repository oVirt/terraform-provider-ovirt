package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtStorageDomainsDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestAccOvirtStorageDomainsDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_storagedomains.name_regex_filtered_storagedomain"),
					resource.TestCheckResourceAttr("data.ovirt_storagedomains.name_regex_filtered_storagedomain", "storagedomains.#", "2"),
					resource.TestMatchResourceAttr("data.ovirt_storagedomains.name_regex_filtered_storagedomain", "storagedomains.0.name", regexp.MustCompile("^DEV_dat.*")),
					resource.TestMatchResourceAttr("data.ovirt_storagedomains.name_regex_filtered_storagedomain", "storagedomains.1.name", regexp.MustCompile("^MAIN_datastore*")),
				),
			},
		},
	})

}

func TestAccOvirtStorageDomainsDataSource_searchFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestAccOvirtStorageDomainsDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_storagedomains.search_filtered_storagedomain"),
					resource.TestCheckResourceAttr("data.ovirt_storagedomains.search_filtered_storagedomain", "storagedomains.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_storagedomains.search_filtered_storagedomain", "storagedomains.0.name", "DS_INTERNAL"),
					testCheckResourceAttrNotEqual("data.ovirt_storagedomains.search_filtered_storagedomain", "storagedomains.0.external_status", true, ""),
				),
			},
		},
	})

}

var TestAccOvirtStorageDomainsDataSourceNameRegexConfig = `
data "ovirt_storagedomains" "name_regex_filtered_storagedomain" {
	name_regex = "^MAIN_dat.*|^DEV_dat.*"
  }
`

var TestAccOvirtStorageDomainsDataSourceSearchConfig = `
data "ovirt_storagedomains" "search_filtered_storagedomain" {
	name_regex = "^DS_*"
	search = {
	  criteria       = "status != unattached and name = DS_INTERNAL and datacenter = MY_DC"
	  case_sensitive = false
	}
  }
`
