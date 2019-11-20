package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtTemplatesDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtTemplatesDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_templates.name_regex_filtered_template"),
					resource.TestCheckResourceAttr("data.ovirt_templates.name_regex_filtered_template", "templates.#", "2"),
					resource.TestMatchResourceAttr("data.ovirt_templates.name_regex_filtered_template", "templates.0.name", regexp.MustCompile("^centOS*")),
					resource.TestMatchResourceAttr("data.ovirt_templates.name_regex_filtered_template", "templates.1.name", regexp.MustCompile("^centOS*")),
				),
			},
		},
	})
}

func TestAccOvirtTemplatesDataSource_searchFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtTemplatesDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_templates.search_filtered_template"),
					resource.TestCheckResourceAttr("data.ovirt_templates.search_filtered_template", "templates.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_templates.search_filtered_template", "templates.0.name", "centOST"),
				),
			},
		},
	})

}

var testAccCheckOvirtTemplatesDataSourceNameRegexConfig = `
data "ovirt_templates" "name_regex_filtered_template" {
  name_regex = "^centOS*"
}
`

var testAccCheckOvirtTemplatesDataSourceSearchConfig = `
data "ovirt_templates" "search_filtered_template" {
  search = {
    criteria       = "name = centOST"
    max            = 2
    case_sensitive = false
  }
}
`
