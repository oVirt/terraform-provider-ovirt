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

func TestAccOvirtStorageDomain_nfs(t *testing.T) {
	var sd ovirtsdk4.StorageDomain
	nfsAddr, nfsPath := "10.1.110.18", "/data161"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckStorageDomainDestroy,
		IDRefreshName: "ovirt_storage_domain.dataNFS",
		Steps: []resource.TestStep{
			{
				Config: testAccStorageDomainNFS(nfsAddr, nfsPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorageDomainExists("ovirt_storage_domain.dataNFS", &sd),
					resource.TestCheckResourceAttr("ovirt_storage_domain.dataNFS", "name", "testAccOvirtStorageDomainNFS"),
					resource.TestCheckResourceAttr("ovirt_storage_domain.dataNFS", "nfs.#", "1"),
					resource.TestCheckResourceAttr("ovirt_storage_domain.dataNFS", "nfs.0.address", nfsAddr),
					resource.TestCheckResourceAttr("ovirt_storage_domain.dataNFS", "nfs.0.path", nfsPath),
				),
			},
		},
	})
}

func testAccCheckStorageDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
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
		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
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

func testAccStorageDomainNFS(nfsAddr, nfsPath string) string {
	return fmt.Sprintf(`
data "ovirt_datacenters" "d" {
  search = {
    criteria = "name = Default"
  }
}

data "ovirt_hosts" "h" {
  search = {
    criteria = "name = host65" 
  }
}

locals {
  datacenter_id = data.ovirt_datacenters.d.datacenters.0.id
  host_id       = data.ovirt_hosts.h.hosts.0.id
}


resource "ovirt_storage_domain" "dataNFS" {
  name              = "testAccOvirtStorageDomainNFS"
  host_id           = local.host_id
  type              = "data"
  datacenter_id     = local.datacenter_id
  description       = "nfs storage domain descriptions"
  wipe_after_delete = "true"

  nfs {
    address = "%s"
    path    = "%s"
  }
}
`, nfsAddr, nfsPath)
}
