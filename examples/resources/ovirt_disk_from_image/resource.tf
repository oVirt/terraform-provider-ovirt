resource "ovirt_disk_from_image" "test" {
  storage_domain_id = var.storage_domain_id
  format           = "raw"
  alias            = "test"
  sparse           = true
  source_file      = "./testimage/image"
}
