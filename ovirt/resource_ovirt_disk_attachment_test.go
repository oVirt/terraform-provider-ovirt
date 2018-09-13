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

func TestAccOvirtDiskAttachment_basic(t *testing.T) {
	var diskAttachment ovirtsdk4.DiskAttachment
	vmID := "d22d9233-8c9f-42f0-a137-ccd4af45dec7"
	diskID := "8d54d1f5-4549-441f-8801-e2660c77a5c8"
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_disk_attachment.attachment",
		CheckDestroy:  testAccCheckDiskAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskAttachmentBasic(vmID, diskID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskAttachmentExists("ovirt_disk_attachment.attachment", &diskAttachment),
					resource.TestCheckResourceAttr("ovirt_disk_attachment.attachment", "interface", "virtio"),
					resource.TestCheckResourceAttr("ovirt_disk_attachment.attachment", "bootable", "true"),
					resource.TestCheckResourceAttr("ovirt_disk_attachment.attachment", "read_only", "true"),
					resource.TestCheckResourceAttr("ovirt_disk_attachment.attachment", "active", "true"),
				),
			},
			{
				Config: testAccDiskAttachmentBasicUpdate(vmID, diskID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDiskAttachmentExists("ovirt_disk_attachment.attachment", &diskAttachment),
					resource.TestCheckResourceAttr("ovirt_disk_attachment.attachment", "interface", "virtio"),
					resource.TestCheckResourceAttr("ovirt_disk_attachment.attachment", "bootable", "false"),
					resource.TestCheckResourceAttr("ovirt_disk_attachment.attachment", "active", "false"),
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
			return fmt.Errorf("Disk attachment %s still exist", rs.Primary.ID)
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

		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().VmsService().
			VmService(rs.Primary.Attributes["vm_id"]).
			DiskAttachmentsService().
			AttachmentService(rs.Primary.Attributes["disk_id"]).
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

func testAccDiskAttachmentBasic(vmID, diskID string) string {
	return fmt.Sprintf(`
resource "ovirt_disk_attachment" "attachment" {
	vm_id     = "%s"
	disk_id   = "%s"
	bootable  = true
	interface = "virtio"
	active    = true
	read_only = true
}  
`, vmID, diskID)
}

func testAccDiskAttachmentBasicUpdate(vmID, diskID string) string {
	return fmt.Sprintf(`
resource "ovirt_disk_attachment" "attachment" {
	vm_id     = "%s"
	disk_id   = "%s"
	bootable  = false
	interface = "virtio"
	active    = false
	read_only = true
}  
`, vmID, diskID)
}
