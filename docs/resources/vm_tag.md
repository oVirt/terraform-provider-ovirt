---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ovirt_vm_tag Resource - terraform-provider-ovirt"
subcategory: ""
description: |-
  The ovirt_vm_tag resource attaches a tag to a virtual machine.
---

# ovirt_vm_tag (Resource)

The ovirt_vm_tag resource attaches a tag to a virtual machine.

## Example Usage

```terraform
data "ovirt_blank_template" "blank" {
}

resource "ovirt_tag" "test" {
  name = random_string.tag_name.result
}

resource "ovirt_vm" "test" {
  name        = random_string.vm_name.result
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id
}

resource "ovirt_vm_tag" "test" {
  tag_id = ovirt_tag.test.id
  vm_id  = ovirt_vm.test.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `tag_id` (String) ID for the tag to be attached
- `vm_id` (String) ID for the VM to be attached

### Read-Only

- `id` (String) The ID of this resource.
