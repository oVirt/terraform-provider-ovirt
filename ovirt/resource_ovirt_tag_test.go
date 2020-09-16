// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func TestAccOvirtTag_basic(t *testing.T) {
	var tag ovirtsdk4.Tag
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_tag.tag",
		CheckDestroy:  testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtTagExists("ovirt_tag.tag", &tag),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "name", "testAccOvirtTagBasic"),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "description", "my new tag"),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "vm_ids.#", "2"),
					testAccCheckOvirtTagAttachedEntities(&tag, "vm_ids", []string{
						"testAccTagBasicVM1",
						"testAccTagBasicVM2",
					}),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "host_ids.#", "1"),
					testAccCheckOvirtTagAttachedEntities(&tag, "host_ids", []string{
						"host65",
					}),
				),
			},
			{
				Config: testAccTagBasicUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtTagExists("ovirt_tag.tag", &tag),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "name", "testAccOvirtTagBasicUpdate"),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "description", "my updated new tag"),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "vm_ids.#", "2"),
					testAccCheckOvirtTagAttachedEntities(&tag, "vm_ids", []string{
						"testAccTagBasicVM1",
						"testAccTagBasicVM3",
					}),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "host_ids.#", "0"),
				),
			},
		},
	})
}

func testAccCheckTagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_tag" {
			continue
		}
		getResp, err := conn.SystemService().TagsService().
			TagService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Tag(); ok {
			return fmt.Errorf("Tag %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtTagExists(n string, v *ovirtsdk4.Tag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Tag ID is set")
		}
		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().TagsService().
			TagService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		tag, ok := getResp.Tag()
		if ok {
			*v = *tag
			return nil
		}
		return fmt.Errorf("Tag %s not exist", rs.Primary.ID)
	}
}

func getVmNamesFromTag(service *ovirtsdk4.VmsService, tagName string) ([]string, error) {
	var vmNames []string
	resp, err := service.List().Search(fmt.Sprintf("tag=%s", tagName)).Send()
	if err != nil {
		return nil, err
	}
	for _, v := range resp.MustVms().Slice() {
		vmNames = append(vmNames, v.MustName())
	}
	return vmNames, nil
}

func getHostNamesFromTag(service *ovirtsdk4.HostsService, tagName string) ([]string, error) {
	var hostNames []string
	resp, err := service.List().Search(fmt.Sprintf("tag=%s", tagName)).Send()
	if err != nil {
		return nil, err
	}
	for _, v := range resp.MustHosts().Slice() {
		hostNames = append(hostNames, v.MustName())
	}
	return hostNames, nil
}

func testAccCheckOvirtTagAttachedEntities(v *ovirtsdk4.Tag, field string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		systemService := testAccProvider.Meta().(*ovirtsdk4.Connection).SystemService()
		var names []string
		var err error
		switch field {
		case "vm_ids":
			names, err = getVmNamesFromTag(systemService.VmsService(), v.MustName())
		case "host_ids":
			names, err = getHostNamesFromTag(systemService.HostsService(), v.MustName())
		default:
			return fmt.Errorf("Unsupported Tag attached to Entity %s", field)
		}
		if err != nil {
			return err
		}
		// Compare after sorted
		sort.Strings(names)
		sort.Strings(expected)
		if !reflect.DeepEqual(names, expected) {
			return fmt.Errorf("Attribute '%s' expected %#v, got %#v", field, expected, names)
		}
		return nil
	}
}

const testAccTagBasicDef = `
data "ovirt_clusters" "c" {
  search = {
    criteria = "name = Default"
  }
}

data "ovirt_hosts" "h" {
  search = {
    criteria = "name = host65" 
  }
}

data "ovirt_templates" "t" {
  search = {
    criteria = "name = testTemplate"
  }
}

locals {
  cluster_id        = data.ovirt_clusters.c.clusters.0.id
  host_id           = data.ovirt_hosts.h.hosts.0.id
  template_id       = data.ovirt_templates.t.templates.0.id
}

resource "ovirt_vm" "vm1" {
  name              = "testAccTagBasicVM1"
  cluster_id        = local.cluster_id
  template_id       = local.template_id
  os {
    type = "other"
  }
}

resource "ovirt_vm" "vm2" {
  name              = "testAccTagBasicVM2"
  cluster_id        = local.cluster_id
  template_id       = local.template_id
  os {
    type = "other"
  }
}

resource "ovirt_vm" "vm3" {
  name              = "testAccTagBasicVM3"
  cluster_id        = local.cluster_id
  template_id       = local.template_id
  os {
    type = "other"
  }
}

`

const testAccTagBasic = testAccTagBasicDef + `
resource "ovirt_tag" "tag" {
  name        = "testAccOvirtTagBasic"
  parent_id   = "00000000-0000-0000-0000-000000000000"
  description = "my new tag"
	
  vm_ids = [
    ovirt_vm.vm1.id,
    ovirt_vm.vm2.id,
  ]

  host_ids = [
    local.host_id,
  ]
}
`

const testAccTagBasicUpdate = testAccTagBasicDef + `
resource "ovirt_tag" "tag" {
  name        = "testAccOvirtTagBasicUpdate"
  parent_id   = "00000000-0000-0000-0000-000000000000"
  description = "my updated new tag"

  vm_ids = [
    ovirt_vm.vm1.id,
    ovirt_vm.vm3.id,
  ]

}
`
