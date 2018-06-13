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

  network_interface {
    label       = "eth0"
    boot_proto  = "static"
    ip_address  = "130.20.232.184"
    gateway     = "130.20.232.1"
    subnet_mask = "255.255.255.0"
  }

  template = "Blank"
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
  bootable  = "false"
  interface = "virtio"
}

resource "ovirt_datacenter" "my_datacenter_1" {
  name              = "my_datacenter_1"
  description      = "Test create datacenter" 
  local            = false
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
