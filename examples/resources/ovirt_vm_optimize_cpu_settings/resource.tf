data "ovirt_blank_template" "blank" {
}

resource "ovirt_vm" "test" {
  name        = "hello_world"
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
}

resource "ovirt_vm_optimize_cpu_settings" "test" {
  vm_id = ovirt_vm.test.id
}