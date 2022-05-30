data "ovirt_blank_template" "blank" {
}

resource "ovirt_vm" "test" {
  name        = random_string.vm_name.result
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
}

resource "ovirt_template" "blueprint" {
  vm_id = ovirt_vm.test.id
  name  = "blueprint1"
}