data "ovirt_disk_attachments" "set" {
  vm_id = ovirt_vm.test.id
  depends_on = [
    ovirt_disk_attachments.test
  ] 
}

output "attachment_set" {
  value = data.ovirt_disk_attachments.set
}