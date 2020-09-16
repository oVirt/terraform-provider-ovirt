provider "ovirt" {
  username = "admin@internal"
  url      = "https://myengine/ovirt-engine/api"
  password = "admin"
}

data "ovirt_datacenters" "default" {
  search = {
    criteria = "name = Default"
  }
}

data "ovirt_clusters" "default" {
  search = {
    criteria = "name = Default"
  }
}

data "ovirt_networks" "ovirtmgmt" {
  search = {
    criteria = "name = ovirtmgmt and datacenter = Default"
  }
}

locals {
  datastores = {
    "data"           = "/data/images/data"
    "DEV_datastore"  = "/data/images/dev"
    "MAIN_datastore" = "/data/images/main"
    "DS_INTERNAL"    = "/data/images/internal"
  }

  disks = {
    "test_disk1"    = 10
    "test_disk2"    = 10
    "hosted_engine" = 10
  }

  default_datacenter_id = data.ovirt_datacenters.default.datacenters.0.id
  default_cluster_id    = data.ovirt_clusters.default.clusters.0.id
}

resource "ovirt_datacenter" "default2" {
  name  = "datacenter2"
  local = false
}

resource "ovirt_cluster" "default2" {
  name                              = "Default2"
  datacenter_id                     = ovirt_datacenter.default2.id
  cpu_arch                          = "x86_64"
  cpu_type                          = "Secure Intel Cascadelake Server Family"
  compatibility_version             = "4.4"
  memory_policy_over_commit_percent = 100
}

resource "ovirt_network" "ovirtmgmt_test" {
  name          = "ovirtmgmt-test"
  datacenter_id = local.default_datacenter_id
}

resource "ovirt_vnic_profile" "mirror" {
  name       = "mirror"
  network_id = ovirt_network.ovirtmgmt_test.id
}

resource "ovirt_vnic_profile" "no_mirror" {
  name       = "no_mirror"
  network_id = ovirt_network.ovirtmgmt_test.id
}

resource "ovirt_host" "host65" {
  name          = "host65"
  root_password = "secret"
  address       = "10.136.6.102"
  cluster_id    = local.default_cluster_id
}

resource "ovirt_storage_domain" "data_local" {
  for_each = local.datastores

  name          = each.key
  host_id       = ovirt_host.host65.id
  datacenter_id = local.default_datacenter_id
  type          = "data"
  localfs {
    path = each.value
  }
}

resource "ovirt_disk" "disks" {
  for_each = local.disks

  name              = each.key
  alias             = each.key
  storage_domain_id = ovirt_storage_domain.data_local["data"].id
  size              = each.value
  format            = "cow"
  sparse            = true
}

data "ovirt_vnic_profiles" "ovirtmgmt" {
  network_id = data.ovirt_networks.ovirtmgmt.networks.0.id
  name_regex = "ovirtmgmt"
}

resource "ovirt_vm" "hostedengine" {
  name       = "HostedEngine"
  cluster_id = local.default_cluster_id
  memory     = 1024

  os {
    type = "other"
  }

  nics {
    name            = "eth0"
    vnic_profile_id = data.ovirt_vnic_profiles.ovirtmgmt.vnic_profiles.0.id
  }

  block_device {
    interface    = "virtio"
    disk_id      = ovirt_disk.disks["hosted_engine"].id
    pass_discard = false
  }
}

resource "ovirt_template" "test_template" {
  name       = "testTemplate"
  cluster_id = local.default_cluster_id
  vm_id      = ovirt_vm.hostedengine.id
  clone      = true
}
