---
layout: "ovirt"
page_title: "oVirt: ovirt_network"
sidebar_current: "docs-ovirt-resource-network"
description: |-
  Manages a Network resource within oVirt.
---

# ovirt\_network

Manages a network resource within oVirt.

## Example Usage

```hcl
resource "ovirt_network" "network" {
  name          = "mynetwork"
  description   = "my new network"
  datacenter_id = "00bfb5f6-1641-4fe5-b634-9f53c36f753b"
  vlan_id       = 488
  mtu           = 0
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the network. Changing this updates the network's name.
* `description` - (Optional) A description of the network. Changing this updates the network's description.
* `datacenter_id` - (Required) The ID of datacenter the network belongs to. Changing this updates the network's datacenter_id.
* `vlan_id` - (Optional) The vlan tag of the network. Changing this updates the network's vlan_id.
* `mtu` - (Optional) A mtu of the network. Changing this updates the network's mtu.

## Attributes Reference

The following attributes are exported:

* `name` - See Argument Reference above
* `description` - See Argument Reference above
* `datacenter_id` - See Argument Reference above
* `vlan_id` - See Argument Reference above
* `mtu` - See Argument Reference above

## Import

Networks can be imported using the `id`, e.g.

```
$ terraform import ovirt_network.network 381e3d4f-dc1e-427d-9e07-9ce72a188304
```