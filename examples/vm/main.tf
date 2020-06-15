provider "ovirt" {
  username = "${var.ovirt_username}"
  url      = "${var.ovirt_url}"
  password = "${var.ovirt_pass}"
}

resource "ovirt_vm" "my_vm_1" {
  name       = "my_vm_1"
  cluster_id = "${var.cluster_id}"

  # Instance type sets memory/cpu and more details of the VM.
  instance_type_id = "0000000b-000b-000b-000b-00000000021f" // this is XLarge

  memory = 1024 # in MiB - don't mix with instance_type_id

  initialization {
    authorized_ssh_key = "${file(pathexpand("~/.ssh/id_rsa.pub"))}"

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

  # The `template_id` and `block_device` need to satisfy the following constraints:
  #   1. One of them must be assigned.
  #   2. If the template specified by `template_id` contains disks attached,
  #      `block_device` must be assigned without 'disk_id'. This will automatically
  #       update the bootable disk of VM (the disk that copied from the template).
  #   3. If the template specified by `template_id` has no disks attached,
  #      `block_device` must be assigned with 'disk_id'. This will attach a new disk.
  #   4. If will be configured alias and storage_domain terraform provider will define
  #      alias for all disks, except bootable disk. Alias will be in format: vmname_Disk2,
  #      vmname_Disk3

  template_id = "${var.template_id}"
  block_device {
    disk_id   = "${ovirt_disk.my_disk_1.id}"  // optional
    interface = "virtio"
    size      = 120   // size in GiB - in case disk_id is not passed, this would extend the disk.
    alias          = "my_vm_1" // optional. human friendly disk name on the disks list.
    storage_domain = "lun-5TB" // optional. define storage domain where will be created disks for VM.
                               // Applied only for deploy VM from Template. disk_id shoud be removed.
  }
}

resource "ovirt_vnic" "nic1" {
  name            = "nic1"
  vm_id           = "${ovirt_vm.my_vm_1.id}"
  vnic_profile_id = "${ovirt_vnic_profile.vm_vnic_profile.id}"
}

resource "ovirt_vnic" "nic2" {
  name            = "nic2"
  vm_id           = "${ovirt_vm.my_vm_1.id}"
  vnic_profile_id = "${ovirt_vnic_profile.vm_vnic_profile.id}"
}

resource "ovirt_disk" "my_disk_1" {
  name              = "my_disk_1"
  alias             = "my_disk_1"
  size              = 2                         # in GiB
  format            = "cow"
  storage_domain_id = "${var.storagedomain_id}"
  sparse            = true
}

resource "ovirt_vnic_profile" "vm_vnic_profile" {
  name           = "vm_vnic_profile"
  network_id     = "${var.network_id}"
  migratable     = true
  port_mirroring = true
}
