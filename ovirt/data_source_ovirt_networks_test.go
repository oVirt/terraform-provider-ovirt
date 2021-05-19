// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.
package ovirt_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtNetworksDataSource_nameRegexFilter(t *testing.T) {
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

	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_networks" "name_regex_filtered_network" {
  name_regex = "^%s$"
}
`, network.MustName()),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_networks.name_regex_filtered_network"),
					resource.TestCheckResourceAttr("data.ovirt_networks.name_regex_filtered_network", "networks.#", "1"),
					resource.TestMatchResourceAttr(
						"data.ovirt_networks.name_regex_filtered_network",
						"networks.0.name",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(network.MustName())))),
				),
			},
		},
	})
}

func TestAccOvirtNetworksDataSource_searchFilter(t *testing.T) {
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
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_networks" "search_filtered_network" {
  search = {
    criteria       = "datacenter = %s and name = %s"
    max            = 1
    case_sensitive = false
  }
}
`, suite.GetTestDatacenterName(), network.MustName()),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_networks.search_filtered_network"),
					resource.TestCheckResourceAttr("data.ovirt_networks.search_filtered_network", "networks.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_networks.search_filtered_network", "networks.0.name", network.MustName()),
				),
			},
		},
	})
}
