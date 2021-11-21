resource "ovirt_vm" "test" {
  name        = "hello_world"
  cluster_id  = var.cluster_id
  template_id = "00000000-0000-0000-0000-000000000000"
}

resource "ovirt_nic" "test" {
  vnic_profile_id = var.vnic_profile_id
  vm_id           = ovirt_vm.test.id
  name            = "eth0"
}

resource "ovirt_vm_start" "test" {
  vm_id = ovirt_vm.test.id

  # Wait with the start until the NIC is attached.
  depends_on = [ovirt_nic.test]
}
