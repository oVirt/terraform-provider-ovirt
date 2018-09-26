output "disk_id" {
  value = "${ovirt_disk.my_disk_1.id}"
}

output "vm_id" {
  value = "${ovirt_vm.my_vm_1.id}"
}
