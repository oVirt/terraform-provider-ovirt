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

  boot_disk = {
    disk_id      = "${ovirt_disk.my_boot_disk_2.id}"
    interface    = "virtio"
    active       = true
    logical_name = "/dev/sda"
  }

  network_interface {
    label       = "eth0"
    boot_proto  = "static"
    ip_address  = "130.20.232.184"
    gateway     = "130.20.232.1"
    subnet_mask = "255.255.255.0"
  }

  template = "Blank"
}

resource "ovirt_disk" "my_boot_disk_2" {
  name              = "my_boot_disk_2"
  alias             = "my_boot_disk_2"
  size              = 23687091200
  format            = "cow"
  storage_domain_id = "cadbe661-0e35-4fcb-a70d-2b17e2559d9c"
  sparse            = true
}

resource "ovirt_disk" "my_disk_1" {
  name              = "my_disk_1"
  alias             = "my_disk_1"
  size              = 23687091200
  format            = "cow"
  storage_domain_id = "cadbe661-0e35-4fcb-a70d-2b17e2559d9c"
  sparse            = true
}

resource "ovirt_disk_attachment" "my_diskattachment_1" {
  disk_id   = "${ovirt_disk.my_disk_1.id}"
  vm_id     = "${ovirt_vm.my_vm_1.id}"
  bootable  = false
  interface = "virtio"
}

data "ovirt_datacenters" "defaultDC" {
  name = "Default"
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
