data "ovirt_cluster_hosts" "list" {
  cluster_id = var.cluster_id
}

output "attachment_set" {
  value = data.ovirt_cluster_hosts.list
}