// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtVMsDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)

	id := suite.GenerateRandomID(5)
	diskName := fmt.Sprintf("tf-test-%s", id)
	vmName := fmt.Sprintf("tf-test-%s", id)

	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "ovirt_image_transfer" "disk" {
  alias = "%s"
  source_url = "%s"
  storage_domain_id = "%s"
  sparse = true
}

resource "ovirt_vm" "vm" {
  name        = "%s"
  cluster_id  = "%s"
  template_id = "%s"
  auto_start  = false

  os {
    type = "other"
  }

  block_device {
    interface = "virtio"
    disk_id   = ovirt_image_transfer.disk.disk_id
    size      = 1
  }
}

data "ovirt_vms" "name_regex_filtered_vm" {
  name_regex = "^%s$"

  depends_on = [ovirt_vm.vm]
}
`,
					diskName,
					suite.TestImageSourceURL(),
					suite.StorageDomainID(),
					vmName,
					suite.ClusterID(),
					suite.BlankTemplateID(),
					regexp.QuoteMeta(vmName),
				),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_vms.name_regex_filtered_vm"),
					resource.TestCheckResourceAttr("data.ovirt_vms.name_regex_filtered_vm", "vms.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_vms.name_regex_filtered_vm", "vms.0.name", vmName),
				),
			},
		},
	})
}

func TestAccOvirtVMsDataSource_searchFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)

	id := suite.GenerateRandomID(5)
	diskName := fmt.Sprintf("tf_test_%s", id)
	vmName := fmt.Sprintf("tf_test_%s", id)

	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "ovirt_image_transfer" "disk" {
  alias = "%s"
  source_url = "%s"
  storage_domain_id = "%s"
  sparse = true
}

resource "ovirt_vm" "vm" {
  name        = "%s"
  cluster_id  = "%s"
  template_id = "%s"
  auto_start  = false

  os {
    type = "other"
  }

  block_device {
    interface = "virtio"
    disk_id   = ovirt_image_transfer.disk.disk_id
    size      = 1
  }
}

data "ovirt_vms" "search_filtered_vm" {
  search = {
    criteria       = "name = %s and status = down"
    max            = 1
    case_sensitive = false
  }

  depends_on = [ovirt_vm.vm]
}
`,
                    diskName,
					suite.TestImageSourceURL(),
					suite.StorageDomainID(),
					vmName,
					suite.ClusterID(),
					suite.BlankTemplateID(),
					regexp.QuoteMeta(vmName),
				),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_vms.search_filtered_vm"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.0.name", vmName),
				),
			},
		},
	})
}
