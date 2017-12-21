provider "ovirt" {
  username = "kschmidt@PNL.GOV"
  url = "https://ovirt.emsl.pnl.gov/ovirt-engine/api"
}

resource "ovirt_vm" "my_vm" {
  name = "my_first_vm"
  cluster = "Default"
  authorized_ssh_key = "/Users/kschmidt/.ssh/id_rsa.pub"
  network_interface {
    label = "eth0"
    boot_proto = "static"
    ip_address = "130.20.232.184"
    gateway = "130.20.232.1"
    subnet_mask = "255.255.255.0"
  }
  
  template = "centos-7.4.1707-cloudinit-mgmt"
  provisioner "remote-exec" {
    inline = [
      "uptime"
    ]
  }
}
