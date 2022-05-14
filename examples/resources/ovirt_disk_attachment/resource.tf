resource "ovirt_disk" "test" {
  storagedomain_id = var.storagedomain_id
  format           = "raw"
  size             = 1048576
  alias            = "test"
  sparse           = true
}

resource "ovirt_vm" "test" {
  name        = random_string.vm_name.result
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = "00000000-0000-0000-0000-000000000000"
}

resource "ovirt_disk_attachment" "test" {
  vm_id          = ovirt_vm.test.id
  disk_id        = ovirt_disk.test.id
  disk_interface = "virtio_scsi"
}