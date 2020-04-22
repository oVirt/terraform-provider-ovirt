provider "ovirt" {
  username = "admin@internal"
  url      = "https://enginefqdn/ovirt-engine/api"
  password = "secret"
}

data "ovirt_clusters" "c" {
  search = {
    criteria       = ""
    case_sensitive = false
  }
}

data "ovirt_vnic_profiles" "vnic_profiles" {
  name_regex = "ovirtmgmt"
  network_id = local.network_id[0]
}

data "ovirt_networks" "n" {
  search = {
    criteria       = ""
    case_sensitive = false
  }
}

data "ovirt_templates" "t" {
  search = {
    criteria       = ""
    case_sensitive = false
  }
}

locals {
  new_id        = [for t in data.ovirt_templates.t.templates : t.id if substr(t.name, 0, 5) != "Blank"]
  cluster       = [for c in data.ovirt_clusters.c.clusters : c if c.name == "Default"]
  datacenter_id = local.cluster.0.datacenter_id
  network_id    = [for n in local.cluster.0.networks : n.id if n.name == "ovirtmgmt"]
}

output "vnic_profile_id" {
  value = data.ovirt_vnic_profiles.vnic_profiles.vnic_profiles.0.id
}

output "networks" {
  value = local.network_id
}

output "cluster_networks" {
  value = local.cluster.0.networks
}
