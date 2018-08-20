#!/bin/bash
# this file should be executed from the root directory of the repo
#
# install specific terraform version to bin/ directory, based on the TF_VERSION env var or
# if TF_VERSION is not set then get it from .terraform-version filecontents
# very basic, nothing fancy

TF_VERSION="${TF_VERSION:-`cat .terraform-version`}"
echo TF_VERSION=${TF_VERSION}

wget https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_linux_amd64.zip -O /tmp/terraform.zip
unzip -o -d ./bin/ /tmp/terraform.zip
chmod +x ./bin/terraform
./bin/terraform --version
