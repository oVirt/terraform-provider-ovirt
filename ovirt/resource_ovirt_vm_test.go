// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.
package ovirt_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func TestAccOvirtVM_basic(t *testing.T) {
	suite := getOvirtTestSuite(t)

	id := suite.GenerateRandomID(5)
	diskName := fmt.Sprintf("tf-test-%s-disk", id)
	vmName := fmt.Sprintf("tf-test-%s-vm", id)

	var vm ovirtsdk4.Vm
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				ResourceName: "ovirt_vm.vm",
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

  os {
    type = "other"
  }

  block_device {
    interface = "virtio"
    disk_id   = ovirt_image_transfer.disk.disk_id
    size      = 1
  }
}`,
					diskName,
					suite.TestImageSourceURL(),
					suite.StorageDomainID(),
					vmName,
					suite.ClusterID(),
					suite.BlankTemplateID()),

				Check: resource.ComposeTestCheckFunc(
					suite.EnsureVM("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", vmName),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
				),
			},
		},
		CheckDestroy: suite.EnsureVMRemoved(&vm),
	})
}

// TODO fix broken test
func DisabledTestAccOvirtVM_memory(t *testing.T) {
	suite := getOvirtTestSuite(t)

	var vm ovirtsdk4.Vm
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				ResourceName: "ovirt_vm.vm",
				Config: fmt.Sprintf(`
resource "ovirt_image_transfer" "disk" {
  alias = "vm_test_disk_memory_1"
  source_url = "%s"
  storage_domain_id = "%s"
  sparse = true
}

resource "ovirt_vm" "vm" {
  name        = "testAccVMMemory"
  cluster_id  = "%s"
  template_id = "%s"

  os {
    type = "other"
  }

  block_device {
    interface = "virtio"
    disk_id   = ovirt_image_transfer.disk.disk_id
    size      = 1
  }
}`,
					suite.TestImageSourceURL(),
					suite.StorageDomainID(),
					suite.ClusterID(),
					suite.BlankTemplateID()),

				Check: resource.ComposeTestCheckFunc(
					suite.EnsureVM("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMMemory"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					// If not provided, memory is automatically set by ovirt-engine
					resource.TestCheckResourceAttrSet("ovirt_vm.vm", "memory"),
				),
			},
			{
				ResourceName: "ovirt_vm.vm",
				// Update Terraform config to have an explicit memory setting now.
				Config: fmt.Sprintf(`
resource "ovirt_image_transfer" "disk" {
  alias = "vm_test_disk_memory_1"
  source_url = "%s"
  storage_domain_id = "%s"
  sparse = true
}

resource "ovirt_vm" "vm" {
  name        = "testAccVMMemory"
  cluster_id  = "%s"
  template_id = "%s"
  memory = 1024

  os {
    type = "other"
  }

  block_device {
    interface = "virtio"
    disk_id   = ovirt_image_transfer.disk.disk_id
    size      = 1
  }
}`,
					fmt.Sprintf("file://%s", strings.ReplaceAll(suite.TestImageSourcePath(), "\\", "/")),
					suite.StorageDomainID(),
					suite.ClusterID(),
					suite.BlankTemplateID()),

				Check: resource.ComposeTestCheckFunc(
					suite.EnsureVM("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMMemory"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "memory", "1024"),
				),
			},
		},
		CheckDestroy: suite.EnsureVMRemoved(&vm),
	})
}
