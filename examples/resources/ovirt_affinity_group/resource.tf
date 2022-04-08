resource "ovirt_affinity_group" "test" {
  name      = "test"
  enforcing = true

  hosts_rule {
    affinity  = "negative"
    enforcing = true
  }
  vms_rule {
    affinity  = "negative"
    enforcing = true
  }
}