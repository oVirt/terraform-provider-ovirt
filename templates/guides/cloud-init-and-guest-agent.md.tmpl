---
page_title: "Using cloud-init and qemu-guest-agent"
subcategory: ""
description: |-
  Requirements for VM initialization and IP address reporting.
---

# Using cloud-init and qemu-guest-agent

The `initialization_*` fields of the `ovirt_vm` resource and the `ovirt_wait_for_ip` data source rely on software
running inside the guest operating system. The provider does not install this software, so it must be present in the
template the VM is created from.

## cloud-init

oVirt passes `initialization_hostname`, `initialization_custom_script` and `initialization_nic` to the VM using
cloud-init. For these settings to take effect, the `cloud-init` package must be installed and enabled in the template.
Without it the VM still starts, but the settings are ignored.

`initialization_custom_script` is cloud-config YAML that oVirt adds to the configuration it generates, so it should not
start with a `#cloud-config` line.

## qemu-guest-agent

oVirt learns the IP addresses of a VM from the guest agent. The `qemu-guest-agent` package must be installed and its
service enabled in the template, otherwise no addresses are reported and `ovirt_wait_for_ip` cannot complete.

## Example Usage

The following example creates a VM from a template that has both packages installed, starts it, and waits for an IP
address. The template is expected to contain a disk and a NIC.

```terraform
resource "ovirt_vm" "example" {
  name        = "example"
  cluster_id  = var.cluster_id
  template_id = var.template_id

  initialization_hostname      = "example"
  initialization_custom_script = <<-EOT
    ssh_authorized_keys:
      - ${var.ssh_public_key}
  EOT
}

resource "ovirt_vm_start" "example" {
  vm_id = ovirt_vm.example.id
}

// Wait for the guest agent to report an IP address.
data "ovirt_wait_for_ip" "example" {
  vm_id = ovirt_vm_start.example.vm_id
}

output "ipv4" {
  value = data.ovirt_wait_for_ip.example.interfaces.*.ipv4_addresses
}
```
