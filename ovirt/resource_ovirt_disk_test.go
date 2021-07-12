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

// TODO fix this test
func DisableTestAccOvirtDisk_basic(t *testing.T) {
	var disk ovirtsdk4.Disk
	storageDomainID := "3be288f3-a43a-41fc-9d7d-0e9606dd67f3"
	quotaID := "1ab0cac2-8200-4e52-9c2d-e636911a7e9b"
	clusterID := "5b90f237-033c-004f-0234-000000000331"
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_disk.disk",
		CheckDestroy:  testAccCheckDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskBasic(clusterID, quotaID, storageDomainID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskExists("ovirt_disk.disk", &disk),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "name", "testAccDiskBasic"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "alias", "testAccDiskBasic"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "size", "2"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "format", "cow"),
				),
			},
			{
				Config: testAccDiskBasicUpdate(clusterID, quotaID, storageDomainID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskExists("ovirt_disk.disk", &disk),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "name", "testAccDiskBasicUpdate"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "alias", "testAccDiskBasicUpdate"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "size", "3"),
					resource.TestCheckResourceAttr("ovirt_disk.disk", "format", "cow"),
				),
			},
		},
	})
}

func testAccCheckDiskDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
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
		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
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

func testAccDiskBasic(clusterID, quotaID, storageDomainID string) string {
	return fmt.Sprintf(`

resource "ovirt_vm" "vm" {
  name        = "testAccVM"
  cluster_id  = "%s"
  memory 	  = 1024 

  block_device {
    disk_id   = "${ovirt_disk.disk.id}"
    interface = "virtio"
  }
}

resource "ovirt_disk" "disk" {
  name        	    = "testAccDiskBasic"
  alias             = "testAccDiskBasic"
  size              = 2
  format            = "cow"
  quota_id          = "%s"
  storage_domain_id = "%s"
  sparse            = true
}
`, clusterID, quotaID, storageDomainID)
}

func testAccDiskBasicUpdate(clusterID, quotaID, storageDomainID string) string {
	return fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name       = "testAccVM"
  cluster_id = "%s"
  memory     = 1024

  block_device {
    disk_id   = "${ovirt_disk.disk.id}"
    interface = "virtio"
  }
}

resource "ovirt_disk" "disk" {
  name        	    = "testAccDiskBasicUpdate"
  alias             = "testAccDiskBasicUpdate"
  size              = 3
  format            = "cow"
  quota_id          = "%s"
  storage_domain_id = "%s"
  sparse            = true
}
`, clusterID, quotaID, storageDomainID)
}
