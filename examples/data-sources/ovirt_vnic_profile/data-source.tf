data "ovirt_vnic_profile" "mgmt" {
  name = "ovirtmgmt"
}

output "vnic_profile_id" {
  value = data.ovirt_vnic_profile.mgmt.id
}
