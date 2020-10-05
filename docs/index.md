# oVirt Provider
The oVirt provider is used to interact with the many resources supported by oVirt. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Configure the oVirt Provider
provider "ovirt" {
  url      = "https://engine-api/ovirt-engine/api"
  username = "admin@internal"
  password = "thepassword"
  headers {
    filter      = true
    all_content = true
  }
}

# Create a VM
resource "ovirt_vm" "test-vm" {
  # ...
}
```

## Configuration Reference

The following arguments are supported:

* `url` - (Required) The oVirt engine API URL. If omitted, the `OVIRT_URL` environment variable is used.
* `username` - (Required) The username for accessing oVirt engine API. If omitted, the `OVIRT_USERNAME` environment variable is used.
* `password` - (Required) The password of the user for accessing oVirt engine API. If omitted, the `OVIRT_PASSWORD` environment variable is used.
* `headers` - (Optional) A bunch of key-value pairs as the headers of the HTTP connection. The headers will be sent along with each API request, and could be overwrote with the header values specified in individual request.
