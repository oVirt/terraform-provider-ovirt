---
layout: "ovirt"
page_title: "oVirt: ovirt_vnic_profiles"
sidebar_current: "docs-ovirt-datasource-vnic-profiles"
description: |-
  Provides details about oVirt vnic profiles
---

# Data Source: ovirt\_vnic\_profiles

The oVirt vNIC profiles data source allows access to details of list of vNIC profiles within oVirt.

## Example Usage

```hcl
data "ovirt_vnic_profiles" "filtered_vnic_profiles" {
  name_regex = ".*mirror$"
  network_id = "649f2d61-7f23-477b-93bd-d55f974d8bc8"
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional) The fully functional regular expression for name
* `network_id` - (Required) The ID of network the vnic profile belongs to

> This data source dose not support for the regular oVirt query language.

## Attributes Reference

`vnic_profiles` is set to the wrapper of the found vnic profiles. Each item of `vnic_profiles` contains the following attributes exported:

* `id` - The ID of oVirt vNIC profile
* `name` - The name of oVirt vNIC profile
* `network_id` - The ID of network the vNIC profile applies to
* `migratable` - Whether `pass_through` vNIC is migratable
* `port_mirroring` - Whether port mirroring is enabled
