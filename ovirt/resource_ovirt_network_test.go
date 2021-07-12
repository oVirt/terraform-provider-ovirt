// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	govirt "github.com/ovirt/go-ovirt-client"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// TODO fix this test
func DisableTestAccOvirtNetwork_basic(t *testing.T) {
	datacenterID := "5baef02d-033c-0252-0168-0000000001d3"
	vlanID, vlanIDUpdate := 2, 3
	desc, descUpdate := "desc-1", "desc-1-update"
	mtu, mtuUpdate := 1500, 2000
	var network ovirtsdk4.Network
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckNetworkDestroy,
		IDRefreshName: "ovirt_network.network",
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkBasic(datacenterID, desc, vlanID, mtu),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtNetworkExists("ovirt_network.network", &network),
					resource.TestCheckResourceAttr("ovirt_network.network", "name", "testAccOvirtNetworkBasic"),
					resource.TestCheckResourceAttr("ovirt_network.network", "datacenter_id", datacenterID),
					resource.TestCheckResourceAttr("ovirt_network.network", "description", desc),
					resource.TestCheckResourceAttr("ovirt_network.network", "vlan_id", strconv.Itoa(vlanID)),
					resource.TestCheckResourceAttr("ovirt_network.network", "mtu", strconv.Itoa(mtu)),
				),
			},
			{
				Config: testAccNetworkBasic(datacenterID, descUpdate, vlanIDUpdate, mtuUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtNetworkExists("ovirt_network.network", &network),
					resource.TestCheckResourceAttr("ovirt_network.network", "name", "testAccOvirtNetworkBasic"),
					resource.TestCheckResourceAttr("ovirt_network.network", "datacenter_id", datacenterID),
					resource.TestCheckResourceAttr("ovirt_network.network", "description", descUpdate),
					resource.TestCheckResourceAttr("ovirt_network.network", "vlan_id", strconv.Itoa(vlanIDUpdate)),
					resource.TestCheckResourceAttr("ovirt_network.network", "mtu", strconv.Itoa(mtuUpdate)),
				),
			},
		},
	})
}

func testAccCheckNetworkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_network" {
			continue
		}

		getResp, err := conn.SystemService().NetworksService().
			NetworkService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Network(); ok {
			return fmt.Errorf("Network %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtNetworkExists(n string, v *ovirtsdk4.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ID is set")
		}

		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
		getResp, err := conn.SystemService().NetworksService().
			NetworkService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		network, ok := getResp.Network()
		if ok {
			*v = *network
			return nil
		}
		return fmt.Errorf("Network %s not exist", rs.Primary.ID)
	}
}

func testAccNetworkBasic(datacenterID, desc string, vlanID, mtu int) string {
	return fmt.Sprintf(`
resource "ovirt_network" "network" {
  name          = "testAccOvirtNetworkBasic"
  datacenter_id = "%s"
  description   = "%s"
  vlan_id       = %d
  mtu           = %d
}
`, datacenterID, desc, vlanID, mtu)
}
