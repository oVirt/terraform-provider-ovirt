package ovirt_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtVNicProfilesDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtVNicProfilesDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_vnic_profiles.name_regex_filtered_cluster"),
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
