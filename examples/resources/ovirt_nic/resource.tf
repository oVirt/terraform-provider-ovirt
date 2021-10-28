resource "ovirt_vm" "test" {
  name        = "hello_world"
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = "00000000-0000-0000-0000-000000000000"
}

resource "ovirt_nic" "test" {
  name            = "eth0"
  vm_id           = ovirt_vm.test.id
  vnic_profile_id = var.vnic_profile_id
}
