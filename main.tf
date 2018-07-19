variable "ovirt_url" {}
variable "ovirt_username" {}
variable "ovirt_pass" {}

provider "ovirt" {
  username = "${var.ovirt_username}"
  url      = "${var.ovirt_url}"
  password = "${var.ovirt_pass}"
}

resource "ovirt_vm" "my_vm_1" {
  name               = "my_vm_1"
  cluster            = "Default"
  authorized_ssh_key = "${file(pathexpand("~/.ssh/id_rsa.pub"))}"

  #  boot_disk = {
  #    disk_id      = "${ovirt_disk.my_boot_disk_2.id}"
  #    interface    = "virtio"
  #    active       = true
  #    logical_name = "/dev/sda"
  #  }

  network_interface {
    network     = "${data.ovirt_networks.my_network_2.networks.0.name}"
    label       = "eth0"
    boot_proto  = "static"
    ip_address  = "130.20.232.184"
    gateway     = "130.20.232.1"
    subnet_mask = "255.255.255.0"
  }
  template = "Blank"
}

#resource "ovirt_disk" "my_boot_disk_2" {
#  name              = "my_boot_disk_2"
#  alias             = "my_boot_disk_2"
#  size              = 23687091200
#  format            = "cow"
#  storage_domain_id = "${data.ovirt_storagedomains.my_ds.storagedomains.0.id}"
#  sparse            = true
#}

resource "ovirt_disk" "my_disk_1" {
  name              = "my_disk_1"
  alias             = "my_disk_1"
  size              = 23687091200
  format            = "cow"
  storage_domain_id = "${data.ovirt_storagedomains.my_ds.storagedomains.0.id}"
  sparse            = true
}

resource "ovirt_disk_attachment" "my_diskattachment_1" {
  disk_id   = "${ovirt_disk.my_disk_1.id}"
  vm_id     = "${ovirt_vm.my_vm_1.id}"
  bootable  = false
  interface = "virtio"
}

resource "ovirt_datacenter" "my_datacenter_1" {
  name        = "my_datacenter_1"
  description = "Datacenter Test1"
  local       = false
}

resource "ovirt_network" "my_network_1" {
  name          = "my_network_1"
  description   = "Network Test1"
  mtu           = 1001
  datacenter_id = "${ovirt_datacenter.my_datacenter_1.id}"
}

data "ovirt_networks" "my_network_2" {
  name_regex = "^my_network_*"

  search = {
    criteria       = "datacenter = Default and name = my_network_2"
    max            = 1
    case_sensitive = false
  }
}

data "ovirt_datacenters" "defaultDC" {
  name_regex = "^De\\w*"

  search = {
    criteria       = "status = up and Storage.name = data"
    max            = 10
    case_sensitive = false
  }
}

data "ovirt_clusters" "defaultCluster" {
  name_regex = "^De\\w*"

  search = {
    criteria       = "name = Default"
    max            = 1
    case_sensitive = false
  }
}

data "ovirt_storagedomains" "my_ds" {
  name_regex = "VM_DATASTORE"
  search = {
    criteria       = "external_status = ok and datacenter = ${data.ovirt_datacenters.defaultDC.datacenters.0.name}"
    case_sensitive = false
  }
}

data "ovirt_disks" "my_disk" {
  name_regex = "^mydisk_*"

  search = {
    criteria       = "status = ok and provisioned_size > 1024000000"
    max            = 1
    case_sensitive = false
  }
}

output "default_dc_id" {
  value = "${data.ovirt_datacenters.defaultDC.datacenters.0.id}"
}

output "disk_id" {
  value = "${ovirt_disk.my_disk_1.id}"
}

output "diskattachment_id" {
  value = "${ovirt_disk_attachment.my_diskattachment_1.id}"
}

output "vm_id" {
  value = "${ovirt_vm.my_vm_1.id}"
}

output "datacenter_id" {
  value = "${ovirt_datacenter.my_datacenter_1.id}"
}

output "network_id" {
  value = "${ovirt_network.my_network_1.id}"
}
