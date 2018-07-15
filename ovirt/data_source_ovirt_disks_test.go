package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtDisksDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtDisksDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_disks.name_regex_filtered_disk"),
					resource.TestCheckResourceAttr("data.ovirt_disks.name_regex_filtered_disk", "disks.#", "2"),
					resource.TestMatchResourceAttr("data.ovirt_disks.name_regex_filtered_disk", "disks.0.name", regexp.MustCompile("^test_disk*")),
					resource.TestMatchResourceAttr("data.ovirt_disks.name_regex_filtered_disk", "disks.1.name", regexp.MustCompile("^test_disk*")),
				),
			},
		},
	})

}

func TestAccOvirtDisksDataSource_searchFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtDisksDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_disks.search_filtered_disk"),
					resource.TestCheckResourceAttr("data.ovirt_disks.search_filtered_disk", "disks.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_disks.search_filtered_disk", "disks.0.name", "test_disk1"),
					testCheckResourceAttrNotEqual("data.ovirt_disks.search_filtered_disk", "disks.0.size", true, 1024000000),
				),
			},
		},
	})

}

var testAccCheckOvirtDisksDataSourceNameRegexConfig = `
data "ovirt_disks" "name_regex_filtered_disk" {
	name_regex = "^test_disk*"
  }
`

var testAccCheckOvirtDisksDataSourceSearchConfig = `
data "ovirt_disks" "search_filtered_disk" {
	search = {
	  criteria       = "name = test_disk1 and provisioned_size > 1024000000"
	  max            = 1
	  case_sensitive = false
	}
  }
`
