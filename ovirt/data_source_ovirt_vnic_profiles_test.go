package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtVNicProfilesDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtVNicProfilesDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_vnic_profiles.name_regex_filtered_cluster"),
					resource.TestCheckResourceAttr("data.ovirt_vnic_profiles.name_regex_filtered_cluster", "vnic_profiles.#", "2"),
					resource.TestMatchResourceAttr("data.ovirt_vnic_profiles.name_regex_filtered_cluster", "vnic_profiles.0.name", regexp.MustCompile("^mirror*")),
					resource.TestMatchResourceAttr("data.ovirt_vnic_profiles.name_regex_filtered_cluster", "vnic_profiles.1.name", regexp.MustCompile("^no_mirror*")),
				),
			},
		},
	})
}

var testAccCheckOvirtVNicProfilesDataSourceNameRegexConfig = `
data "ovirt_vnic_profiles" "name_regex_filtered_cluster" {
	name_regex = ".*mirror$"
	network_id = "649f2d61-7f23-477b-93bd-d55f974d8bc8"
  }
`
