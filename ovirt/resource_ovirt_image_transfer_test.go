// Copyright (C) 2019 oVirt Maintainers
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
	govirt "github.com/oVirt/go-ovirt-client"
	ovirtsdk "github.com/ovirt/go-ovirt"
)

// TODO fix this test
func DisableTestAccOvirtImageTransfer_basic(t *testing.T) {
	var imageTransfer ovirtsdk.ImageTransfer
	suite := getOvirtTestSuite(t)
	id := suite.GenerateRandomID(5)
	alias := fmt.Sprintf("tf_test_%s", id)
	sourceUrl := suite.TestImageSourceURL()
	sdId := suite.StorageDomainID()

	resource.Test(t, resource.TestCase{
		PreCheck:      suite.PreCheck,
		Providers:     suite.Providers(),
		IDRefreshName: "ovirt_image_transfer.transfer.id",
		CheckDestroy:  testAccCheckImageTransferDestroy(suite),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "ovirt_image_transfer" "transfer" {
  alias             = "%s"
  source_url        = "%s"
  storage_domain_id = "%s"
  sparse            = true
}
`, alias, sourceUrl, sdId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtImageTransferExists("ovirt_image_transfer", &imageTransfer, suite),
					resource.TestCheckResourceAttr("ovirt_image_transfer.transfer", "alias", alias),
				),
			},
		},
	})
}

func testAccCheckImageTransferDestroy(suite OvirtTestSuite) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		conn := suite.Client().(govirt.ClientWithLegacySupport).GetSDKClient()
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
				if _, ok := err.(*ovirtsdk.NotFoundError); ok {
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
}

func testAccCheckOvirtImageTransferExists(
	n string,
	imageTransfer *ovirtsdk.ImageTransfer,
	suite OvirtTestSuite,
) resource.TestCheckFunc {
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

		conn := suite.Client().(govirt.ClientWithLegacySupport).GetSDKClient()
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
