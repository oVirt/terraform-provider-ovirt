provider "ovirt" {
  username = "admin@internal"
  url      = "https://engine-fqdn/ovirt-engine/api"
  password = "123"
}

resource "ovirt_image_transfer" "cirros_transfer" {
  alias             = "cirros4"
  source_url        = "http://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
  storage_domain_id = "d787bf6b-fae1-4a3e-b773-2ac466599d29"
  sparse            = true
}

resource "ovirt_vm" "tmpvm" {
  name       = "tmpvm-for-${ovirt_image_transfer.cirros_transfer.alias}"
  cluster_id = "5c8f6906-f14b-43ee-83df-5f800f36eb70"
  block_device {
    disk_id   = ovirt_image_transfer.cirros_transfer.disk_id
    interface = "virtio_scsi"
  }
}

resource "ovirt_template" "cirrus_template_1" {
  name       = "template-for-${ovirt_image_transfer.cirros_transfer.alias}"
  cluster_id = ovirt_vm.tmpvm.cluster_id
  // create from vm
  vm_id = ovirt_vm.tmpvm.id
}




