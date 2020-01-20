// Copyright (C) 2019 oVirt Maintainers
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

func TestAccOvirtImageTransfer_basic(t *testing.T) {
	var imageTransfer ovirtsdk4.ImageTransfer
	alias := "cirros-disk"
	sourceUrl := "http://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
	sdId := "d787bf6b-fae1-4a3e-b773-2ac466599d29"
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_image_transfer.transfer.id",
		CheckDestroy:  testAccCheckImageTransferDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageTransferBasic(alias, sourceUrl, sdId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtImageTransferExists("ovirt_image_transfer", &imageTransfer),
					resource.TestCheckResourceAttr("ovirt_image_transfer.transfer", "alias", alias),
				),
			},
		},
	})
}

func testAccCheckImageTransferDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_image_transfer" {
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
			return fmt.Errorf("Image transger %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtImageTransferExists(n string, imageTransfer *ovirtsdk4.ImageTransfer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Image-transfer ID is set")
		}

		_, err := parseResourceID(rs.Primary.ID, 2)
		if err != nil {
			return err
		}
		//alias, sourceUrl, sdId := parts[0], parts[1], parts[2]

		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().ImageTransfersService().
			ImageTransferService(imageTransfer.MustId()).
			Get().
			Send()
		if err != nil {
			return err
		}
		if v, ok := getResp.ImageTransfer(); ok {
			*imageTransfer = *v
			return nil
		}
		return fmt.Errorf("Image Transfer %s not exist", rs.Primary.ID)
	}
}

func testAccImageTransferBasic(alias, sourceUrl, sdId string) string {
	return fmt.Sprintf(`
resource "ovirt_image_transfer" "transfer" {
  alias             = "%s"
  source_url        = "%s"
  storage_domain_id = "%s"
}
`, alias, sourceUrl, sdId)
}
