package ovirt_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtTemplatesDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtTemplatesDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_templates.name_regex_filtered_template"),
					resource.TestCheckResourceAttr("data.ovirt_templates.name_regex_filtered_template", "templates.#", "2"),
					resource.TestMatchResourceAttr("data.ovirt_templates.name_regex_filtered_template", "templates.0.name", regexp.MustCompile("^centOS*")),
					resource.TestMatchResourceAttr("data.ovirt_templates.name_regex_filtered_template", "templates.1.name", regexp.MustCompile("^centOS*")),
				),
			},
		},
	})
}

func TestAccOvirtTemplatesDataSource_searchFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtTemplatesDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_templates.search_filtered_template"),
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
