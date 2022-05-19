data "ovirt_blank_template" "blank" {
}

resource "ovirt_tag" "test" {
  name = random_string.tag_name.result
}

resource "ovirt_vm" "test" {
  name        = random_string.vm_name.result
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
}

resource "ovirt_vm_tag" "test" {
  tag_id = ovirt_tag.test.id
  vm_id  = ovirt_vm.test.id
}
