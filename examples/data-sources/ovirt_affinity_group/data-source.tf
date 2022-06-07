provider "ovirt" {
  mock = true
}

data "ovirt_affinity_group" "main" {
  cluster_id = var.cluster_id
  name = "main_affinity_group"

  depends_on = [ovirt_affinity_group.main]
}

output "main_affinity_id" {
  value = data.ovirt_affinity_group.main.id
}