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

func TestAccOvirtDisksDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	testDisk, err := suite.CreateDisk()
	if err != nil {
		t.Fatal(err)
	}
	if testDisk != nil {
		defer func() {
			err := suite.RemoveDisk(testDisk)
			if err != nil {
				t.Fatal(fmt.Errorf("failed to remove disk after test (%w)", err))
			}
		}()
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_disks" "name_regex_filtered_disk" {
  name_regex = "^%s$"
}
`, regexp.QuoteMeta(testDisk.MustName())),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_disks.name_regex_filtered_disk"),
					resource.TestCheckResourceAttr(
						"data.ovirt_disks.name_regex_filtered_disk",
						"disks.#",
						"1",
					),
					resource.TestMatchResourceAttr(
						"data.ovirt_disks.name_regex_filtered_disk",
						"disks.0.name",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(testDisk.MustName())))),
				),
			},
		},
	})
}

func TestAccOvirtDisksDataSource_searchFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	testDisk, err := suite.CreateDisk()
	if err != nil {
		t.Fatal(err)
	}
	if testDisk != nil {
		defer func() {
			err := suite.RemoveDisk(testDisk)
			if err != nil {
				t.Fatal(fmt.Errorf("failed to remove disk after test (%w)", err))
			}
		}()
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_disks" "search_filtered_disk" {
  search = {
    criteria       = "name=%s"
    # Max has a bug in the oVirt API, see BZ 1962177
    # max = 1
    case_sensitive = false
  }
}
`, testDisk.MustName()),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_disks.search_filtered_disk"),
					resource.TestCheckResourceAttr("data.ovirt_disks.search_filtered_disk", "disks.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_disks.search_filtered_disk", "disks.0.name", testDisk.MustName()),
					resource.TestCheckResourceAttr(
						"data.ovirt_disks.search_filtered_disk",
						"disks.0.size",
						fmt.Sprintf("%d", testDisk.MustProvisionedSize()),
					),
				),
			},
		},
	})
}
