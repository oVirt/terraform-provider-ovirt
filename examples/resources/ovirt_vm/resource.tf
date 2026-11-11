data "ovirt_blank_template" "blank" {
}

resource "ovirt_vm" "test" {
  name        = random_string.vm_name.result
  comment     = "Hello world!"
  cluster_id  = var.cluster_id
  template_id = data.ovirt_blank_template.blank.id

  # Optional cloud-init initialization:
  # initialization_hostname      = "my-vm"
  # initialization_custom_script = yamlencode({
  #   "ssh_authorized_keys" = ["ssh-rsa AAAAB3... user@example.com"]
  #   "runcmd"              = ["echo 'ready' > /tmp/ready.txt"]
  # })
}