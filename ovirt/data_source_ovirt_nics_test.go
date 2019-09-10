package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtNicsDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtNicsDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_nics.name_regex_filtered_nics"),
					resource.TestCheckResourceAttr("data.ovirt_nics.name_regex_filtered_nics", "nics.#", "1"),
					resource.TestMatchResourceAttr("data.ovirt_nics.name_regex_filtered_nics", "nics.0.name", regexp.MustCompile("^eth0*")),
					resource.TestCheckResourceAttrSet("data.ovirt_nics.name_regex_filtered_nics", "nics.0.reported_devices.0.ips.0.address"),
				),
			},
		},
	})
}

var testAccCheckOvirtNicsDataSourceNameRegexConfig = `
data "ovirt_nics" "name_regex_filtered_nics" {
  name_regex = "^eth0*"
  vm_id = "b9ea419c-7ce0-4508-8d04-8e75f60041ea"
}
`
