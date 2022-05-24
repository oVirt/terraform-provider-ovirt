terraform {
  required_providers {
    ovirt = {
      source  = "ovirt/ovirt"
    }
  }

  required_version = ">= 0.15"
}

provider "ovirt" {
  url           = var.url
  username      = var.username
  password      = var.password
  tls_ca_bundle = var.tls_ca_bundle
  tls_system    = var.tls_system
  tls_ca_dirs   = var.tls_ca_dirs
  tls_ca_files  = var.tls_ca_files
  tls_insecure  = var.tls_insecure
}

