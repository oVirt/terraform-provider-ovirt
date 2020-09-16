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

func TestAccOvirtVM_basic(t *testing.T) {
	var vm ovirtsdk4.Vm
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMBasic"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "memory", "2048"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.dns_servers", "8.8.8.8 8.8.4.4"),
				),
			},
			{
				Config: testAccVMBasicUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMBasic"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "memory", "2048"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "initialization.0.dns_servers", "114.114.114.114"),
				),
			},
		},
	})
}

func TestAccOvirtVM_bootDevice(t *testing.T) {
	var vm ovirtsdk4.Vm
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMBootDevice(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMBootDevice"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "boot_devices.#", "1"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "boot_devices.0", "network"),
				),
			},
		},
	})
}

func TestAccOvirtVM_noBootDevice(t *testing.T) {
	var vm ovirtsdk4.Vm
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMNoBootDevice(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMNoBootDevice"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
				),
			},
		},
	})
}

func TestAccOvirtVM_blockDevice(t *testing.T) {
	var vm ovirtsdk4.Vm
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMBlockDevice(),
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
				Config: testAccVMBlockDeviceUpdate(),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMTemplate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMTemplate"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
				),
			},
			//{
			//	Config: testAccVMTemplateUpdate(),
			//	Check: resource.ComposeTestCheckFunc(
			//		testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
			//		resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMTemplate"),
			//		resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
			//	),
			//},
		},
	})
}

func TestAccOvirtVM_templateClone(t *testing.T) {
	var vm ovirtsdk4.Vm
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMTemplateClone(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMTemplateClone"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "clone", "true"),
				),
			},
		},
	})
}

func TestAccOvirtVM_vnic(t *testing.T) {
	var vm ovirtsdk4.Vm
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMVnic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMVnic"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttrSet("ovirt_vnic.vm_nic1", "id"),
					resource.TestCheckResourceAttr("ovirt_vnic.vm_nic1", "name", "nic1"),
				),
			},
			{
				Config: testAccVMVnicUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMVnic"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					resource.TestCheckResourceAttrSet("ovirt_vnic.vm_nic2", "id"),
					resource.TestCheckResourceAttr("ovirt_vnic.vm_nic2", "name", "nic2"),
				),
			},
		},
	})
}

func TestAccOvirtVM_memory(t *testing.T) {
	var vm ovirtsdk4.Vm
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  testAccCheckVMDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMMemory(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "name", "testAccVMMemory"),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "status", "up"),
					// If not provided, memory is automatically set by ovirt-engine
					resource.TestCheckResourceAttrSet("ovirt_vm.vm", "memory"),
				),
			},
			{
				Config: testAccVMMemoryUpdate(),
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

func TestAccOvirtVM_OperatingSystem(t *testing.T) {
	var vm ovirtsdk4.Vm
	os := "rhcos_x64"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_vm.vm",
		CheckDestroy:  func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: testAccVMOperatingSystem(os),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtVMExists("ovirt_vm.vm", &vm),
					resource.TestCheckResourceAttr("ovirt_vm.vm", "os.0.type", os),
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

const testAccVMDef = `
data "ovirt_datacenters" "d" {
  search = {
    criteria = "name = Default"
  }
}

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

data "ovirt_storagedomains" "s" {
  search = {
    criteria = "name = data"
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
  datacenter_id     = data.ovirt_datacenters.d.datacenters.0.id
  cluster_id        = data.ovirt_clusters.c.clusters.0.id
  host_id           = data.ovirt_hosts.h.hosts.0.id
  storage_domain_id = data.ovirt_storagedomains.s.storagedomains.0.id
  vnic_profile_id   = data.ovirt_vnic_profiles.v.vnic_profiles.0.id
  template_id       = data.ovirt_templates.t.templates.0.id
}
`


func testAccVMBasic() string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name        = "testAccVMBasic"
  cluster_id  = local.cluster_id
  template_id = local.template_id
  memory      = 2048
  initialization {
    host_name     = "vm-basic-1"
    timezone      = "Asia/Shanghai"
    user_name     = "root"
    custom_script = "echo hello"
    dns_search    = "university.edu"
    dns_servers   = "8.8.8.8 8.8.4.4"
  }
  os {
    type = "other"
  }
}
`)
}

func testAccVMBasicUpdate() string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
	name        = "testAccVMBasic"
	cluster_id  = local.cluster_id
	template_id = local.template_id
	memory      = 2048
	initialization {
	  host_name     = "vm-basic-1"
	  timezone      = "Asia/Shanghai"
	  user_name     = "root"
	  custom_script = "echo hello"
	  dns_search    = "university.edu"
	  dns_servers   = "114.114.114.114"
	}
  os {
    type = "other"
  }
  }
`)
}

func testAccVMBootDevice() string {
	return testAccVMDef + `
resource "ovirt_vm" "vm" {
  name       = "testAccVMBootDevice"
  cluster_id = local.cluster_id

  block_device {
    disk_id   = ovirt_disk.vm_disk.id
    interface = "virtio"
  }

  nics {
	name            = "data"
	vnic_profile_id = local.vnic_profile_id
  }

  boot_devices = ["network"]
  os {
    type = "other"
  }
}

resource "ovirt_disk" "vm_disk" {
  name              = "vm_disk_boot"
  alias             = "vm_disk_boot"
  size              = 1
  format            = "cow"
  storage_domain_id = local.storage_domain_id
  sparse            = true
}
`
}

func testAccVMNoBootDevice() string {
	return testAccVMDef + `
resource "ovirt_vm" "vm" {
  name       = "testAccVMNoBootDevice"
  cluster_id = local.cluster_id
   
  block_device {
    disk_id   = ovirt_disk.vm_disk.id
    interface = "virtio"
  }
  
  nics {
    name            = "data"
    vnic_profile_id = local.vnic_profile_id
  }
  os {
    type = "other"
  }
}
  
resource "ovirt_disk" "vm_disk" {
  name              = "vm_disk_noboot"
  alias             = "vm_disk_noboot"
  size              = 1
  format            = "cow"
  storage_domain_id = local.storage_domain_id
  sparse            = true
}
`
}

func testAccVMBlockDevice() string {
	return testAccVMDef + `
resource "ovirt_vm" "vm" {
  name       = "testAccVMBlockDevice"
  cluster_id = local.cluster_id

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
    disk_id   = ovirt_disk.vm_disk.id
    interface = "virtio"
  }
  os {
    type = "other"
  }
}

resource "ovirt_disk" "vm_disk" {
  name              = "vm_disk_blockdevice"
  alias             = "vm_disk_blockdevice"
  size              = 2
  format            = "cow"
  storage_domain_id = local.storage_domain_id
  sparse            = true
}
`
}

func testAccVMBlockDeviceUpdate() string {
	return testAccVMDef + `
resource "ovirt_vm" "vm" {
  name       = "testAccVMBlockDevice"
  cluster_id = local.cluster_id

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
    disk_id   = ovirt_disk.vm_disk.id
    interface = "virtio_scsi"
  }
  os {
    type = "other"
  }
}

resource "ovirt_disk" "vm_disk" {
  name              = "vm_disk_blockdevice"
  alias             = "vm_disk_blockdevice"
  size              = 2
  format            = "cow"
  storage_domain_id = local.storage_domain_id
  sparse            = true
}
`
}

func testAccVMTemplate() string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name              = "testAccVMTemplate"
  cluster_id        = local.cluster_id
  template_id       = local.template_id
  high_availability = true
  os {
    type = "other"
  }
}
`)
}

func testAccVMTemplateUpdate() string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name        = "testAccVMTemplate"
  cluster_id  = local.cluster_id
  template_id = local.template_id

  block_device {
    disk_id   = ovirt_disk.vm_disk.id
    interface = "virtio"
  }
  os {
    type = "other"
  }
}

resource "ovirt_disk" "vm_disk" {
  name              = "vm_disk_template"
  alias             = "vm_disk_template"
  size              = 2
  format            = "cow"
  storage_domain_id = local.storage_domain_id
  sparse            = true
}
`)
}

func testAccVMTemplateClone() string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name              = "testAccVMTemplateClone"
  cluster_id        = local.cluster_id
  template_id       = local.template_id
  high_availability = true
  clone             = true
  os {
    type = "other"
  }
}
`)
}

func testAccVMVnic() string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name        = "testAccVMVnic"
  cluster_id  = local.cluster_id
  template_id = local.template_id
  os {
    type = "other"
  }
}

resource "ovirt_vnic" "vm_nic1" {
  vm_id           = ovirt_vm.vm.id
  name            = "nic1"
  vnic_profile_id = local.vnic_profile_id
}
`)
}

func testAccVMVnicUpdate() string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name        = "testAccVMVnic"
  cluster_id  = local.cluster_id
  template_id = local.template_id
  os {
    type = "other"
  }
}

resource "ovirt_vnic" "vm_nic2" {
  vm_id           = ovirt_vm.vm.id
  name            = "nic2"
  vnic_profile_id = local.vnic_profile_id
}
`)
}

func testAccVMMemory() string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name              = "testAccVMMemory"
  cluster_id        = local.cluster_id
  template_id       = local.template_id
  high_availability = true
  os {
    type = "other"
  }
}
`)
}

func testAccVMMemoryUpdate() string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name              = "testAccVMMemory"
  cluster_id        = local.cluster_id
  template_id       = local.template_id
  memory            = 2048
  high_availability = true
  os {
    type = "other"
  }
}
`)
}

func testAccVMOperatingSystem(os string) string {
	return testAccVMDef + fmt.Sprintf(`
resource "ovirt_vm" "vm" {
  name              = "testAccVMOperatingSystem"
  cluster_id        = local.cluster_id
  template_id       = local.template_id
  memory            = 1024
  os {
    type = "%s"
  }
  initialization {
    custom_script = ""
    host_name     = "master-1"
  }
}
`, os)
}
