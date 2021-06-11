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
	"github.com/janoszen/govirt"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func TestAccOvirtVnicProfile_basic(t *testing.T) {
	var profile ovirtsdk4.VnicProfile
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVnicProfileDestroy,
		IDRefreshName: "ovirt_vnic_profile.profile",
		Steps: []resource.TestStep{
			{
				Config: testAccVnicProfileBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVnicProfileExists("ovirt_vnic_profile.profile", &profile),
					resource.TestCheckResourceAttr("ovirt_vnic_profile.profile", "name", "testAccOvirtVnicProfileBasic"),
					resource.TestCheckResourceAttr("ovirt_vnic_profile.profile", "migratable", "false"),
					resource.TestCheckResourceAttr("ovirt_vnic_profile.profile", "port_mirroring", "false"),
				),
			},
			{
				Config: testAccVnicProfileBasicUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVnicProfileExists("ovirt_vnic_profile.profile", &profile),
					resource.TestCheckResourceAttr("ovirt_vnic_profile.profile", "name", "testAccOvirtVnicProfileBasicUpdate"),
					resource.TestCheckResourceAttr("ovirt_vnic_profile.profile", "migratable", "true"),
					resource.TestCheckResourceAttr("ovirt_vnic_profile.profile", "port_mirroring", "true"),
				),
			},
		},
	})
}

func testAccCheckVnicProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.Client).GetSDKClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_vnic_profile" {
			continue
		}
		getResp, err := conn.SystemService().VnicProfilesService().
			ProfileService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Profile(); ok {
			return fmt.Errorf("VnicProfile %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtVnicProfileExists(n string, v *ovirtsdk4.VnicProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No VnicProfile ID is set")
		}
		conn := testAccProvider.Meta().(govirt.Client).GetSDKClient()
		getResp, err := conn.SystemService().VnicProfilesService().
			ProfileService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		profile, ok := getResp.Profile()
		if ok {
			*v = *profile
			return nil
		}
		return fmt.Errorf("VnicProfile %s not exist", rs.Primary.ID)
	}
}

const testAccVnicProfileBasic = `
data "ovirt_networks" "ovirtmgmt-test" {
  search {
    criteria       = "datacenter = Default and name = ovirtmgmt-test"
    max            = 1
    case_sensitive = false
  }
}

resource "ovirt_vnic_profile" "profile" {
  name        	 = "testAccOvirtVnicProfileBasic"
  network_id  	 = "${data.ovirt_networks.ovirtmgmt-test.networks.0.id}"
  migratable  	 = false
  port_mirroring = false
}
`

const testAccVnicProfileBasicUpdate = `
data "ovirt_networks" "ovirtmgmt-test" {
  search {
    criteria       = "datacenter = Default and name = ovirtmgmt-test"
    max            = 1
    case_sensitive = false
  }
}

resource "ovirt_vnic_profile" "profile" {
  name        	 = "testAccOvirtVnicProfileBasicUpdate"
  network_id  	 = "${data.ovirt_networks.ovirtmgmt-test.networks.0.id}"
  migratable  	 = true
  port_mirroring = true
}
`
