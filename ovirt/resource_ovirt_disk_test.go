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

func TestAccOvirtDisk_basic(t *testing.T) {
	var disk ovirtsdk4.Disk
	const resourceName = "ovirt_disk.disk"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: resourceName,
		CheckDestroy:  testAccCheckDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskExists(resourceName, &disk),
					resource.TestCheckResourceAttr(resourceName, "name", "testAccDiskBasic"),
					resource.TestCheckResourceAttr(resourceName, "alias", "testAccDiskBasic"),
					resource.TestCheckResourceAttr(resourceName, "size", "2"),
					resource.TestCheckResourceAttr(resourceName, "format", "cow"),
				),
			},
			{
				Config: testAccDiskBasicUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskExists(resourceName, &disk),
					resource.TestCheckResourceAttr(resourceName, "name", "testAccDiskBasicUpdate"),
					resource.TestCheckResourceAttr(resourceName, "alias", "testAccDiskBasicUpdate"),
					resource.TestCheckResourceAttr(resourceName, "size", "3"),
					resource.TestCheckResourceAttr(resourceName, "format", "cow"),
				),
			},
		},
	})
}

func testAccCheckDiskDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_disk" {
			continue
		}
		getResp, err := conn.SystemService().DisksService().
			DiskService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Disk(); ok {
			return fmt.Errorf("Disk %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtDiskExists(n string, v *ovirtsdk4.Disk) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Disk ID is set")
		}
		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().DisksService().
			DiskService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		disk, ok := getResp.Disk()
		if ok {
			*v = *disk
			return nil
		}
		return fmt.Errorf("Disk %s not exist", rs.Primary.ID)
	}
}

func testAccDiskBasicDef() string {
	return `
data "ovirt_clusters" "c" {
  search = {
    criteria = "name = Default"
  }
}

data "ovirt_storagedomains" "s" {
  search = {
    criteria = "datacenter = Default and name = data"
  }
}

locals {
  cluster_id        = data.ovirt_clusters.c.clusters.0.id
  storage_domain_id = data.ovirt_storagedomains.s.storagedomains.0.id
}

resource "ovirt_vm" "vm" {
  name        = "testAccVM"
  cluster_id  = local.cluster_id
  memory 	  = 1024 
  os {
    type = "other"
  }

  block_device {
    disk_id   = ovirt_disk.disk.id
    interface = "virtio"
  }
}
`
}

func testAccDiskBasic() string {
	return testAccDiskBasicDef() + `
resource "ovirt_disk" "disk" {
  name        	    = "testAccDiskBasic"
  alias             = "testAccDiskBasic"
  size              = 2
  format            = "cow"
  storage_domain_id = local.storage_domain_id
  sparse            = true
}
`
}

func testAccDiskBasicUpdate() string {
	return testAccDiskBasicDef() + `
resource "ovirt_disk" "disk" {
  name        	    = "testAccDiskBasicUpdate"
  alias             = "testAccDiskBasicUpdate"
  size              = 3
  format            = "cow"
  storage_domain_id = local.storage_domain_id
  sparse            = true
}
`
}
