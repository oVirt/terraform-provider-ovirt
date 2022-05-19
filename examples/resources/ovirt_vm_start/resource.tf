resource "ovirt_disk_from_image" "test" {
  storagedomain_id = var.storagedomain_id
  format           = "raw"
  alias            = "test"
  sparse           = true
  source_file      = "./testimage/full.qcow"
}

resource "ovirt_vm" "test" {
  name        = random_string.vm_name.result
  cluster_id  = var.cluster_id
  template_id = "00000000-0000-0000-0000-000000000000"
}

resource "ovirt_disk_attachment" "test" {
  vm_id          = ovirt_vm.test.id
  disk_id        = ovirt_disk_from_image.test.id
  disk_interface = "virtio_scsi"
}

resource "ovirt_nic" "test" {
  vnic_profile_id = var.vnic_profile_id
  vm_id           = ovirt_vm.test.id
  name            = "eth0"
}

resource "ovirt_vm_start" "test" {
  vm_id = ovirt_vm.test.id
  // How to stop the VM. Defaults to "shutdown" for an ACPI shutdown.
  stop_behavior = "stop"
  // Force-stop the VM even if a backup is currently running.
  force_stop = false

  # Wait with the start until the NIC and disks are attached.
  depends_on = [ovirt_nic.test, ovirt_disk_attachment.test]
}
