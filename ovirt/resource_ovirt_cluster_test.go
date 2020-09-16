// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func TestAccOvirtCluster_basic(t *testing.T) {
	datacenterName := "datacenter2"
	var cluster ovirtsdk4.Cluster
	const resourceName = "ovirt_cluster.cluster"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: resourceName,
		CheckDestroy:  testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterBasic(datacenterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "name", "testAccOvirtClusterBasic"),
					resource.TestCheckResourceAttr(resourceName, "ballooning", "true"),
					resource.TestCheckResourceAttr(resourceName, "gluster", "true"),
					resource.TestCheckResourceAttr(resourceName, "cpu_arch", "x86_64"),
					resource.TestCheckResourceAttr(resourceName, "compatibility_version", "4.4"),

				),
			},
			{
				Config: testAccClusterBasicUpdate(datacenterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "name", "testAccOvirtClusterBasicUpdate"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "ballooning", "false"),
					resource.TestCheckResourceAttr(resourceName, "gluster", "false"),
				),
			},
		},
	})
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_cluster" {
			continue
		}
		getResp, err := conn.SystemService().ClustersService().
			ClusterService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Cluster(); ok {
			return fmt.Errorf("Cluster %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtClusterExists(n string, v *ovirtsdk4.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cluster ID is set")
		}
		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().ClustersService().
			ClusterService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		cluster, ok := getResp.Cluster()
		if ok {
			*v = *cluster
			return nil
		}
		return fmt.Errorf("Cluster %s not exist", rs.Primary.ID)
	}
}

func testAccClusterDef(datacenterName string) string {
	return fmt.Sprintf(`
data "ovirt_datacenters" "search_filtered_datacenter" {
  search = {
    criteria       = "name = %s"
    max            = 2
    case_sensitive = false
  }
}

data "ovirt_networks" "search_filtered_network" {
  search = {
    criteria       = "datacenter = %s and name = ovirtmgmt"
    max            = 1
    case_sensitive = false
  }
}

locals {
  datacenter_id = data.ovirt_datacenters.search_filtered_datacenter.datacenters.0.id
  network_id = data.ovirt_networks.search_filtered_network.networks.0.id
}
`, datacenterName, datacenterName)
}

func testAccClusterBasic(datacenterName string) string {
	return testAccClusterDef(datacenterName) + `
resource "ovirt_cluster" "cluster" {
  name                              = "testAccOvirtClusterBasic"
  description                       = "Desc of cluster"
  datacenter_id                     = local.datacenter_id
  management_network_id             = local.network_id
  memory_policy_over_commit_percent = 100
  ballooning                        = true
  gluster                           = true
  threads_as_cores                  = true
  cpu_arch                          = "x86_64"
  cpu_type                          = "Intel SandyBridge Family"
  compatibility_version             = "4.4"
}`
}

func testAccClusterBasicUpdate(datacenterName string) string {
	return testAccClusterDef(datacenterName) + `
resource "ovirt_cluster" "cluster" {
  name                              = "testAccOvirtClusterBasicUpdate"
  datacenter_id                     = local.datacenter_id
  management_network_id             = local.network_id
  memory_policy_over_commit_percent = 100
  ballooning                        = false
  gluster                           = false
  threads_as_cores                  = true
  cpu_arch                          = "x86_64"
  cpu_type                          = "Intel SandyBridge Family"
  compatibility_version             = "4.4"
}
`
}
