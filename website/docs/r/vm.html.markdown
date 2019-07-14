---
layout: "ovirt"
page_title: "oVirt: ovirt_vm"
sidebar_current: "docs-ovirt-resource-vm"
description: |-
  Manages a VM resource within oVirt.
---

# ovirt\_vm

Manages a VM resource within oVirt.

## Example Usage

### Boot VM From an Existing Template (Disk)

```hcl
resource "ovirt_vm" "vm" {
  name       = "my_vm"
  cluster_id = "3e7e71ed-24ea-4812-8ef9-a09a858d31e4"
  memory     = 4096
  # there has one or more disks in the specified template
  template_id = "5ba458ca-00a4-0358-00cb-000000000223"
}
```

### Boot VM From a New Disk

```hcl
resource "ovirt_vm" "vm" {
  name       = "my_vm"
  cluster_id = "3e7e71ed-24ea-4812-8ef9-a09a858d31e4"
  memory     = 4096 # in megabytes

  block_device {
    disk_id   = "${ovirt_disk.boot_disk_1.id}"
    interface = "virtio"
  }

}

resource "ovirt_disk" "boot_disk_1" {
  name              = "boot_disk_1"
  alias             = "boot_disk_1"
  size              = 60 # in gigabytes
  format            = "cow"
  storage_domain_id = "5ba458ca-00a4-0358-00cb-000000000223"
  sparse            = true
}
```

### Boot VM From an Existing Disk

```hcl
resource "ovirt_vm" "vm" {
  name       = "my_vm"
  cluster_id = "3e7e71ed-24ea-4812-8ef9-a09a858d31e4"
  memory     = 4096 # in megabytes

  block_device {
    disk_id   = "${data.boot_disk.disks.0.id}"
    interface = "virtio"
  }

}

data "ovirt_disks" "boot_disk" {
  name_regex = "boot_disk_1"
}
```

### Attach a New Disks to VM

```hcl
resource "ovirt_vm" "vm" {
  name       = "my_vm"
  cluster_id = "3e7e71ed-24ea-4812-8ef9-a09a858d31e4"
  memory     = 4096 # in megabytes
  # there has one or more disks in the specified template
  template_id = "5ba458ca-00a4-0358-00cb-000000000223"
}

resource "ovirt_disk" "attached_disk_1" {
  name              = "attached_disk_1"
  alias             = "attached_disk_1"
  size              = 60 # in gigabytes
  format            = "cow"
  storage_domain_id = "5ba458ca-00a4-0358-00cb-000000000223"
  sparse            = true
}

resource "ovirt_disk_attachment" "attachment" {
  vm_id     = "${ovirt_vm.vm.id}"
  disk_id   = "${ovirt_disk.attached_disk_1.id}"
  bootable  = false
  interface = "virtio"
  active    = true
  read_only = false
}
```

### Attach multiple vNICs to VM

```hcl
resource "ovirt_vm" "vm" {
  name       = "my_vm"
  cluster_id = "3e7e71ed-24ea-4812-8ef9-a09a858d31e4"
  memory     = 4096 # in megabytes
  # there has one or more disks in the specified template
  template_id = "5ba458ca-00a4-0358-00cb-000000000223"
}

resource "ovirt_vnic" "nic1" {
  name            = "nic1"
  vm_id           = "${ovirt_vm.vm.id}"
  vnic_profile_id = "ce6f1f2e-7262-40f6-a005-531c9cec0f28"
}

resource "ovirt_vnic" "nic2" {
  name            = "nic2"
  vm_id           = "${ovirt_vm.vm.id}"
  vnic_profile_id = "ce6f1f2e-7262-40f6-a005-531c9cec0f28"
}
```

### VM with User Data

```hcl
resource "ovirt_vm" "my_vm_1" {
  name        = "my_vm_1"
  cluster_id  = "b0280bd4-4152-42ad-aa37-1e73ab30b0da"
  template_id = "5ba458ca-00a4-0358-00cb-000000000223"
  memory      = 4096 # in megabytes

  initialization {
    authorized_ssh_key = "${file(pathexpand("~/.ssh/id_rsa.pub"))}"
    host_name          = "vm-basic-updated"
    timezone           = "Asia/Shanghai"
    user_name          = "root"
    custom_script      = "echo hello2"
    dns_search         = "university.edu"
    dns_servers        = "8.8.8.8"

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
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the VM. Changing this creates a new VM.
* `cluster_id` - (Required) The ID of cluster the VM belongs to. Changing this creates a new VM.
* `template_id` - (Optional) The ID of template the VM based on. Default is `00000000-0000-0000-0000-000000000000`. Changing this creates a new VM.
* `memory` - (Optional) The amount of memory of the VM (in metabytes). Changing this creates a new VM.
* `cores` - (Optional) The amount of cores. Default is `1`. Changing this creates a new VM.
* `sockets` - (Optional) The amount of sockets. Default is `1`. Changing this creates a new VM.
* `threads` - (Optional) The amount of threads. Default is `1`. Changing this creates a new VM.
* `block_device` - (Optional) Configurations of bootable disk block device. The block_device structure is documented below. Changing this creates a new VM. You can specify at most one block_device.
* `initialization` - (Optional) Configurations of initialization. The initialization structure is documented below. Changint this updates the VM's initialization. You can specify at most one initialization.

The `block_device` block supports:

* `disk_id` - (Required) The ID of attached disk. Changing this creates a new disk attachment.
* `active` - (Optional) The flag to indicate whether the disk is active. Default is `true`. Changing this updates the attachment's active.
* `bootable` - (Optional) The flag to indicate whether the disk is bootable. Default is `false`. Changing this updates the attachment's bootable.
* `interface` - (Required) The interface of the attachment. Valid values are `ide`, `sata`, `spapr_vscsi`, `virtio` and `virtio_scsi`. Changing this creates a new attachment.
* `pass_discard` - (Optional) The flag to indicate whether the VM passes discard commands to the storage. Changing this creates a new attachment.
* `read_only` - (Optional) The flag to indicate whether the disk is connected to the VM as read only. Default is `false`. Changing this creates a new attachment.
* `use_scsi_reservation` - (Optional) The flag to indicate whether SCSI reservation is enabled for this disk. Default is `false`. Changing this creates a new attachment.

The `initialization` block supports:

* `host_name` - (Optional) Set the hostname for the VM.
* `timezone` - (Optional) Set the timezone for the VM.
* `user_name` - (Optional) Set the user name for the VM.
* `custom_scripit` - (Optional) Set the custom script for the VM.
* `dns_servers` - (Optional) Set the dns server for the VM.
* `dns_search` - (Optional) Set the dns server for the VM.
* `nic_configuration` - (Optional) Configurations to initilize the vNICs in VM. The nic_configuration structure is documented below. 
* `authorized_ssh_key` - (Optional) Set the ssh key for the VM. Default is `""`.

The `nic_configuration` block supports:

* `label` - (Required) Speficy the vNIC to apply this configuration.
* `boot_proto` - (Required) Set the boot protocol for the vNIC configuration. Valid values are `autoconf`, `dhcp`, `none` and `static`.
* `address` - (Optional) Set the IP address for the vNIC.
* `netmask` - (Optional) Set the netmask for the vNIC.
* `gateway` - (Optional) Set the gateway for the vNIC.
* `on_boot` - (Optional) The flag to indicate whether the vNIC will be activated at VM booting. Default is `true`.

## Attributes Reference

The following attributes are exported:

* `name` - See Argument Reference above
* `status` - The detected status of the VM.
* `cluster_id` - See Argument Reference above
* `template_id` - See Argument Reference above
* `memory` - See Argument Reference above
* `cores` - See Argument Reference above
* `sockets` - See Argument Reference above
* `threads` - See Argument Reference above
* `block_device` - See Argument Reference above
* `initialization` - See Argument Reference above