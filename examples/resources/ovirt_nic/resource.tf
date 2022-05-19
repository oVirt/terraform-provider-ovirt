data "ovirt_blank_template" "blank" {
}

resource "ovirt_vm" "test" {
  name        = random_string.vm_name.result
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
}

resource "ovirt_nic" "test" {
  name            = "eth0"
  vm_id           = ovirt_vm.test.id
  vnic_profile_id = var.vnic_profile_id
}
