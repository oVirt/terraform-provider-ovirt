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
					resource.TestMatchResourceAttr("data.ovirt_storagedomains.name_regex_filtered_storagedomain", "storagedomains.0.name", regexp.MustCompile("^test_ds*")),
					resource.TestMatchResourceAttr("data.ovirt_storagedomains.name_regex_filtered_storagedomain", "storagedomains.1.name", regexp.MustCompile("^test_ds*")),
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
					resource.TestCheckResourceAttr("data.ovirt_storagedomains.search_filtered_storagedomain", "storagedomains.0.name", "test_ds1"),
					testCheckResourceAttrNotEqual("data.ovirt_storagedomains.search_filtered_storagedomain", "storagedomains.0.size", true, 1024000000),
				),
			},
		},
	})

}

var TestAccOvirtStorageDomainsDataSourceNameRegexConfig = `
data "ovirt_storagedomains" "name_regex_filtered_storagedomain" {
	name_regex = "^test_ds*"
  }
`

var TestAccOvirtStorageDomainsDataSourceSearchConfig = `
data "ovirt_storagedomains" "search_filtered_storagedomain" {
	search = {
	  criteria       = "name = test_ds1 and datecenter = myDC"
	  max            = 1
	  case_sensitive = false
	}
  }
`
