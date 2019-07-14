---
layout: "ovirt"
page_title: "oVirt: ovirt_clusters"
sidebar_current: "docs-ovirt-datasource-clusters"
description: |-
  Provides details about oVirt clusters
---

# Data Source: ovirt\_clusters

The oVirt Clusters data source allows access to details of list of clusters within oVirt.

## Example Usage

```hcl
data "ovirt_clusters" "filtered_clusters" {
  name_regex = "^default*"

  search = {
    criteria       = "architecture = x86_64 and Storage.name = data"
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

`clusters` is set to the wrapper of the found clusters. Each item of `clusters` contains the following attributes exported:

* `id` - The ID of oVirt Cluster
* `name` - The name of oVirt Cluster
* `description` - The description of oVirt Cluster
* `datacenter_id` - The ID of oVirt Datacenter the cluster belongs to