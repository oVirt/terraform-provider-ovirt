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

func TestAccOvirtMacPool_basic(t *testing.T) {
	var macpool ovirtsdk4.MacPool
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_mac_pool.pool",
		CheckDestroy:  testAccCheckMacPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMacPoolBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtMacPoolExists("ovirt_mac_pool.pool", &macpool),
					resource.TestCheckResourceAttr("ovirt_mac_pool.pool", "name", "testAccOvirtMacPoolBasic"),
					resource.TestCheckResourceAttr("ovirt_mac_pool.pool", "ranges.#", "2"),
				),
			},
			{
				Config: testAccMacPoolBasicUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtMacPoolExists("ovirt_mac_pool.pool", &macpool),
					resource.TestCheckResourceAttr("ovirt_mac_pool.pool", "name", "testAccOvirtMacPoolBasicUpdate"),
					resource.TestCheckResourceAttr("ovirt_mac_pool.pool", "ranges.#", "3"),
				),
			},
		},
	})
}

func testAccCheckMacPoolDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_mac_pool" {
			continue
		}
		getResp, err := conn.SystemService().MacPoolsService().
			MacPoolService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Pool(); ok {
			return fmt.Errorf("MacPool %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtMacPoolExists(n string, v *ovirtsdk4.MacPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No MacPool ID is set")
		}
		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
		getResp, err := conn.SystemService().MacPoolsService().
			MacPoolService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		macpool, ok := getResp.Pool()
		if ok {
			*v = *macpool
			return nil
		}
		return fmt.Errorf("MacPool %s not exist", rs.Primary.ID)
	}
}

func testAccMacPoolBasic() string {
	return fmt.Sprintf(`
resource "ovirt_mac_pool" "pool" {
  name             = "testAccOvirtMacPoolBasic"
  description      = "Desc of mac pool"
  allow_duplicates = true
	
  ranges = [
    "00:1a:4a:16:01:51,00:1a:4a:16:01:61",
    "00:1a:4a:16:01:71,00:1a:4a:16:01:81",
  ]
}
`)
}

func testAccMacPoolBasicUpdate() string {
	return fmt.Sprintf(`
resource "ovirt_mac_pool" "pool" {
  name             = "testAccOvirtMacPoolBasicUpdate"
  description      = "Desc of mac pool"
  allow_duplicates = true
	
  ranges = [
    "00:1a:4a:16:01:51,00:1a:4a:16:01:61",
    "00:1a:4a:16:01:91,00:1a:4a:16:01:a1",
    "00:1a:4a:16:01:b1,00:1a:4a:16:01:c1",
  ]
}
`)
}
