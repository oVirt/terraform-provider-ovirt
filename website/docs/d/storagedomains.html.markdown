---
layout: "ovirt"
page_title: "oVirt: ovirt_storagedomains"
sidebar_current: "docs-ovirt-datasource-storagedomains"
description: |-
  Provides details about oVirt storagedomains
---

# Data Source: ovirt\_storagedomains

The oVirt Storagedomains data source allows access to details of list of storagedomains within oVirt.

## Example Usage

```hcl
data "ovirt_storagedomains" "filtered_storagedomains" {
  name_regex = "^MAIN_dat.*|^DEV_dat.*"

  search = {
    criteria       = "status != unattached and name = DS_INTERNAL and datacenter = MY_DC"
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

`storagedomains` is set to the wrapper of the found storagedomains. Each item of `storagedomains` contains the following attributes exported:

* `id` - The ID of oVirt Storagedomain
* `name` - The name of oVirt Storagedomain
* `status` - The status of oVirt Storagedomain
* `external_status` - The external status of oVirt Storagedomain
* `type` - The type of oVirt Storagedomain
* `description` - The description of oVirt Storagedomain
* `datacenter_id` - The ID of oVirt Datacenter the Storagedomain belongs to
