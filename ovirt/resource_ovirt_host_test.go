// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
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

func TestAccOvirtHost_basic(t *testing.T) {
	var host ovirtsdk4.Host
	address, updateAddress := "10.10.0.171", "10.10.0.171"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_host.host",
		CheckDestroy:  testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostBasic(address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtHostExists("ovirt_host.host", &host),
					resource.TestCheckResourceAttr("ovirt_host.host", "name", "jnode1"),
					resource.TestCheckResourceAttr("ovirt_host.host", "address", address),
					resource.TestCheckResourceAttr("ovirt_host.host", "description", "my new host"),
				),
			},
			{
				Config: testAccHostBasicUpdate(updateAddress),
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
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
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
		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
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

func testAccHostBasicDef() string {
	return `
data "ovirt_clusters" "c" {
  search = {
    criteria = "name = Default2"
  }
}

locals {
  cluster_id = data.ovirt_clusters.c.clusters.0.id
}
`
}

func testAccHostBasic(address string) string {
	return testAccHostBasicDef() + fmt.Sprintf(`
resource "ovirt_host" "host" {
  name          = "jnode1"
  description   = "my new host"
  address       = "%s"
  root_password = "secret"
  cluster_id    = local.cluster_id
}
`, address)
}

func testAccHostBasicUpdate(address string) string {
	return testAccHostBasicDef() + fmt.Sprintf(`
resource "ovirt_host" "host" {
  name          = "jnode1"
  description   = "my updated new host"
  address       = "%s"
  root_password = "secret"
  cluster_id    = local.cluster_id
}
`, address)
}
