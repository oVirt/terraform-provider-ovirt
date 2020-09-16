package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtNicsDataSource_nameRegexFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
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
data "ovirt_vms" "search_filtered_vm" {
  search = {
    criteria       = "name = HostedEngine and status = up"
    max            = 2
    case_sensitive = false
  }
}

data "ovirt_nics" "name_regex_filtered_nics" {
  name_regex = "^eth0*"
  vm_id = data.ovirt_vms.search_filtered_vm.vms.0.id
}
`
