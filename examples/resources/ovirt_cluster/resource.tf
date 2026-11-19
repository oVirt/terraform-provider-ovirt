resource "ovirt_cluster" "test" {
  name                              = "my-cluster"
  datacenter_id                     = var.datacenter_id
  cpu_type                          = "Intel SandyBridge Family"
  cpu_arch                          = "x86_64"
  description                       = "Production Compute Cluster"
  comment                           = "Managed by Terraform"
  ballooning                        = true
  memory_policy_over_commit_percent = 150
}
