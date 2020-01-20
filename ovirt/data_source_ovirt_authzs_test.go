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

func TestAccOvirtAuthzsDataSource_nameRegexFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtAuthzsDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_authzs.name_regex_filtered_authz"),
					resource.TestCheckResourceAttr("data.ovirt_authzs.name_regex_filtered_authz", "authzs.#", "1"),
					resource.TestMatchResourceAttr("data.ovirt_authzs.name_regex_filtered_authz", "authzs.0.name", regexp.MustCompile("^internal-*")),
				),
			},
		},
	})
}

var testAccCheckOvirtAuthzsDataSourceNameRegexConfig = `
data "ovirt_authzs" "name_regex_filtered_authz" {
  name_regex = "^internal-*"
  search {
    max = 1
  }
}
`
