#!/bin/bash -xe

checkout_ost() {
  git clone "https://gerrit.ovirt.org/ovirt-system-tests"
  cd ovirt-system-tests
}
checkout_ost
./setup_for_ost.sh -y 
./ost.sh run tr-suite-master el8stream 
