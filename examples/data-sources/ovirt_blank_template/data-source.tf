data "ovirt_blank_template" "blank" {
}

output "attachment_set" {
  value = data.ovirt_blank_template.blank.id
}