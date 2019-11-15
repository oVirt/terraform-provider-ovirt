---
layout: "ovirt"
page_title: "oVirt: ovirt_snapshot"
sidebar_current: "docs-ovirt-resource-snapshot"
description: |-
  Manages a Snapshot of vm resource within oVirt.
---

# ovirt\_snapshot

Manages a Snapshot of VM resource within oVirt.

## Example Usage

```hcl
resource "ovirt_snapshot" "snapshot" {
  description = "description-of-snasphot"
  vm_id       = "53000b15-82ad-4ed4-9f86-bffb95e3c28b"
  save_memory = true
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Required) A description of the snapshot. Changing this creates a new snapshot.
* `vm_id` - (Required) The ID of vm the snapshot taken from. Changing this creates a new snapshot.
* `save_memory` - (Optional) The flag to indicate whether the content of the memory of the vm is included in the snapshot. Default is `true`. Changing this creates a new snapshot.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The composite ID of oVirt Snapshot which is constituted by the ID of the vm the snapshot taken from and the ID of the snapshot within oVirt.
* `status` - The status of oVirt Snapshot. Can be "in_preview", "locked" or "ok".
* `type` - The type of oVirt Snapshot. Can be "active", "preview", "regular" or "stateless".
* `date` - The string representation of the creation time of oVirt Snapshot in RFC3339 format.

## Import

Snapshots can be imported using the composite `id`, e.g.

```
$ terraform import ovirt_snapshot.snapshot 53000b15-82ad-4ed4-9f86-bffb95e3c28b:df736600-b8be-4029-be98-4b0611be6be4
```