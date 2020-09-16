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

func TestAccOvirtVnic_basic(t *testing.T) {
	var nic ovirtsdk4.Nic
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVnicDestroy,
		IDRefreshName: "ovirt_vnic.nic",
		Steps: []resource.TestStep{
			{
				Config: testAccVnicBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVnicExists("ovirt_vnic.nic", &nic),
					resource.TestCheckResourceAttr("ovirt_vnic.nic", "name", "testAccOvirtVnicBasic"),
				),
			},
			{
				Config: testAccVnicBasicUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVnicExists("ovirt_vnic.nic", &nic),
					resource.TestCheckResourceAttr("ovirt_vnic.nic", "name", "testAccOvirtVnicBasicUpdate"),
				),
			},
		},
	})
}

func testAccCheckVnicDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_vnic" {
			continue
		}
		parts, err := parseResourceID(rs.Primary.ID, 2)
		if err != nil {
			return err
		}
		vmID, nicID := parts[0], parts[1]

		getResp, err := conn.SystemService().VmsService().
			VmService(vmID).
			NicsService().
			NicService(nicID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Nic(); ok {
			return fmt.Errorf("Vnic %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtVnicExists(n string, v *ovirtsdk4.Nic) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Vnic ID is set")
		}

		parts, err := parseResourceID(rs.Primary.ID, 2)
		if err != nil {
			return err
		}
		vmID, nicID := parts[0], parts[1]

		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().VmsService().
			VmService(vmID).
			NicsService().
			NicService(nicID).
			Get().
			Send()
		if err != nil {
			return err
		}
		nic, ok := getResp.Nic()
		if ok {
			*v = *nic
			return nil
		}
		return fmt.Errorf("Vnic %s not exist", rs.Primary.ID)
	}
}

const testAccVnicDef = `
data "ovirt_clusters" "c" {
  search = {
    criteria = "name = Default"
  }
}

data "ovirt_networks" "n" {
  search = {
    criteria = "datacenter = Default and name = ovirtmgmt"
  }
}

data "ovirt_vnic_profiles" "v" {
  name_regex = "ovirtmgmt"
  network_id = data.ovirt_networks.n.networks.0.id
}

data "ovirt_templates" "t" {
  search = {
    criteria = "name = testTemplate"
  }
}

locals {
  cluster_id        = data.ovirt_clusters.c.clusters.0.id
  vnic_profile_id   = data.ovirt_vnic_profiles.v.vnic_profiles.0.id
  template_id       = data.ovirt_templates.t.templates.0.id
}

resource "ovirt_vm" "vm1" {
  name        = "testAccVnicVM1"
  cluster_id  = local.cluster_id
  template_id = local.template_id
  memory      = 2048
  os {
    type = "other"
  }
}

resource "ovirt_vm" "vm2" {
  name        = "testAccVnicVM2"
  cluster_id  = local.cluster_id
  template_id = local.template_id
  memory      = 2048
  os {
    type = "other"
  }
}
`

const testAccVnicBasic = testAccVnicDef + `
resource "ovirt_vnic" "nic" {
  name            = "testAccOvirtVnicBasic"
  vm_id           = ovirt_vm.vm1.id
  vnic_profile_id = local.vnic_profile_id
}
`

const testAccVnicBasicUpdate = testAccVnicDef + `
resource "ovirt_vnic" "nic" {
  name            = "testAccOvirtVnicBasicUpdate"
  vm_id           = ovirt_vm.vm2.id
  vnic_profile_id = local.vnic_profile_id
}
`
