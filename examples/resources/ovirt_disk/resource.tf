resource "ovirt_disk" "test" {
  storagedomain_id = var.storagedomain_id
  format           = "raw"
  size             = 1048576
  alias            = "test"
  sparse           = true
}