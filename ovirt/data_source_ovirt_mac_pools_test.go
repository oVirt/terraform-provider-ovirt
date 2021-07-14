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

func TestAccOvirtMacPoolsDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)

	macPools, err := suite.GetMACPoolList()
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_mac_pools" "name_regex_filtered_pool" {
  name_regex = "\\A%s\\z"
  search = {
    max = 1
  }
}
`, regexp.QuoteMeta(macPools[0])),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.ovirt_mac_pools.name_regex_filtered_pool",
						"mac_pools.#",
						"1",
					),
					resource.TestMatchResourceAttr(
						"data.ovirt_mac_pools.name_regex_filtered_pool",
						"mac_pools.0.name",
						regexp.MustCompile(regexp.QuoteMeta(macPools[0])),
					),
				),
			},
		},
	})
}
