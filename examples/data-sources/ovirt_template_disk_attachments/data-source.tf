data "ovirt_template_disk_attachments" "list" {
  template_id = ovirt_template.blueprint.id

  depends_on = [ovirt_template.blueprint]
}

output "attachment_list" {
  value = data.ovirt_template_disk_attachments.list
}