package ovirt

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtDataCentersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtDataCentersDataSourceBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_datacenters.name_filtered_datacenter"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.name_filtered_datacenter", "name", "Default"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.name_filtered_datacenter", "datacenters.#", "1"),
				),
			},
		},
	})
}

var testAccCheckOvirtDataCentersDataSourceBasicConfig = `
data "ovirt_datacenters" "name_filtered_datacenter" {
	name = "Default"
  }
`
