---
layout: "ovirt"
page_title: "oVirt: ovirt_disk_attachment"
sidebar_current: "docs-ovirt-resource-disk-attachment"
description: |-
  Manages a Disk attachment resource within oVirt.
---

# ovirt\_disk\_attachment

Manages a Disk attachment resource within oVirt.

## Example Usage

```hcl
resource "ovirt_disk_attachment" "diskattachment" {
  vm_id                = "5ba458c1-01fd-00eb-0140-000000000351"
  disk_id              = "67f88160-396b-441b-8824-f2c22e80bf82"
  active               = true
  bootable             = true
  interface            = "virtio"
  pass_discard         = true
  read_only            = true
  use_scsi_reservation = false
}
```

## Argument Reference

The following arguments are supported:

* `vm_id` - (Required) The ID of VM the disk attached to. Changing this creates a new disk attachment.
* `disk_id` - (Required) The ID of attached disk. Changing this creates a new disk attachment.
* `active` - (Optional) The flag to indicate whether the disk is active. Default is `true`. Changing this updates the attachment's active.
* `bootable` - (Optional) The flag to indicate whether the disk is bootable. Default is `false`. Changing this updates the attachment's bootable.
* `interface` - (Required) The interface of the attachment. Valid values are `ide`, `sata`, `spapr_vscsi`, `virtio` and `virtio_scsi`. Changing this creates a new attachment.
* `pass_discard` - (Optional) The flag to indicate whether the VM passes discard commands to the storage. Changing this creates a new attachment.
* `read_only` - (Optional) The flag to indicate whether the disk is connected to the VM as read only. Default is `false`. Changing this creates a new attachment.
* `use_scsi_reservation` - (Optional) The flag to indicate whether SCSI reservation is enabled for this disk. Default is `false`. Changing this creates a new attachment.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The composite ID of oVirt Disk Attachment which is constituted by the ID of the vm and the ID of the disk within oVirt.

## Import

Disk attachment can be imported using the composite `id`, e.g.

```
$ terraform import ovirt_disk_attachment.diskattachment 3d88d40c-3230-4266-9228-fff5c1348081:c76d73db-2e81-49a5-a2d2-7065650680e5
```