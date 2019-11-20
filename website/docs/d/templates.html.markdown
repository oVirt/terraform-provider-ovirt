---
layout: "ovirt"
page_title: "oVirt: ovirt_templates"
sidebar_current: "docs-ovirt-datasource-templates"
description: |-
  Provides details about oVirt templates
---

# Data Source: ovirt\_templates

The oVirt Templates data source allows access to details of list of templates within oVirt.

## Example Usage

```hcl
data "ovirt_templates" "search_filtered_template" {
  search = {
    criteria       = "name = centOST"
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

`templates` is set to the wrapper of the found templates. Each item of `templates` contains the following attributes exported:

* `id` - The ID of oVirt Template
* `name` - The name of oVirt Template
* `cpu_shares` - The cpu shares of VM which associated with oVirt Template
* `memory` - The memory of VM which associated with oVirt Template, in Megabytes(MB)
* `creation_time` - The creation time of VM which associated with oVirt Template
* `status` - The status of oVirt Template
