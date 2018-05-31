variable "ovirt_username" {}
variable "ovirt_url" {}
variable "ovirt_pass" {}

provider "ovirt" {
  username = "${var.ovirt_username}"
  url      = "${var.ovirt_url}"
  password = "${var.ovirt_pass}"
}

resource "ovirt_vm" "joey_vm_1" {
  name               = "joey_vm_1"
  cluster            = "Default"
  authorized_ssh_key = "${file(pathexpand("~/.ssh/id_rsa.pub"))}"

  network_interface {
    label       = "eth0"
    boot_proto  = "static"
    ip_address  = "130.20.232.184"
    gateway     = "130.20.232.1"
    subnet_mask = "255.255.255.0"
  }

  attached_disks = [{
    disk_id   = "${ovirt_disk.joey_disk_1.id}"
    bootable  = "false"
    interface = "virtio"
  }]

  template = "Blank"

  provisioner "remote-exec" {
    inline = [
      "uptime",
    ]
  }
}

resource "ovirt_disk" "joey_disk_1" {
  name              = "joey_disk_1"
  size              = 33687091200
  format            = "cow"
  storage_domain_id = "cadbe661-0e35-4fcb-a70d-2b17e2559d9c"
  sparse            = true
}
