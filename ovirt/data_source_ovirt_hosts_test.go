// Copyright (C) 2018 Chunguang Wu <chokko@126.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOvirtHostsDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtHostsDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_hosts.name_regex_filtered_host"),
					resource.TestCheckResourceAttr("data.ovirt_hosts.name_regex_filtered_host", "hosts.#", "2"),
					resource.TestMatchResourceAttr("data.ovirt_hosts.name_regex_filtered_host", "hosts.0.name", regexp.MustCompile("^host*")),
					resource.TestMatchResourceAttr("data.ovirt_hosts.name_regex_filtered_host", "hosts.1.name", regexp.MustCompile("^host*")),
				),
			},
		},
	})
}

func TestAccOvirthostsDataSource_searchFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtHostsDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_hosts.search_filtered_host"),
					resource.TestCheckResourceAttr("data.ovirt_hosts.search_filtered_host", "hosts.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_hosts.search_filtered_host", "hosts.0.name", "host64"),
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
			  criteria       = "name = host64"
			  max            = 1
			  case_sensitive = false
			}
	}
`
