#!/bin/bash -xe

checkout_ost() {
  git clone "https://gerrit.ovirt.org/ovirt-system-tests"
  cd ovirt-system-tests
}
checkout_ost
./ost.sh run tr-suite-master el8stream 
