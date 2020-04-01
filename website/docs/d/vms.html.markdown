---
layout: "ovirt"
page_title: "oVirt: ovirt_vms"
sidebar_current: "docs-ovirt-datasource-vms"
description: |-
  Provides details about oVirt VMs
---

# Data Source: ovirt\_vms

The oVirt VMs data source allows access to details of list of VMs within oVirt.

## Example Usage

```hcl
data "ovirt_vms" "filtered_vms" {
  name_regex = "^HostedEngine*"

  search = {
    criteria       = "name = HostedEngine and status = up"
    max            = 2
    case_sensitive = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional) The fully functional regular expression for name
* `search` - (Optional) The general search criteria representation, fitting the rules of [Searching](http://ovirt.github.io/ovirt-engine-api-model/master/#_searching)
    * criteria - (Optional) The criteria for searching, using the same syntax as the oVirt query language
    * max - (Optional) The maximum amount of objects returned. If not specified, the search will return all the objects.
    * case_sensitive - (Optional) If the search are case sensitive, default value is `false`

> The `search.criteria` also supports asterisk for searching by name, to indicate that any string matches, including the empty string. For example, the criteria `search=name=myobj*` will return all the objects with names beginning with `myobj`, such as `myobj2`, `myobj-test`. So, you could use `name_regex` for searching by complicated regular expression, and `search.criteria` for simple case accordingly.

## Attributes Reference

`vms` is set to the wrapper of the found VMs. Each item of `vms` contains the following attributes exported:

* `id` - The ID of oVirt VM
* `name` - The name of oVirt VM
* `cluster_id` - The ID of oVirt Cluster the VM belongs to
* `status` - The current status of the VM
* `template_id` - The ID of oVirt Template the VM creates from
* `instance_type_id` - The ID of the Instance Type
* `high_availability` - Defines if the HA is enabled
* `memory` - The VM's memory, in Megabytes(MB)
* `cores` - The CPU cores of the VM
* `sockets` - The CPU sockets of the VM
* `threads` - The CPU threads of the VM
* `reported_devices` - A collection of reported devices that are associated with the network interface of the VM
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
