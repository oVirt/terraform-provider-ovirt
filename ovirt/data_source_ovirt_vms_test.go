// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtVMsDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
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
	suite := getOvirtTestSuite(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtVMsDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_vms.search_filtered_vm"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.0.name", "HostedEngine"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.0.reported_devices.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.0.reported_devices.0.name", "eth0"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.0.reported_devices.0.ips.#", "3"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.0.reported_devices.0.ips.0.address", "10.1.111.64"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.0.reported_devices.0.ips.0.version", "v4"),
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
