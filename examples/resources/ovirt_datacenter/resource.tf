resource "ovirt_datacenter" "test" {
  name        = "my-datacenter"
  description = "Production Datacenter"
  comment     = "Managed by Terraform"
  local       = false
  quota_mode  = "audit"
}
