---
layout: "ovirt"
page_title: "oVirt: ovirt_nics"
sidebar_current: "docs-ovirt-datasource-nics"
description: |-
  Provides details about oVirt NICs
---

# Data Source: ovirt\_nics

The oVirt NICs data source allows access to details of list of NICs of a VM within oVirt.

## Example Usage

```hcl
data "ovirt_nics" "filtered_nics" {
  name_regex = "^ovirtmgmt-t*"
  vm_id      = "b9ea419c-7ce0-4508-8d04-8e75f60041ea"
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional) The fully functional regular expression for name
* `vm_id` - (Required) The ID of the VM that the NIC belongs to

> This data source dose not support for the regular oVirt query language.

## Attributes Reference

`nics` is set to the wrapper of the found NICs. Each item of `nic` contains the following attributes exported:

* `id` - The ID of oVirt NIC
* `name` - The name of oVirt NIC
* `boot_protocol` - Defines how an IP address is assigned to the NIC
* `comment` - Free text containing comments about the NIC
* `description` - A human-readable description in plain text about the NIC
* `interface` - The type of driver used for the NIC
* `linked` - Defines if the NIC is linked to the VM
* `on_boot` - Defines if the network interface should be activated upon operation system startup
* `plugged` - Defines if the NIC is plugged in to the virtual machine
* `mac_address` - The MAC address of the interface
* `reported_devices` - A collection of reported devices that are associated with the virtual network interface
  * `id` - The ID of the reported device
  * `name` - The name of the reported device
  * `mac_address` - The MAC address of the reported device
  * `description` - A human-readable description in plain text about the reported device
  * `comment` - Free text containing comments about the reported device
  * `type` - The type of the reported device
  * `ips` - A list of IP configurations of the reported device
    * `address` - The text representation of the IP address
    * `gateway` - The address of the default gateway
    * `netmask` - The network mask
    * `version` - The version of the IP protocol
