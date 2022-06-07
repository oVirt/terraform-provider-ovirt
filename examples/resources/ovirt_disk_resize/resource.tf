# this data-source requires a VM created in another configuration and referenced here
data "ovirt_disk_attachments" "templated" {
  vm_id = var.vm_id
}

resource "ovirt_disk_resize" "resized" {
  # looping through all the disk attachments
  # using the attachment id (a.id) as key and the disk id (a.disk_id) as value
  for_each = {for a in data.ovirt_disk_attachments.templated.attachments: a.id => a.disk_id}

  disk_id = "${each.value}"
  size = 2*1048576
}