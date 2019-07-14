---
layout: "ovirt"
page_title: "oVirt: ovirt_networks"
sidebar_current: "docs-ovirt-datasource-networks"
description: |-
  Provides details about oVirt networks
---

# Data Source: ovirt\_networks

The oVirt Networks data source allows access to details of list of networks within oVirt.

## Example Usage

```hcl
data "ovirt_networks" "filtered_networks" {
  name_regex = "^ovirtmgmt-t*"

  search = {
    criteria       = "datacenter = Default and name = ovirtmgmt-test"
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

`networks` is set to the wrapper of the found networks. Each item of `networks` contains the following attributes exported:

* `id` - The ID of oVirt Network
* `name` - The name of oVirt Network
* `datacenter_id` - The ID of oVirt Datacenter the Network belongs to
* `description` - The description of oVirt Network
* `vlan_id` - The vlan tag
* `mtu` - The mtu of oVirt Network
