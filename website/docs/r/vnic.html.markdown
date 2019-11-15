---
layout: "ovirt"
page_title: "oVirt: ovirt_vnic"
sidebar_current: "docs-ovirt-resource-vnic"
description: |-
  Manages a vNIC resource within oVirt.
---

# ovirt\_vnic

Manages a vNIC resource within oVirt.

## Example Usage

```hcl
resource "ovirt_vnic" "vnic" {
  name            = "myvnic"
  vm_id           = "fd0dc842-57d4-4ae4-82ea-3a16516fbef7"
  vnic_profile_id = "3e7644a7-1f54-49b9-87e0-d46819fba4c5"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name of the vNIC. Changing this creates a new vNIC.
* `vm_id` - (Required) The ID of vm the vNIC attached to. Changing this creates a new vNIC.
* `vnic_profile_id` - (Required) The ID of the vNIC profile applied to the vNIC. Changing this create a new vNIC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of oVirt vNIC

## Import

vNIC can be imported using the `id`, e.g.

```
$ terraform import ovirt_vnic.vnic 43631f2d-2558-4a42-adaa-2e9807144dc8
```