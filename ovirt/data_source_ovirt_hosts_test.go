// Copyright (C) 2019 Joey Ma <majunjiev@gmail.com>
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

func TestAccOvirtHostsDataSource_nameRegexFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtHostsDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_hosts.name_regex_filtered_host"),
					resource.TestCheckResourceAttr("data.ovirt_hosts.name_regex_filtered_host", "hosts.#", "1"),
					resource.TestMatchResourceAttr("data.ovirt_hosts.name_regex_filtered_host", "hosts.0.name", regexp.MustCompile("^host*")),
				),
			},
		},
	})
}

func TestAccOvirtHostsDataSource_searchFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtHostsDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_hosts.search_filtered_host"),
					resource.TestCheckResourceAttr("data.ovirt_hosts.search_filtered_host", "hosts.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_hosts.search_filtered_host", "hosts.0.name", "host65"),
				),
			},
		},
	})

}

var testAccCheckOvirtHostsDataSourceNameRegexConfig = `
data "ovirt_hosts" "name_regex_filtered_host" {
  name_regex = "^host*"
}
`

var testAccCheckOvirtHostsDataSourceSearchConfig = `
data "ovirt_hosts" "search_filtered_host" {
  search = {
    criteria       = "name = host65"
    max            = 1
    case_sensitive = false
  }
}
`
