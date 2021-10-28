resource "ovirt_disk" "test" {
  storagedomain_id = var.storagedomain_id
  format           = "raw"
  size             = 512
  alias            = "test"
  sparse           = true
}