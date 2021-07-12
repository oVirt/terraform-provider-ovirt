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
func DisableTestAccOvirtStorageDomain_nfs(t *testing.T) {
	var sd ovirtsdk4.StorageDomain
	hostID, dcID := "e92e4a4b-2960-4b28-927b-17d8eb800b98", "5baef02d-033c-0252-0168-0000000001d3"
	nfsAddr, nfsPath := "10.1.110.18", "/data161"
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageDomainDestroy,
		IDRefreshName: "ovirt_storage_domain.dataNFS",
		Steps: []resource.TestStep{
			{
				Config: testAccStorageDomainNFS(hostID, dcID, nfsAddr, nfsPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageDomainExists("ovirt_storage_domain.dataNFS", &sd),
					resource.TestCheckResourceAttr("ovirt_storage_domain.dataNFS", "name", "testAccOvirtStorageDomainNFS"),
					resource.TestCheckResourceAttr("ovirt_storage_domain.dataNFS", "datacenter_id", dcID),
					resource.TestCheckResourceAttr("ovirt_storage_domain.dataNFS", "nfs.#", "1"),
					resource.TestCheckResourceAttr("ovirt_storage_domain.dataNFS", "nfs.0.address", nfsAddr),
					resource.TestCheckResourceAttr("ovirt_storage_domain.dataNFS", "nfs.0.path", nfsPath),
				),
			},
		},
	})
}

func testAccCheckStorageDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_storage_domain" {
			continue
		}
		getResp, err := conn.SystemService().StorageDomainsService().
			StorageDomainService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.StorageDomain(); ok {
			return fmt.Errorf("StorageDomain %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckStorageDomainExists(n string, v *ovirtsdk4.StorageDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No StorageDomain ID is set")
		}
		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
		getResp, err := conn.SystemService().StorageDomainsService().
			StorageDomainService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		sd, ok := getResp.StorageDomain()
		if ok {
			*v = *sd
			return nil
		}
		return fmt.Errorf("StorageDomain %s not exist", rs.Primary.ID)
	}
}

func testAccStorageDomainNFS(hostID, dcID, nfsAddr, nfsPath string) string {
	return fmt.Sprintf(`
resource "ovirt_storage_domain" "dataNFS" {
  name              = "testAccOvirtStorageDomainNFS"
  host_id           = "%s"
  type              = "data"
  datacenter_id     = "%s"
  description       = "nfs storage domain descriptions"
  wipe_after_delete = "true"

  nfs {
    address = "%s"
    path    = "%s"
  }
}
`, hostID, dcID, nfsAddr, nfsPath)
}
