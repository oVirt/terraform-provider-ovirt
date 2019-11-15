---
layout: "ovirt"
page_title: "oVirt: ovirt_datacenter"
sidebar_current: "docs-ovirt-resource-datacenter"
description: |-
  Manages a Datacenter resource within oVirt.
---

# ovirt\_datacenter

Manages a Datacenter resource within oVirt.

## Example Usage

```hcl
resource "ovirt_datacenter" "dc" {
  name        = "mydc"
  description = "my new dc"
  local       = false
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the datacenter. Changing this updates the datacenter's name.
* `description` - (Optional) A description of the datacenter. Changing this updates the datacenter's description.
* `local` - (Required) A flag to indicate if the datacener uses local storage. Changing this update the datacenter's local flag.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of oVirt Datacenter
* `status` - The current status of oVirt Datacenter, possible values are `contend`, `maintenance`, `not_operational`, `problematic`, `uninitialized` and `up`.

## Import

Datacenters can be imported using the `id`, e.g.

```
$ terraform import ovirt_datacenter.dc 43631f2d-2558-4a42-adaa-2e9807144dc8
```