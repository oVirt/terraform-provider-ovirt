// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// Copyright (C) 2018 Chunguang Wu <chokko@126.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.
package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

func TestAccOvirtGlusterVolume_basic(t *testing.T) {
	var glustervolume ovirtsdk4.GlusterVolume
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_gluster_volume.volume",
		CheckDestroy:  testAccCheckGlusterVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlusterVolumeBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtGlusterVolumeExists("ovirt_gluster_volume.volume", &glustervolume),
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "name", "testvolume"),
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "volume_type", "distribute"),
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "replica_count", "0"),
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "volume_status", "up"),
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "bricks.#", "6"),
				),
			},
			{
				Config: testAccGlusterVolumeBasicUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "name", "testvolume"),
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "volume_type", "distribute"),
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "replica_count", "0"),
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "volume_status", "down"),
					resource.TestCheckResourceAttr("ovirt_gluster_volume.volume", "bricks.#", "5"),
				),
			},
		},
	})
}

func testAccCheckGlusterVolumeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_gluster_volume" {
			continue
		}
		getResp, err := conn.SystemService().ClustersService().
			ClusterService(rs.Primary.Attributes["cluster_id"]).
			GlusterVolumesService().
			VolumeService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Volume(); ok {
			return fmt.Errorf("Volume %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtGlusterVolumeExists(n string, v *ovirtsdk4.GlusterVolume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Volume ID is set")
		}
		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().ClustersService().
			ClusterService(rs.Primary.Attributes["cluster_id"]).
			GlusterVolumesService().
			VolumeService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		volumeResponseVolume, ok := getResp.Volume()
		if ok {
			*v = *volumeResponseVolume
			return nil
		}
		return fmt.Errorf("Volume %s not exist", rs.Primary.ID)
	}
}

const testAccGlusterVolumeBasic = `
data "ovirt_clusters" "my-cluster" {
	search = {
			  criteria       = "name = Default"
			  max            = 1
			  case_sensitive = false
			}
	}
	
	data "ovirt_hosts" "my-host1" {
	search = {
			  criteria       = "name = host64"
			  max            = 1
			  case_sensitive = false
			}
	}
	
	
	data "ovirt_hosts" "my-host2" {
	search = {
			  criteria       = "name = host65"
			  max            = 1
			  case_sensitive = false
			}
	}
	
	resource "ovirt_gluster_volume" "volume" {
	  name = "testvolume"
	  cluster_id = "${data.ovirt_clusters.my-cluster.clusters.0.id}"
	  volume_type = "distribute"
	  replica_count = 0
	  volume_status = "up"
	  volume_rebalance = ""
	  bricks = [
				{
				 server_id = "${data.ovirt_hosts.my-host1.hosts.0.id}"
				 brick_dir="/tt1"
				},
				{
				 server_id = "${data.ovirt_hosts.my-host2.hosts.0.id}"
				 brick_dir="/tt2"
				},
				{
				 server_id = "${data.ovirt_hosts.my-host1.hosts.0.id}"
				 brick_dir="/tt3"
				},
				{
				 server_id = "${data.ovirt_hosts.my-host2.hosts.0.id}"
				 brick_dir="/tt4"
				},
				{
				 server_id = "${data.ovirt_hosts.my-host1.hosts.0.id}"
				 brick_dir="/tt5"
				},
				{
				 server_id = "${data.ovirt_hosts.my-host2.hosts.0.id}"
				 brick_dir="/tt6"
				}
	
			  ]
	}
	
`
const testAccGlusterVolumeBasicUpdate = `
data "ovirt_clusters" "my-cluster" {
	search = {
			  criteria       = "name = Default"
			  max            = 1
			  case_sensitive = false
			}
	}
	
	data "ovirt_hosts" "my-host1" {
	search = {
			  criteria       = "name = host64"
			  max            = 1
			  case_sensitive = false
			}
	}
	
	
	data "ovirt_hosts" "my-host2" {
	search = {
			  criteria       = "name = host65"
			  max            = 1
			  case_sensitive = false
			}
	}
	
	resource "ovirt_gluster_volume" "volume" {
	  name = "testvolume"
	  cluster_id = "${data.ovirt_clusters.my-cluster.clusters.0.id}"
	  volume_type = "distribute"
	  replica_count = 0
	  volume_status = "down"
	  volume_rebalance = ""
	  bricks = [
				{
				 server_id = "${data.ovirt_hosts.my-host1.hosts.0.id}"
				 brick_dir="/tt1"
				},
				{
				 server_id = "${data.ovirt_hosts.my-host2.hosts.0.id}"
				 brick_dir="/tt2"
				},
				{
				 server_id = "${data.ovirt_hosts.my-host1.hosts.0.id}"
				 brick_dir="/tt3"
				},
				{
				 server_id = "${data.ovirt_hosts.my-host2.hosts.0.id}"
				 brick_dir="/tt4"
				},
				{
				 server_id = "${data.ovirt_hosts.my-host1.hosts.0.id}"
				 brick_dir="/tt5"
				}
			  ]
	}
	
`
