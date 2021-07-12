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
	govirt "github.com/ovirt/go-ovirt-client"
	ovirtsdk "github.com/ovirt/go-ovirt"
)

func TestAccOvirtImageTransfer_basic(t *testing.T) {

	suite := getOvirtTestSuite(t)
	id := suite.GenerateRandomID(5)
	alias := fmt.Sprintf("tf_test_%s", id)
	sourceUrl := suite.TestImageSourceURL()
	sdId := suite.StorageDomainID()

	resource.Test(t, resource.TestCase{
		PreCheck:      suite.PreCheck,
		Providers:     suite.Providers(),
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
					testAccCheckOvirtImageTransferExists("ovirt_image_transfer.transfer", suite),
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

			_, err := conn.SystemService().
				DisksService().
				DiskService(rs.Primary.ID).
				Get().
				Send()
			if err != nil {
				if _, ok := err.(*ovirtsdk.NotFoundError); ok {
					continue
				}
				return err
			}
			return fmt.Errorf("Image transger %s still exist", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckOvirtImageTransferExists(
	n string,
	suite OvirtTestSuite,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no disk ID is set")
		}

		conn := suite.Client().(govirt.ClientWithLegacySupport).GetSDKClient()
		_, err := conn.SystemService().DisksService().
			DiskService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return fmt.Errorf("image transfer %s not exist (%w)", rs.Primary.ID, err)
		}
		return nil
	}
}
