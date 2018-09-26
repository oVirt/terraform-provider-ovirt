provider "ovirt" {
  username = "${var.ovirt_username}"
  url      = "${var.ovirt_url}"
  password = "${var.ovirt_pass}"
}

resource "ovirt_vm" "my_vm_1" {
  name       = "my_vm_1"
  cluster_id = "${var.cluster_id}"

  memory = 1024 # in MiB

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
  #   1. One of them must be assgined
  #   2. If the template speficified by `template_id` contains disks attached,
  #      `block_device` can not be assigend
  #   3. If the template speficified by `template_id` has no disks attached,
  #      `block_device` must be assigned

  template_id = "${var.template_id}"
  block_device {
    disk_id   = "${ovirt_disk.my_disk_1.id}"
    interface = "virtio"
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
