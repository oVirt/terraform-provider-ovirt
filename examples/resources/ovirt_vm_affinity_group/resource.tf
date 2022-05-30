resource "ovirt_affinity_group" "ag1" {
  name       = "affinity_group_1"
  enforcing  = true
  cluster_id = var.cluster_id

  hosts_rule {
    affinity  = "negative"
    enforcing = true
  }
  vms_rule {
    affinity  = "negative"
    enforcing = true
  }
}

data "ovirt_blank_template" "blank" {}

resource "ovirt_vm" "vm1" {
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
  name        = "vm_1"
}

resource "ovirt_vm_affinity_group" "vm1_to_ag1" {
  cluster_id = var.cluster_id
  vm_id = ovirt_vm.vm1.id
  affinity_group_id = ovirt_affinity_group.ag1.id

  depends_on = [ovirt_vm.vm1, ovirt_affinity_group.ag1]
}
