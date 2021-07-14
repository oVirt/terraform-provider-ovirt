package ovirt_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtVNicProfilesDataSource_nameRegexFilter(t *testing.T) {

	suite := getOvirtTestSuite(t)

	network, err := suite.CreateTestNetwork()
	if network != nil {
		defer func() {
			if err := suite.DeleteTestNetwork(network); err != nil {
				t.Fatal(fmt.Errorf("failed to delete test network (%w)", err))
			}
		}()
	}

	if err != nil {
		t.Fatal(err)
	}

	networkID := network.MustId()
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_vnic_profiles" "name_regex_filtered_cluster" {
  name_regex = "terraform-test.*"
  network_id = "%s"
}
`, networkID),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_vnic_profiles.name_regex_filtered_cluster"),
					resource.TestCheckResourceAttr("data.ovirt_vnic_profiles.name_regex_filtered_cluster", "vnic_profiles.#", "1"),
					resource.TestMatchResourceAttr("data.ovirt_vnic_profiles.name_regex_filtered_cluster", "vnic_profiles.0.name", regexp.MustCompile("^terraform-test*")),
				),
			},
		},
	})
}
