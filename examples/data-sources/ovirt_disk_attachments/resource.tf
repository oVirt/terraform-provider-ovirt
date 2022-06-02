data "ovirt_blank_template" "blank" {
}

resource "ovirt_disk" "test1" {
  storage_domain_id = var.storage_domain_id
  format           = "raw"
  size             = 1048576
  alias            = "test"
  sparse           = true
}

resource "ovirt_disk" "test2" {
  storage_domain_id = var.storage_domain_id
  format           = "raw"
  size             = 1048576
  alias            = "test"
  sparse           = true
}

resource "ovirt_vm" "test" {
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
  name        = "test"
}

resource "ovirt_disk_attachments" "test" {
  vm_id = ovirt_vm.test.id

  attachment {
    disk_id        = ovirt_disk.test1.id
    disk_interface = "virtio_scsi"
  }
  attachment {
    disk_id        = ovirt_disk.test2.id
    disk_interface = "virtio_scsi"
  }

  depends_on = [
    ovirt_vm.test, ovirt_disk.test1, ovirt_disk.test2
  ]
}
