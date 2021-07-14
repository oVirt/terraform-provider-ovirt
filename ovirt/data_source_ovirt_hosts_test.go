// Copyright (C) 2019 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtHostsDataSource_hostsList(t *testing.T) {
	suite := getOvirtTestSuite(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: `
data "ovirt_hosts" "name_regex_filtered_host" {
  name_regex = ".*"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.ovirt_hosts.name_regex_filtered_host",
						"hosts.#",
						fmt.Sprintf("%d", suite.GetHostCount()),
					),
				),
			},
		},
	})
}
