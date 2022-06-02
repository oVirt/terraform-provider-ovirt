data "ovirt_blank_template" "blank" {
}

resource "ovirt_disk" "test" {
  storage_domain_id = var.storage_domain_id
  format           = "raw"
  size             = 1048576
  alias            = "test"
  sparse           = true
}

resource "ovirt_vm" "test" {
  name        = random_string.vm_name.result
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
}

resource "ovirt_disk_attachment" "test" {
  vm_id          = ovirt_vm.test.id
  disk_id        = ovirt_disk.test.id
  disk_interface = "virtio_scsi"
}