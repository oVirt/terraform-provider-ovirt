module github.com/ovirt/terraform-provider-ovirt

go 1.12

require (
	github.com/hashicorp/terraform v0.12.2
	github.com/ovirt/go-ovirt v0.0.0-20200320082526-4e97a11ff083
	github.com/stretchr/testify v1.3.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999 // indirect
