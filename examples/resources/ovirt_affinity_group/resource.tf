resource "ovirt_affinity_group" "test" {
  name       = "test"
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