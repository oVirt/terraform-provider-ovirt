terraform {
  required_providers {
    ovirt = {
      source = "oVirt/ovirt"
    }
  }
}

provider "ovirt" {
  # Set this to your oVirt Engine URL, e.g. https://example.com/ovirt-engine/api/
  url = var.url
  # Set this to your oVirt username, e.g. admin@internal
  username = var.username
  # Set this to your oVirt password.
  password = var.password
  # Take trusted certificates from the specified files (list).
  tls_ca_files = var.tls_ca_files
  # Take trusted certificates from the specified directories (list).
  tls_ca_dirs = var.tls_ca_dirs
  # Take the trusted certificates from the provided variable. Certificates must be in PEM format.
  tls_ca_bundle = var.tls_ca_bundle
  # Set this to true to use the system certificate storage to verify the engine certificate. You must
  # add the certificate to your trusted roots before running. This option doesn't work on Windows.
  tls_system = var.tls_system
  # Set this to true to disable certificate verification. This is a terrible idea.
  tls_insecure = var.tls_insecure
  # Set to true if you want to run an in-memory test. In this mode all other options will be ignored.
  mock = var.mock
  # Set extra headers to add to each request.
  extra_headers = {
    "X-Custom-Header" = "Hello world!"
  }
}
