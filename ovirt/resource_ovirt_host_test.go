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
func DisableTestAccOvirtHost_basic(t *testing.T) {
	var host ovirtsdk4.Host
	clusterID, updateClusterID := "ffeb3172-342e-11e9-8787-0cc47a7c8ea6", "ffeb3172-342e-11e9-8787-0cc47a7c8ea6"
	address, updateAddress := "10.10.0.171", "10.10.0.171"
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_host.host",
		CheckDestroy:  testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostBasic(address, clusterID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtHostExists("ovirt_host.host", &host),
					resource.TestCheckResourceAttr("ovirt_host.host", "name", "jnode1"),
					resource.TestCheckResourceAttr("ovirt_host.host", "address", address),
					resource.TestCheckResourceAttr("ovirt_host.host", "description", "my new host"),
				),
			},
			{
				Config: testAccHostBasicUpdate(updateAddress, updateClusterID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtHostExists("ovirt_host.host", &host),
					resource.TestCheckResourceAttr("ovirt_host.host", "name", "jnode1"),
					resource.TestCheckResourceAttr("ovirt_host.host", "address", updateAddress),
					resource.TestCheckResourceAttr("ovirt_host.host", "description", "my updated new host"),
				),
			},
		},
	})
}

func testAccCheckHostDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_host" {
			continue
		}
		getResp, err := conn.SystemService().HostsService().
			HostService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Host(); ok {
			return fmt.Errorf("Host %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtHostExists(n string, v *ovirtsdk4.Host) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Host ID is set")
		}
		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
		getResp, err := conn.SystemService().HostsService().
			HostService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		host, ok := getResp.Host()
		if ok {
			*v = *host
			return nil
		}
		return fmt.Errorf("Host %s not exist", rs.Primary.ID)
	}
}

func testAccHostBasic(address, clusterID string) string {
	return fmt.Sprintf(`
resource "ovirt_host" "host" {
  name          = "jnode1"
  description   = "my new host"
  address       = "%s"
  root_password = "secret"
  cluster_id    = "%s"
}
`, address, clusterID)
}

func testAccHostBasicUpdate(address, clusterID string) string {
	return fmt.Sprintf(`
resource "ovirt_host" "host" {
  name          = "jnode1"
  description   = "my updated new host"
  address       = "%s"
  root_password = "secret"
  cluster_id    = "%s"
}
`, address, clusterID)
}
