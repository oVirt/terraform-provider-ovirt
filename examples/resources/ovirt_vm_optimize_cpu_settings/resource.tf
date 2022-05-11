resource "ovirt_vm" "test" {
  name        = "hello_world"
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = "00000000-0000-0000-0000-000000000000"
}

resource "ovirt_vm_optimize_cpu_settings" "test" {
  vm_id = ovirt_vm.test.id
}