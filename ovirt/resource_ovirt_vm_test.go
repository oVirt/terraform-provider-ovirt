// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.
package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func TestAccOvirtVM_basic(t *testing.T) {
	clusterID := "5b878de2-019e-0348-0293-000000000323"
	templateID := "333c72d1-8fa9-4968-b892-fc8c047c0b88"
	var vm ovirtsdk4.Vm
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMBasic(clusterID, templateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMBasic"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "memory", "2048"),
				),
			},
		},
	})
}

func TestAccOvirtVM_blockDevice(t *testing.T) {
	var vm ovirtsdk4.Vm
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMBlockDevice,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMBlockDevice"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "block_device.#", "1"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "block_device.0.interface", "virtio"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.#", "1"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.nic_configuration.#", "2"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.host_name", "vm-basic-1"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.timezone", "Asia/Shanghai"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.user_name", "root"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.custom_script", "echo hello"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.dns_search", "university.edu"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.dns_servers", "8.8.8.8 8.8.4.4"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.nic_configuration.0.label", "eth0"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.nic_configuration.0.address", "10.1.60.60"),
				),
			},
			{
				Config: testAccVMBlockDeviceUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMBlockDevice"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "block_device.#", "1"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "block_device.0.interface", "virtio_scsi"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.#", "1"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.nic_configuration.#", "1"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.host_name", "vm-basic-updated"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.timezone", "Asia/Shanghai"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.user_name", "root"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.custom_script", "echo hello2"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.dns_search", "university.edu"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.dns_servers", "8.8.8.8"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.authorized_ssh_key", ""),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.nic_configuration.0.label", "eth0"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.nic_configuration.0.address", "10.1.60.66"),
				),
			},
		},
	})
}

func TestAccOvirtVM_template(t *testing.T) {
	var vm ovirtsdk4.Vm
	clusterID := "5b6ab335-0251-028e-00ef-000000000326"
	templateID := "ad89bd73-941f-473a-9667-afaed8c7cbd1"
	newTemplateID := "3c24e89c-7af4-47f8-87d5-de5c4b11d25e"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMTemplate(clusterID, templateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMTemplate"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "template_id", templateID),
				),
			},
			{
				Config: testAccVMTemplateUpdate(clusterID, newTemplateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMTemplate"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "template_id", newTemplateID),
				),
			},
		},
	})
}

func TestAccOvirtVM_templateClone(t *testing.T) {
	var vm ovirtsdk4.Vm
	clusterID := "5bd12e84-025a-0171-03aa-0000000003d6"
	templateID := "02ff8100-360f-4daa-8624-e70591bac22e"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMTemplateClone(clusterID, templateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMTemplate"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "template_id", templateID),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "clone", "true"),
				),
			},
		},
	})
}

func TestAccOvirtVM_vnic(t *testing.T) {
	var vm ovirtsdk4.Vm
	clusterID := "5b6ab335-0251-028e-00ef-000000000326"
	templateID := "ad89bd73-941f-473a-9667-afaed8c7cbd1"
	vnicProfileID := "0000000a-000a-000a-000a-000000000398"
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMVnic(clusterID, templateID, vnicProfileID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMVnic"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "template_id", templateID),
					resource.TestCheckResourceAttrSet("ovirt_vnic.vm_nic1", "id"),
					resource.TestCheckResourceAttr("ovirt_vnic.vm_nic1", "name", "nic1"),
				),
			},
			{
				Config: testAccVMVnicUpdate(clusterID, templateID, vnicProfileID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMVnic"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "template_id", templateID),
					resource.TestCheckResourceAttrSet("ovirt_vnic.vm_nic2", "id"),
					resource.TestCheckResourceAttr("ovirt_vnic.vm_nic2", "name", "nic2"),
				),
			},
		},
	})
}

func TestAccOvirtVM_memory(t *testing.T) {
	var vm ovirtsdk4.Vm
	clusterID := "5bd12e84-025a-0171-03aa-0000000003d6"
	templateID := "02ff8100-360f-4daa-8624-e70591bac22e"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMMemory(clusterID, templateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMMemory"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					// If not provided, memory is automatically set by ovirt-engine
					resource.TestCheckResourceAttrSet("ovirt_vm.vm", "memory"),
				),
			},
			{
				Config: testAccVMMemoryUpdate(clusterID, templateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMMemory"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "memory", "2048"),
				),
			},
		},
	})
}

func testAccCheckVMDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_vm" {
			continue
		}
		getResp, err := conn.SystemService().VmsService().
			VmService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Vm(); ok {
			return fmt.Errorf("VM %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtVMExists(n string, v *ovirtsdk4.Vm) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No VM ID is set")
		}
		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().VmsService().
			VmService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		vm, ok := getResp.Vm()
		if ok {
			*v = *vm
			return nil
		}
		return fmt.Errorf("VM %s not exist", rs.Primary.ID)
	}
}

func testAccVMBasic(clusterID, templateID string) string {
	return fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name        = "testAccVMBasic"
  cluster_id  = "%s"
  template_id = "%s"
  memory      = 2048
  initialization {
    host_name     = "vm-basic-1"
    timezone      = "Asia/Shanghai"
    user_name     = "root"
    custom_script = "echo hello"
    dns_search    = "university.edu"
    dns_servers   = "8.8.8.8 8.8.4.4"
  }
}
`, clusterID, templateID)
}

const testAccVMBlockDevice = `
resource "ovirt_vm" "vm" {
  name       = "testAccVMBlockDevice"
  cluster_id = "5b6ab335-0251-028e-00ef-000000000326"

  initialization {
    host_name     = "vm-basic-1"
    timezone      = "Asia/Shanghai"
    user_name     = "root"
    custom_script = "echo hello"
    dns_search    = "university.edu"
    dns_servers   = "8.8.8.8 8.8.4.4"

    nic_configuration {
      label      = "eth0"
      boot_proto = "static"
      address    = "10.1.60.60"
      gateway    = "10.1.60.1"
      netmask    = "255.255.255.0"
    }
    nic_configuration {
      label      = "eth1"
      boot_proto = "static"
      address    = "10.1.60.61"
      gateway    = "10.1.60.1"
      netmask    = "255.255.255.0"
    }
  }

  block_device {
    disk_id   = "${ovirt_disk.vm_disk.id}"
    interface = "virtio"
  }
}

resource "ovirt_disk" "vm_disk" {
  name              = "vm_disk"
  alias             = "vm_disk"
  size              = 23687091200
  format            = "cow"
  storage_domain_id = "f78ab25e-ee16-42fe-80fa-b5f86b35524d"
  sparse            = true
}
`

const testAccVMBlockDeviceUpdate = `
resource "ovirt_vm" "vm" {
  name       = "testAccVMBlockDevice"
  cluster_id = "5b6ab335-0251-028e-00ef-000000000326"

  initialization {
    host_name     = "vm-basic-updated"
    timezone      = "Asia/Shanghai"
    user_name     = "root"
    custom_script = "echo hello2"
    dns_search    = "university.edu"
    dns_servers   = "8.8.8.8"
    nic_configuration {
      label      = "eth0"
      boot_proto = "static"
      address    = "10.1.60.66"
      gateway    = "10.1.60.1"
      netmask    = "255.255.255.0"
    }
  }

  block_device {
    disk_id   = "${ovirt_disk.vm_disk.id}"
    interface = "virtio_scsi"
  }
}

resource "ovirt_disk" "vm_disk" {
  name              = "vm_disk"
  alias             = "vm_disk"
  size              = 23687091200
  format            = "cow"
  storage_domain_id = "f78ab25e-ee16-42fe-80fa-b5f86b35524d"
  sparse            = true
}
`

func testAccVMTemplate(clusterID, templateID string) string {
	return fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name              = "testAccVMTemplate"
  cluster_id        = "%s"
  template_id       = "%s"
  high_availability = true
}
`, clusterID, templateID)
}

func testAccVMTemplateUpdate(clusterID, templateID string) string {
	return fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name        = "testAccVMTemplate"
  cluster_id  = "%s"
  template_id = "%s"

  block_device {
    disk_id   = "${ovirt_disk.vm_disk.id}"
    interface = "virtio"
  }
}

resource "ovirt_disk" "vm_disk" {
  name              = "vm_disk"
  alias             = "vm_disk"
  size              = 23687091200
  format            = "cow"
  storage_domain_id = "f78ab25e-ee16-42fe-80fa-b5f86b35524d"
  sparse            = true
}
`, clusterID, templateID)
}

func testAccVMTemplateClone(clusterID, templateID string) string {
	return fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name              = "testAccVMTemplate"
  cluster_id        = "%s"
  template_id       = "%s"
  high_availability = true
  clone             = true
}
`, clusterID, templateID)
}

func testAccVMVnic(clusterID, templateID, vnicProfileID string) string {
	return fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name        = "testAccVMVnic"
  cluster_id  = "%s"
  template_id = "%s"
}

resource "ovirt_vnic" "vm_nic1" {
  vm_id           = "${ovirt_vm.vm.id}"
  name            = "nic1"
  vnic_profile_id = "%s"
}
`, clusterID, templateID, vnicProfileID)
}

func testAccVMVnicUpdate(clusterID, templateID, vnicProfileID string) string {
	return fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name        = "testAccVMVnic"
  cluster_id  = "%s"
  template_id = "%s"
}

resource "ovirt_vnic" "vm_nic2" {
  vm_id           = "${ovirt_vm.vm.id}"
  name            = "nic2"
  vnic_profile_id = "%s"
}
`, clusterID, templateID, vnicProfileID)
}

func testAccVMMemory(clusterID, templateID string) string {
	return fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name              = "testAccVMMemory"
  cluster_id        = "%s"
  template_id       = "%s"
  high_availability = true
}
`, clusterID, templateID)
}

func testAccVMMemoryUpdate(clusterID, templateID string) string {
	return fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name              = "testAccVMMemory"
  cluster_id        = "%s"
  template_id       = "%s"
  memory            = 2048
  high_availability = true
}
`, clusterID, templateID)
}
