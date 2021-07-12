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
func DisableTestAccOvirtUser_basic(t *testing.T) {
	var user ovirtsdk4.User
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUserDestroy,
		IDRefreshName: "ovirt_user.user",
		Steps: []resource.TestStep{
			{
				Config: testAccUserBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtUserExists("ovirt_user.user", &user),
					resource.TestCheckResourceAttr("ovirt_user.user", "name", "user1"),
					resource.TestCheckResourceAttr("ovirt_user.user", "namespace", "*"),
					resource.TestCheckResourceAttr("ovirt_user.user", "authz_name", "example.com-authz"),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_user" {
			continue
		}

		getResp, err := conn.SystemService().UsersService().
			UserService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.User(); ok {
			return fmt.Errorf("User %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtUserExists(n string, v *ovirtsdk4.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No User ID is set")
		}

		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
		getResp, err := conn.SystemService().UsersService().
			UserService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		user, ok := getResp.User()
		if ok {
			*v = *user
			return nil
		}
		return fmt.Errorf("User %s not exist", rs.Primary.ID)
	}
}

func testAccUserBasic() string {
	return fmt.Sprintf(`
resource "ovirt_user" "user" {
  name       = "testAccOvirtUserBasic@internal"
  namespace  = "*"
  authz_name = "example.com-authz"
}
`)
}
