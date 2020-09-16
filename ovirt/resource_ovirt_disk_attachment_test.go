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

func TestAccOvirtDiskAttachment_basic(t *testing.T) {
	var diskAttachment ovirtsdk4.DiskAttachment
	vmID := "detached_vm"
	const resourceName = "ovirt_disk_attachment.attachment"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: resourceName,
		CheckDestroy:  testAccCheckDiskAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskAttachmentBasic(vmID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskAttachmentExists(resourceName, &diskAttachment),
					resource.TestCheckResourceAttr(resourceName, "interface", "virtio"),
					resource.TestCheckResourceAttr(resourceName, "bootable", "true"),
					resource.TestCheckResourceAttr(resourceName, "read_only", "true"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
				),
			},
			{
				Config: testAccDiskAttachmentBasicUpdate(vmID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskAttachmentExists(resourceName, &diskAttachment),
					resource.TestCheckResourceAttr(resourceName, "interface", "virtio"),
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
				),
			},
		},
	})
}

func testAccCheckDiskAttachmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_disk_attachment" {
			continue
		}

		parts, err := parseResourceID(rs.Primary.ID, 2)
		if err != nil {
			return err
		}
		vmID, diskID := parts[0], parts[1]

		getResp, err := conn.SystemService().VmsService().
			VmService(vmID).
			DiskAttachmentsService().
			AttachmentService(diskID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Attachment(); ok {
			return fmt.Errorf("Disk attachment %s still exist", rs.Primary.ID)
		}

		// Delete disk created
		diskResp, err := conn.SystemService().DisksService().
			DiskService(diskID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if disk, ok := diskResp.Disk(); ok {
			return fmt.Errorf("Disk  %s still exist", disk.MustId())
		}
	}
	return nil
}

func testAccCheckOvirtDiskAttachmentExists(n string, diskAttachment *ovirtsdk4.DiskAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Disk attachment ID is set")
		}

		parts, err := parseResourceID(rs.Primary.ID, 2)
		if err != nil {
			return err
		}
		vmID, diskID := parts[0], parts[1]

		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().VmsService().
			VmService(vmID).
			DiskAttachmentsService().
			AttachmentService(diskID).
			Get().
			Send()
		if err != nil {
			return err
		}
		if v, ok := getResp.Attachment(); ok {
			*diskAttachment = *v
			return nil
		}
		return fmt.Errorf("Disk attachment %s not exist", rs.Primary.ID)
	}
}

func testAccDiskAttachmentDef(vmName string) string {
	return fmt.Sprintf(`
data "ovirt_vms" "search_filtered_vm" {
  search = {
    criteria       = "name = %s"
  }
}

data "ovirt_storagedomains" "sd" {
  search = {
    criteria       = "name = data"
    case_sensitive = false
  }
}

resource "ovirt_disk" "disk" {
  name        	    = "testAccDiskBasic"
  alias             = "testAccDiskBasic"
  size              = 10
  format            = "cow"
  storage_domain_id = data.ovirt_storagedomains.sd.storagedomains.0.id
  sparse            = true
}

locals {
  vm_id = data.ovirt_vms.search_filtered_vm.vms.0.id
  disk_id = ovirt_disk.disk.id
}
`, vmName)
}

func testAccDiskAttachmentBasic(vmName string) string {
	return testAccDiskAttachmentDef(vmName) + `
resource "ovirt_disk_attachment" "attachment" {
  vm_id     = local.vm_id
  disk_id   = local.disk_id
  bootable  = true
  interface = "virtio"
  active    = true
  read_only = true
}  
`
}

func testAccDiskAttachmentBasicUpdate(vmName string) string {
	return testAccDiskAttachmentDef(vmName) + `
resource "ovirt_disk_attachment" "attachment" {
  vm_id     = local.vm_id
  disk_id   = local.disk_id
  bootable  = true
  interface = "virtio"
  active    = false
  read_only = true
}  
`
}
