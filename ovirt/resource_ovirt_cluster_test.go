// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	govirt "github.com/ovirt/go-ovirt-client"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

//TODO fix this test
func DisabledTestAccOvirtCluster_basic(t *testing.T) {
	suite := getOvirtTestSuite(t)

	network, err := suite.CreateTestNetwork()
	if network != nil {
		defer func() {
			if err := suite.DeleteTestNetwork(network); err != nil {
				t.Fatal(fmt.Errorf("failed to delete test network (%w)", err))
			}
		}()
	}

	if err != nil {
		t.Fatal(err)
	}

	datacenterID := suite.GetTestDatacenterID()
	networkID := network.MustId()
	clusterName := fmt.Sprintf("tf-test-%s", suite.GenerateRandomID(5))
	updateClusterName := fmt.Sprintf("tf-test-upd-%s", suite.GenerateRandomID(5))

	var cluster ovirtsdk4.Cluster
	resource.Test(t, resource.TestCase{
		PreCheck:      suite.PreCheck,
		Providers:     suite.Providers(),
		IDRefreshName: "ovirt_cluster.cluster",
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "ovirt_cluster" "cluster" {
  name                              = "%s"
  description                       = "Desc of cluster"
  datacenter_id                     = "%s"
  management_network_id             = "%s"
  memory_policy_over_commit_percent = 100
  ballooning                        = true
  gluster                           = true
  threads_as_cores                  = true
  cpu_arch                          = "x86_64"
  cpu_type                          = "Intel SandyBridge Family"
  compatibility_version             = "4.4"
}
`, clusterName, datacenterID, networkID),
				Check: resource.ComposeTestCheckFunc(
					suite.EnsureCluster("ovirt_cluster.cluster", &cluster),
					resource.TestCheckResourceAttr("ovirt_cluster.cluster", "name", clusterName),
					resource.TestCheckResourceAttr("ovirt_cluster.cluster", "datacenter_id", datacenterID),
					resource.TestCheckResourceAttr("ovirt_cluster.cluster", "management_network_id", networkID),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "ovirt_cluster" "cluster" {
  name                              = "%s"
  datacenter_id                     = "%s"
  management_network_id             = "%s"
  memory_policy_over_commit_percent = 100
  ballooning                        = true
  gluster                           = true
  threads_as_cores                  = true
  cpu_arch                          = "x86_64"
  cpu_type                          = "Intel SandyBridge Family"
  compatibility_version             = "4.4"
}
`, updateClusterName, datacenterID, networkID),
				Check: resource.ComposeTestCheckFunc(
					suite.EnsureCluster("ovirt_cluster.cluster", &cluster),
					resource.TestCheckResourceAttr("ovirt_cluster.cluster", "name", updateClusterName),
					resource.TestCheckResourceAttr("ovirt_cluster.cluster", "datacenter_id", datacenterID),
					resource.TestCheckResourceAttr("ovirt_cluster.cluster", "management_network_id", networkID),
					resource.TestCheckResourceAttr("ovirt_cluster.cluster", "description", ""),
				),
			},
		},
		CheckDestroy:  suite.TestClusterDestroy(&cluster),
	})
}

// Deprecated: use suite instead.
func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
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

// Deprecated: use suite instead.
func testAccCheckOvirtClusterExists(n string, v *ovirtsdk4.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cluster ID is set")
		}
		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
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
