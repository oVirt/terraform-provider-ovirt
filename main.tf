variable "ovirt_pass" {}

provider "ovirt" {
  username = "kschmidt@PNL.GOV"
  url = "https://ovirt.emsl.pnl.gov/ovirt-engine/api"
  password = "${var.ovirt_pass}"
}

resource "ovirt_vm" "my_vm" {
  name = "my_first_vm"
  cluster = "Default"
  authorized_ssh_key = "${file(pathexpand("~/.ssh/id_rsa.pub"))}"
  network_interface {
    label = "eth0"
    boot_proto = "static"
    ip_address = "130.20.232.184"
    gateway = "130.20.232.1"
    subnet_mask = "255.255.255.0"
  }

  attached_disks = [{
    disk_id = "${ovirt_disk.my_disk.id}"
    bootable = "false"
    interface  = "virtio"
  }]

  template = "centos-7.4.1707-cloudinit-mgmt"
  provisioner "remote-exec" {
    inline = [
      "uptime"
    ]
  }
}

resource "ovirt_disk" "my_disk" {
  name = "my_first_disk"
  size = 1024
  format = "cow"
  storage_domain_id = "dfe8e7be-e495-49a7-be2d-71aba891ceb4"
  sparse = true
}
