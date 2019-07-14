---
layout: "ovirt"
page_title: "oVirt: ovirt_disks"
sidebar_current: "docs-ovirt-datasource-disks"
description: |-
  Provides details about oVirt disks
---

# Data Source: ovirt\_disks

The oVirt Disks data source allows access to details of list of disks within oVirt.

## Example Usage

```hcl
data "ovirt_disks" "filtered_disks" {
  name_regex = "^test_disk*"

  search = {
    criteria       = "name = test_disk1 and provisioned_size > 1024000000"
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

`disks` is set to the wrapper of the found disks. Each item of `disks` contains the following attributes exported:

* `id` - The ID of oVirt Disk
* `name` - The name of oVirt Disk
* `alias` - The alias of oVirt Disk
* `format` - The format of oVirt Disk
* `quota_id` - The ID of quota of oVirt Disk
* `storage_domain_id` - The ID of storage domain the Disk belongs to
* `size` - The provisioned size of oVirt Disk
* `sharable` - Whether oVirt Disk could be attached to multiple vms
* `sparse` - Whether the physical storage for oVirt Disk should not be preallocated