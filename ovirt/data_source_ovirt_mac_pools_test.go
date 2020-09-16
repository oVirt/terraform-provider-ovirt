// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtMacPoolsDataSource_nameRegexFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtMacPoolsDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_mac_pools.name_regex_filtered_pool"),
					resource.TestCheckResourceAttr("data.ovirt_mac_pools.name_regex_filtered_pool", "mac_pools.#", "1"),
					resource.TestMatchResourceAttr("data.ovirt_mac_pools.name_regex_filtered_pool", "mac_pools.0.name", regexp.MustCompile("\\w*efault*")),
				),
			},
		},
	})
}

var testAccCheckOvirtMacPoolsDataSourceNameRegexConfig = `
data "ovirt_mac_pools" "name_regex_filtered_pool" {
  name_regex = "\\w*efault*"
  search = {
    max = 1
  }
}
`
