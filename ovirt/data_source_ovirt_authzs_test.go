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

func TestAccOvirtAuthzsDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	authzName := regexp.QuoteMeta(suite.GetTestAuthzName())
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_authzs" "name_regex_filtered_authz" {
  name_regex = "^%s$"
  search = {
    max = 1
  }
}
`, authzName),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_authzs.name_regex_filtered_authz"),
					resource.TestCheckResourceAttr("data.ovirt_authzs.name_regex_filtered_authz", "authzs.#", "1"),
					resource.TestMatchResourceAttr(
						"data.ovirt_authzs.name_regex_filtered_authz",
						"authzs.0.name",
						regexp.MustCompile(fmt.Sprintf("^%s$", authzName)),
					),
				),
			},
		},
	})
}
