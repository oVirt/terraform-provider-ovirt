---
layout: "ovirt"
page_title: "oVirt: ovirt_vnic_profile"
sidebar_current: "docs-ovirt-resource-vnic-profile"
description: |-
  Manages a vNIC profile resource within oVirt.
---

# ovirt\_vnic\_profile

Manages a vNIC profile resource within oVirt.

## Example Usage

```hcl
resource "ovirt_vnic_profile" "vnicprofile" {
  name           = "myvnicprofile"
  migratable     = false
  network_id     = "43631f2d-2558-4a42-adaa-2e9807144dc8"
  port_mirroring = false
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name of the vNIC profile. Changing this updates the vNIC profile's name.
* `network_id` - (Required) The ID of network the vNIC profile applies to. Changing this creates a new vNIC profile.
* `migratable` - (Optional) A flag to indicate whether `pass_through` vNIC is migratable. Default is `false`. Changing this updates the vNIC profile's migratable.
* `port_mirroring` - (Optional) A flag to indicate whether port mirroring is enabled. Default is `false`. Changing this updates the vNIC profile's port mirroring.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of oVirt vNIC Profile

## Import

vNIC profiles can be imported using the `id`, e.g.

```
$ terraform import ovirt_vnic_profile.vnicprofile fe98758d-60f8-4206-8ffc-f772e906d752
```