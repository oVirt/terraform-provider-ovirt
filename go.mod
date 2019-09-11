module github.com/ovirt/terraform-provider-ovirt

go 1.12

require (
	github.com/hashicorp/terraform v0.12.2
	github.com/ovirt/go-ovirt v4.3.4+incompatible
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999 // indirect