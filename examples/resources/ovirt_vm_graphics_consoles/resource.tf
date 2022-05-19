data "ovirt_blank_template" "blank" {
}

resource "ovirt_vm" "test" {
  name        = "hello_world"
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
}

# This will remove all graphics consoles from the specified VM. Adding graphics consoles is currently not supported.
resource "ovirt_vm_graphics_consoles" "test" {
  vm_id = ovirt_vm.test.id
}