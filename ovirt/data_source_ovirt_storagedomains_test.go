// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtStorageDomainsDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: TestAccOvirtStorageDomainsDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_storagedomains.name_regex_filtered_storagedomain"),
					resource.TestCheckResourceAttr("data.ovirt_storagedomains.name_regex_filtered_storagedomain", "storagedomains.#", "2"),
					resource.TestMatchResourceAttr("data.ovirt_storagedomains.name_regex_filtered_storagedomain", "storagedomains.0.name", regexp.MustCompile("^DEV_dat.*")),
					resource.TestMatchResourceAttr("data.ovirt_storagedomains.name_regex_filtered_storagedomain", "storagedomains.1.name", regexp.MustCompile("^MAIN_datastore*")),
				),
			},
		},
	})

}

func TestAccOvirtStorageDomainsDataSource_searchFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: TestAccOvirtStorageDomainsDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_storagedomains.search_filtered_storagedomain"),
					resource.TestCheckResourceAttr("data.ovirt_storagedomains.search_filtered_storagedomain", "storagedomains.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_storagedomains.search_filtered_storagedomain", "storagedomains.0.name", "DS_INTERNAL"),
					suite.TestResourceAttrNotEqual("data.ovirt_storagedomains.search_filtered_storagedomain", "storagedomains.0.external_status", true, ""),
				),
			},
		},
	})

}

var TestAccOvirtStorageDomainsDataSourceNameRegexConfig = `
data "ovirt_storagedomains" "name_regex_filtered_storagedomain" {
  name_regex = "^MAIN_dat.*|^DEV_dat.*"
}
`

var TestAccOvirtStorageDomainsDataSourceSearchConfig = `
data "ovirt_storagedomains" "search_filtered_storagedomain" {
  name_regex = "^DS_*"
  search = {
    criteria       = "status != unattached and name = DS_INTERNAL and datacenter = MY_DC"
    case_sensitive = false
  }
}
`
