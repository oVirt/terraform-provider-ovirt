data "ovirt_blank_template" "blank" {
}

data "ovirt_cluster_hosts" "list" {
  cluster_id = var.cluster_id
}

output "attachment_set" {
  value = data.ovirt_cluster_hosts.list
}

resource "ovirt_vm" "test" {
  name                      = "hello_world"
  comment                   = "Hello world!"
  cluster_id                = var.cluster_id
  template_id               = data.ovirt_blank_template.blank.id
  placement_policy_host_ids = data.ovirt_cluster_hosts.list.hosts.*.id
  placement_policy_affinity = "migratable"
}

resource "ovirt_vm_optimize_cpu_settings" "test" {
  vm_id = ovirt_vm.test.id
}