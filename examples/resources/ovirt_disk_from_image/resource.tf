resource "ovirt_disk_from_image" "test" {
  storagedomain_id = var.storagedomain_id
  format           = "raw"
  alias            = "test"
  sparse           = true
  source_file      = "./testimage/testimage"
}
