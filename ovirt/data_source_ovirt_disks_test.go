package ovirt

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtDisksDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtDisksDataSourceBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_disks.name_filtered_disk"),
					resource.TestCheckResourceAttr("data.ovirt_disks.name_filtered_disk", "name", "aaaa"),
					resource.TestCheckResourceAttr("data.ovirt_disks.name_filtered_disk", "disks.#", "1"),
				),
			},
		},
	})

}

var testAccCheckOvirtDisksDataSourceBasicConfig = `
data "ovirt_disks" "name_filtered_disk" {
	name = "aaaa"
  }
`
