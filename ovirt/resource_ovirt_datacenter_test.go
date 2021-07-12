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

func TestAccOvirtDataCenter_basic(t *testing.T) {
	var dc ovirtsdk4.DataCenter
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_datacenter.datacenter",
		CheckDestroy:  testAccCheckDataCenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCenterBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataCenterExists("ovirt_datacenter.datacenter", &dc),
					resource.TestCheckResourceAttr("ovirt_datacenter.datacenter", "name", "testAccOvirtDataCenterBasic"),
					resource.TestCheckResourceAttr("ovirt_datacenter.datacenter", "local", "false"),
					resource.TestCheckResourceAttrSet("ovirt_datacenter.datacenter", "status"),
				),
			},
			{
				Config: testAccDataCenterBasicUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtDataCenterExists("ovirt_datacenter.datacenter", &dc),
					resource.TestCheckResourceAttr("ovirt_datacenter.datacenter", "name", "testAccOvirtDataCenterBasicUpdate"),
					resource.TestCheckResourceAttr("ovirt_datacenter.datacenter", "local", "true"),
				),
			},
		},
	})
}

func testAccCheckDataCenterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_datacenter" {
			continue
		}
		getResp, err := conn.SystemService().DataCentersService().
			DataCenterService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.DataCenter(); ok {
			return fmt.Errorf("DataCenter %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtDataCenterExists(n string, v *ovirtsdk4.DataCenter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No DataCenter ID is set")
		}
		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
		getResp, err := conn.SystemService().DataCentersService().
			DataCenterService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		dc, ok := getResp.DataCenter()
		if ok {
			*v = *dc
			return nil
		}
		return fmt.Errorf("DataCenter %s not exist", rs.Primary.ID)
	}
}

const testAccDataCenterBasic = `
resource "ovirt_datacenter" "datacenter" {
  name        = "testAccOvirtDataCenterBasic"
  description = "my new dc"
  local       = false
}
`

const testAccDataCenterBasicUpdate = `
resource "ovirt_datacenter" "datacenter" {
  name        = "testAccOvirtDataCenterBasicUpdate"
  description = "my updated new dc"
  local       = true
}
`
