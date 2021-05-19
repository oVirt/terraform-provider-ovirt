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

func TestAccOvirtClustersDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_clusters" "name_regex_filtered_cluster" {
  name_regex = "^%s$"
}
`, suite.Cluster().MustName()),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_clusters.name_regex_filtered_cluster"),
					resource.TestMatchResourceAttr("data.ovirt_clusters.name_regex_filtered_cluster", "clusters.0.name", regexp.MustCompile(fmt.Sprintf("^%s$", suite.Cluster().MustName()))),
				),
			},
		},
	})
}

func TestAccOvirtClustersDataSource_searchFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ovirt_clusters" "search_filtered_cluster" {
  search = {
    criteria       = "name = %s"
    max            = 1
    case_sensitive = false
  }
}
`, suite.Cluster().MustName()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataSourceID("data.ovirt_clusters.search_filtered_cluster"),
					resource.TestCheckResourceAttr("data.ovirt_clusters.search_filtered_cluster", "clusters.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_clusters.search_filtered_cluster", "clusters.0.name", suite.Cluster().MustName()),
				),
			},
		},
	})
}
