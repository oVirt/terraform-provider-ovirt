#!/usr/bin/env bash
# check if .tf files need reformatting
# this file should be executed from the root directory of the repo

# check if terraform exists and is a valid version
TF_VERSION="${TF_VERSION:-`cat .terraform-version`}"
set +e
TF_VERSION_INSTALLED=$(bin/terraform --version | head -n1 | awk -Fv '{ print $2}')
set -e

if ! [ "${TF_VERSION}" == "${TF_VERSION_INSTALLED}" ]; then
	scripts/terraform_install.sh
fi

# Check terraform fmt
echo "==> Checking that code passes terraform fmt requirements..."
tffmt_files=$(bin/terraform fmt -write=false -check=true )
if [[ -n ${tffmt_files} ]]; then
    echo 'terraform fmt needs running on the following files:'
    echo "${tffmt_files}"
    echo "You can use the command: \`make tf-fmt\` to reformat code."
    exit 1
fi

exit 0
