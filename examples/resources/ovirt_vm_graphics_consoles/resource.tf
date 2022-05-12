resource "ovirt_vm" "test" {
  name        = "hello_world"
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = "00000000-0000-0000-0000-000000000000"
}

# This will remove all graphics consoles from the specified VM. Adding graphics consoles is currently not supported.
resource "ovirt_vm_graphics_consoles" "test" {
  vm_id = ovirt_vm.test.id
}