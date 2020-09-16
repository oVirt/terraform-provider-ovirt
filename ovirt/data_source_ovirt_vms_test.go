// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtVMsDataSource_nameRegexFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtVMsDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_vms.name_regex_filtered_vm"),
					resource.TestCheckResourceAttr("data.ovirt_vms.name_regex_filtered_vm", "vms.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_vms.name_regex_filtered_vm", "vms.0.name", "HostedEngine"),
				),
			},
		},
	})
}

func TestAccOvirtVMsDataSource_searchFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtVMsDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_vms.search_filtered_vm"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.0.name", "HostedEngine"),
				),
			},
		},
	})
}

var testAccCheckOvirtVMsDataSourceNameRegexConfig = `
data "ovirt_vms" "name_regex_filtered_vm" {
  name_regex = "\\w*ostedEn*"
}
`

var testAccCheckOvirtVMsDataSourceSearchConfig = `
data "ovirt_vms" "search_filtered_vm" {
  search = {
    criteria       = "name = HostedEngine and status = up"
    max            = 2
    case_sensitive = false
  }
}
`
