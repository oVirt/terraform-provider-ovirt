data "ovirt_blank_template" "blank" {
}

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
  template_id = data.ovirt_blank_template.blank.id
}

resource "ovirt_disk_attachments" "test" {
  vm_id = ovirt_vm.test.id
  # Set the following to true to completely remove non-listed attached disks. This can be used to wipe disks from the
  # template.
  remove_unmanaged = false

  # You can repeat this section as many times as you need.
  attachment {
    disk_id        = ovirt_disk.test.id
    disk_interface = "virtio_scsi"
  }
}