resource "ovirt_vm" "test" {
  name        = random_string.vm_name.result
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = "00000000-0000-0000-0000-000000000000"
}

resource "ovirt_template" "blueprint" {
  vm_id	= ovirt_vm.test.id
  name	= "blueprint1"
}