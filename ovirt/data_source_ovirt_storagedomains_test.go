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

func TestAccOvirtStorageDomainsDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	storageDomain := suite.StorageDomain()
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`data "ovirt_storagedomains" "name_regex_filtered_storagedomain" {
  name_regex = "^%s$"
}`, regexp.QuoteMeta(storageDomain.MustName())),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_storagedomains.name_regex_filtered_storagedomain"),
					resource.TestCheckResourceAttr(
						"data.ovirt_storagedomains.name_regex_filtered_storagedomain",
						"storagedomains.#",
						"1",
					),
					resource.TestMatchResourceAttr(
						"data.ovirt_storagedomains.name_regex_filtered_storagedomain",
						"storagedomains.0.name",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(storageDomain.MustName())))),
				),
			},
		},
	})
}

func TestAccOvirtStorageDomainsDataSource_searchFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	storageDomain := suite.StorageDomain()
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_storagedomains" "search_filtered_storagedomain" {
  search = {
    criteria       = "status != unattached and name = %s"
    case_sensitive = false
  }
}
`, storageDomain.MustName()),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_storagedomains.search_filtered_storagedomain"),
					resource.TestCheckResourceAttr(
						"data.ovirt_storagedomains.search_filtered_storagedomain",
						"storagedomains.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"data.ovirt_storagedomains.search_filtered_storagedomain",
						"storagedomains.0.name",
						storageDomain.MustName(),
					),
				),
			},
		},
	})
}
