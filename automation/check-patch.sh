#!/bin/bash -xe

checkout_ost() {
  git clone "https://gerrit.ovirt.org/ovirt-system-tests"
  cd ovirt-system-tests
}
checkout_ost
./automation/tr_suite_master.sh
