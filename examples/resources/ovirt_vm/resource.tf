resource "ovirt_vm" "test" {
  name        = "hello_world"
  comment     = "Hello world!"

  cluster_id  = var.cluster_id
  template_id = "00000000-0000-0000-0000-000000000000"

  cpu_cores = 2
  cpu_threads = 3
  cpu_sockets = 4
  memory = 2147483648
  os {
    type = "rhcos_x64"
  }
  initialization {
    custom_script = "echo 'Hello world!'"
    hostname = "test"
  }
}

resource "ovirt_disk" "test" {
  storagedomain_id = var.storagedomain_id
  format           = "cow"
  size             = 1048576
  alias            = "test"
  sparse           = true
}

resource "ovirt_disk_attachment" "test" {
  vm_id = ovirt_vm.test.id
  disk_id = ovirt_disk.test.id
  disk_interface = "virtio_scsi"
}

resource "ovirt_vm_start" "test" {
  vm_id = ovirt_vm.test.id

  // Add a dependency to the disk attachment so the VM doesn't start until the disk is added.
  depends_on = [ovirt_disk_attachment.test]
}