// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt_test

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	govirt "github.com/ovirt/go-ovirt-client"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// TODO fix this test
func DisableTestAccOvirtTag_basic(t *testing.T) {
	var tag ovirtsdk4.Tag
	resource.Test(t, resource.TestCase{
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
					resource.TestCheckResourceAttr("ovirt_tag.tag", "vm_ids.#", "3"),
					testAccCheckOvirtTagAttachedEntities(&tag, "vm_ids", []string{
						"9c993532-9f70-4c56-88a2-b40d6b48283a",
						"900bc22d-c776-4c87-93a6-41bb36eb4d8b",
						"dcb76ed3-f7e6-4c53-a0be-87bde821e431",
					}),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "host_ids.#", "1"),
					testAccCheckOvirtTagAttachedEntities(&tag, "host_ids", []string{
						"fa0e3d1b-f3a7-49d7-8e72-045e562f81a6",
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
						"9c993532-9f70-4c56-88a2-b40d6b48283a",
						"423ebce9-30e8-8894-8216-a6f0ab803c4c",
					}),
					resource.TestCheckResourceAttr("ovirt_tag.tag", "host_ids.#", "1"),
					testAccCheckOvirtTagAttachedEntities(&tag, "host_ids", []string{
						"269d7afe-6e70-4712-b179-0cd8821d7d30",
					}),
				),
			},
		},
	})
}

func testAccCheckTagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
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
		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
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

func testAccCheckOvirtTagAttachedEntities(v *ovirtsdk4.Tag, field string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		systemService := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient().SystemService()
		var ids []string
		var err error
		switch field {
		case "vm_ids":
			ids, err = searchVmsByTag(systemService.VmsService(), v.MustName())
		case "host_ids":
			ids, err = searchHostsByTag(systemService.HostsService(), v.MustName())
		default:
			return fmt.Errorf("Unsupported Tag attached to Entity %s", field)
		}
		if err != nil {
			return err
		}
		// Compare after sorted
		sort.Strings(ids)
		sort.Strings(expected)
		if !reflect.DeepEqual(ids, expected) {
			return fmt.Errorf("Attribute '%s' expected %#v, got %#v", field, expected, ids)
		}
		return nil
	}
}

const testAccTagBasic = `
resource "ovirt_tag" "tag" {
  name        = "testAccOvirtTagBasic"
  parent_id   = "00000000-0000-0000-0000-000000000000"
  description = "my new tag"
	
  vm_ids = [
    "9c993532-9f70-4c56-88a2-b40d6b48283a",
    "900bc22d-c776-4c87-93a6-41bb36eb4d8b",
    "dcb76ed3-f7e6-4c53-a0be-87bde821e431",
  ]

  host_ids = [
    "fa0e3d1b-f3a7-49d7-8e72-045e562f81a6",
  ]
}
`

const testAccTagBasicUpdate = `
resource "ovirt_tag" "tag" {
  name        = "testAccOvirtTagBasicUpdate"
  parent_id   = "00000000-0000-0000-0000-000000000000"
  description = "my updated new tag"

  vm_ids = [
    "9c993532-9f70-4c56-88a2-b40d6b48283a",
    "423ebce9-30e8-8894-8216-a6f0ab803c4c",
  ]

  host_ids = [
    "269d7afe-6e70-4712-b179-0cd8821d7d30",
  ]
}
`
