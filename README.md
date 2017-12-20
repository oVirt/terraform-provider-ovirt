Installing Terraform oVirt plugin
=================================
Set GOPATH (usually ~/go)
brew install go
GIT_TERMINAL_PROMPT=1 go get github.com/EMSL-MSC/terraform-provider-ovirt
cd $GOPATH/src/github.com/EMSL-MSC/terraform-provider-ovirt
go build

In Terraform module
===================
mkdir -p terraform.d/plugins/darwin_amd64
cp $GOPATH/src/github.com/EMSL-MSC/terraform-provider-ovirt/terraform-provider-ovirt terraform.d/plugins/darwin_amd64
terraform init
