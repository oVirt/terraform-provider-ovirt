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
func DisabledTestAccOvirtDiskAttachment_basic(t *testing.T) {
	var diskAttachment ovirtsdk4.DiskAttachment
	vmID := "437d0f69-d1eb-441f-bf6b-0e97797fe11e"
	diskID := "230349f6-59a9-47e9-bc90-7c1221645b07"
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
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
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

		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
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
