resource "ovirt_storage_domain" "test" {
  name                          = "my-storage-domain"
  host_id                       = var.host_id
  storage_type                  = "nfs"
  storage_address               = "nfs.example.com"
  storage_path                  = "/data/export"
  storage_format                = "v5"
  description                   = "NFS Production Storage Domain"
  comment                       = "Managed by Terraform"
  wipe_after_delete             = true
  critical_space_action_blocker = 5
  warning_low_space_indicator   = 10
}
