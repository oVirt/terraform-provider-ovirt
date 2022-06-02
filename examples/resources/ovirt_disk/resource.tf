resource "ovirt_disk" "test" {
  storage_domain_id = var.storage_domain_id
  format           = "raw"
  size             = 1048576
  alias            = "test"
  sparse           = true
}