resource "ovirt_affinity_group" "main" {
    cluster_id = var.cluster_id
    name = "main_affinity_group"
}
