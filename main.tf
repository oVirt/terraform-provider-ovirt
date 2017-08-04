resource "ovirt_vm" "my_vm" {
  name = "my_first_vm"
  cluster = "Default"
  template = "centos-7-b"
}
