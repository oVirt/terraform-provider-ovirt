resource "ovirt_vm" "my_vm" {
  name = "my_first_vm"
  cluster = "Default"
  network_interface {
    label = "eth0"
    boot_proto = "static"
    ip_address = "130.20.232.184"
    gateway = "130.20.232.1"
  }
  provisioner "remote-exec" {
    inline = [
      "uptime"
    ]
  }
}
