// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
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

func TestAccOvirtDisk_basic(t *testing.T) {
	var disk ovirtsdk4.Disk
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_disk.disk",
		CheckDestroy:  testAccCheckDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskExists("ovirt_disk.disk", &disk),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "name", "testAccDiskBasic"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "alias", "testAccDiskBasic"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "size", "23687091200"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "format", "cow"),
				),
			},
			{
				Config: testAccDiskBasicUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskExists("ovirt_disk.disk", &disk),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "name", "testAccDiskBasicUpdate"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "alias", "testAccDiskBasicUpdate"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "size", "33687091200"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "format", "raw"),
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

const testAccDiskBasic = `
resource "ovirt_disk" "disk" {
	name        	  = "testAccDiskBasic"
	alias             = "testAccDiskBasic"
	size              = 23687091200
	format            = "cow"
	storage_domain_id = "${data.ovirt_storagedomains.data.storagedomains.0.id}"
	sparse            = true
}

data "ovirt_clusters" "default" {
	search = {
		criteria       = "name = Default"
		max            = 1
		case_sensitive = false
	}
}

resource "ovirt_vm" "vm" {
	name        = "testAccVMBasic"
	cluster_id  = "${data.ovirt_clusters.default.clusters.0.id}"
	attached_disk {
		disk_id = "${ovirt_disk.disk.id}"
		bootable = true
		interface = "virtio"
	}
}

data "ovirt_storagedomains" "data" {
	name_regex = "^data"
	search = {
	  criteria       = "name = data and datacenter = ${data.ovirt_datacenters.default.datacenters.0.name}"
	  case_sensitive = false
	}
}

data "ovirt_datacenters" "default" {
	search = {
		criteria       = "name = Default"
		max            = 1
		case_sensitive = false
	}
}

`

const testAccDiskBasicUpdate = `
resource "ovirt_disk" "disk" {
	name        	  = "testAccDiskBasicUpdate"
	alias             = "testAccDiskBasicUpdate"
	size              = 33687091200
	format            = "raw"
	storage_domain_id = "${data.ovirt_storagedomains.data.storagedomains.0.id}"
	sparse            = true
}

data "ovirt_clusters" "default" {
	search = {
		criteria       = "name = Default"
		max            = 1
		case_sensitive = false
	}
}

resource "ovirt_vm" "vm" {
	name        = "testAccVMBasic"
	cluster_id  = "${data.ovirt_clusters.default.clusters.0.id}"
	attached_disk {
		disk_id = "${ovirt_disk.disk.id}"
		bootable = true
		interface = "virtio"
	}
}

data "ovirt_storagedomains" "data" {
	name_regex = "^data"
	search = {
	  criteria       = "name = data and datacenter = ${data.ovirt_datacenters.default.datacenters.0.name}"
	  case_sensitive = false
	}
}

data "ovirt_datacenters" "default" {
	search = {
		criteria       = "name = Default"
		max            = 1
		case_sensitive = false
	}
}
`
