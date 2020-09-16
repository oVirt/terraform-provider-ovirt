// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func TestAccOvirtNetwork_basic(t *testing.T) {
	vlanID, vlanIDUpdate := 2, 3
	desc, descUpdate := "desc-1", "desc-1-update"
	mtu, mtuUpdate := 1500, 2000
	var network ovirtsdk4.Network
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckNetworkDestroy,
		IDRefreshName: "ovirt_network.network",
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkBasic(desc, vlanID, mtu),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtNetworkExists("ovirt_network.network", &network),
					resource.TestCheckResourceAttr("ovirt_network.network", "name", "testAccOvirtNetworkBasic"),
					resource.TestCheckResourceAttr("ovirt_network.network", "description", desc),
					resource.TestCheckResourceAttr("ovirt_network.network", "vlan_id", strconv.Itoa(vlanID)),
					resource.TestCheckResourceAttr("ovirt_network.network", "mtu", strconv.Itoa(mtu)),
				),
			},
			{
				Config: testAccNetworkBasic(descUpdate, vlanIDUpdate, mtuUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtNetworkExists("ovirt_network.network", &network),
					resource.TestCheckResourceAttr("ovirt_network.network", "name", "testAccOvirtNetworkBasic"),
					resource.TestCheckResourceAttr("ovirt_network.network", "description", descUpdate),
					resource.TestCheckResourceAttr("ovirt_network.network", "vlan_id", strconv.Itoa(vlanIDUpdate)),
					resource.TestCheckResourceAttr("ovirt_network.network", "mtu", strconv.Itoa(mtuUpdate)),
				),
			},
		},
	})
}

func testAccCheckNetworkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
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

		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
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

func testAccNetworkBasic(desc string, vlanID, mtu int) string {
	return fmt.Sprintf(`
data "ovirt_datacenters" "search_filtered_datacenter" {
  search = {
    criteria       = "name = Default"
    max            = 2
    case_sensitive = false
  }
}

locals {
  datacenter_id = data.ovirt_datacenters.search_filtered_datacenter.datacenters.0.id
}

resource "ovirt_network" "network" {
  name          = "testAccOvirtNetworkBasic"
  datacenter_id = local.datacenter_id
  description   = "%s"
  vlan_id       = %d
  mtu           = %d
}
`, desc, vlanID, mtu)
}
