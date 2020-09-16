Testing
=======

The acceptance tests must be run against a live oVirt environment. The test requires environment variables to be set to point it at this.

Tests are then run like:
```
 OVIRT_URL=https://myengine/ovirt-engine/api OVIRT_PASSWORD=admin OVIRT_USERNAME=admin@internal TF_ACC=1 make testacc
 # Specific tests
 OVIRT_URL=https://myengine/ovirt-engine/api OVIRT_PASSWORD=admin OVIRT_USERNAME=admin@internal TF_ACC=1 go test ./ovirt -run=testAccOvirtHost_ -v
```


Where possible, the tests will create & destroy resources to test against, but they do currently still require some pre-existing
 infrastructure. With time, this list should be reduced. 

List of required infrastructure:
--------------------------------
examples/testing/main.tf can be used as a guide to create most of this infrastructure (with the exception of detached_vm). ]
This is tested against a single engine/host combination, with a local storage datacenter.   

* The default datacentre, cluster and network created by the engine installer
* A second datacenter "datacenter2", and a second cluster "Default2" 

* Datastores called:
  * data
  * DEV_datastore
  * MAIN_datastore
  * DS_INTERNAL
  
* Disks called:
  * test_disk1
  * test_disk2
  
* Networks called:
  * ovirtmgmt-test
  
* VNic profiles called:
  * mirror
  * no_mirror
  
* VMs called:
  * "HostedEngine". This must be a running VM, but can be a powered on blank disk
  * "detached_vm". This must be a VM without an attached disk (can be powered down)
  
* Template called:
  * "testTemplate". Must contain a disk (that can be blank)
  
* Hosts called:
  * "host65", with root password "secret"
  
Issues
------
Some tests will still fail even with all the above set up, for example:
* TestAccOvirtHost_basic - Adding a host fails (need another host, and hardcoded IP)
* TestAccOvirtNicsDataSource_nameRegexFilter - Checking the address on HostedEngine will fail (unless it is further configured)
* TestAccOvirtStorageDomain_nfs - Attempts to mount a hardcoded external NFS server
