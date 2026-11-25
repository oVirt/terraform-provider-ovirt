resource "ovirt_network" "test" {
  name          = "my-network"
  datacenter_id = var.datacenter_id
  description   = "My logical network"
  comment       = "Managed by Terraform"
  vlan          = 100
}
