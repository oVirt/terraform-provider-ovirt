#!/bin/bash -xe

checkout_ost() {
  git init ovirt-system-tests && cd ovirt-system-tests
  git fetch "https://gerrit.ovirt.org/ovirt-system-tests" +refs/changes/00/112600/38:myhead
  git checkout myhead
  git reset --hard HEAD
  git clean -fdx
}
checkout_ost
./automation/tr_suite_master.sh
