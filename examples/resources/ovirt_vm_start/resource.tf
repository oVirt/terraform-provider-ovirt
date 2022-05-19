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
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
}

resource "ovirt_disk_attachment" "test" {
  vm_id          = ovirt_vm.test.id
  disk_id        = ovirt_disk.test.id
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
  force_stop = true

  # Wait with the start until the NIC and disks are attached.
  depends_on = [ovirt_nic.test, ovirt_disk_attachment.test]
}
