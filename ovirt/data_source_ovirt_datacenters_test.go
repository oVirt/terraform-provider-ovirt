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

func TestAccOvirtDataCentersDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	datacenterName := suite.GetTestDatacenterName()
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_datacenters" "name_regex_filtered_datacenter" {
  name_regex = "^%s$"
}
`,
					regexp.QuoteMeta(datacenterName),
				),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_datacenters.name_regex_filtered_datacenter"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.name_regex_filtered_datacenter", "datacenters.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.name_regex_filtered_datacenter", "datacenters.0.name", datacenterName),
				),
			},
		},
	})
}

func TestAccOvirtDataCentersDataSource_searchFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	datacenterName := suite.GetTestDatacenterName()
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_datacenters" "search_filtered_datacenter" {
  search = {
    criteria       = "name = %s and status = up"
    max            = 1
    case_sensitive = false
  }
}
`, datacenterName),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_datacenters.search_filtered_datacenter"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.search_filtered_datacenter", "datacenters.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_datacenters.search_filtered_datacenter", "datacenters.0.name", datacenterName),
				),
			},
		},
	})
}
