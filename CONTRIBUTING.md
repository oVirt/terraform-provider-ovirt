# Contribution guide

Hey there, and thank you for helping out! This guide will help you set up your development environment, and submit your pull request to the oVirt Terraform Provider.

## Setting up a development environment

Terraform providers are written in the [Go programming language](https://golang.org/), so make sure you set up a local Go SDK. At the time of writing you will need Go version `1.14`, but make sure to check the [go.mod](go.mod) file for the current version.

Other than that, you will need a working oVirt engine with at least one host and a working storage domain. The test suite requires roughly 2 GB of disk space.

## Developing your patch

When developing your patch you will need to keep a few things in mind. In no particular order, these are the following:

### One change = one PR

Please make sure that your patch contains as little as feasible. Having a smaller change makes it easier to test and review.

Formatting changes, adding or removing empty lines should be avoided under all circumstances because they unnecessarily bloat the patch size. If you want to submit a formatting change please do so in a separate PR.

### Use and write tests

Tests are an integral part of writing a Terraform provider as manual testing every change is nearly impossible.

You can run the oVirt provider tests using the following command:

```bash
OVIRT_URL=https://your-ovirt-engine-host/ovirt-engine/api/;OVIRT_USERNAME=admin@internal;OVIRT_PASSWORD=your-ovirt-password;TF_ACC=1;OVIRT_INSECURE=1 go test -v ./...
```

This will run the tests against your oVirt setup. The tests are written in such a way that they should do their best to clean up after themselves.

As you may think, we would like to keep our test coverage up. That means that any non-trivial patch will require a test to be written. Please submit your tests along with your code changes.

If your test involves a specific setup (e.g. only applies to NFS storage domains, etc) then make sure your test fails gracefully on other oVirt setups. Also make sure to note testing requirements in your PR text.

### Code quality

The provider is currently unfortunately not in the best shape. Especially the VM creation is extremely long and not well tested. Please make sure that your patch makes the situation better. If you can, try to separate out your change into *separate functions*.

Tools that can help you with code quality:

- `go vet`
- `go fmt`
- [`golangci-lint`](https://golangci-lint.run/)

### Terraform correctness

Terraform is a tricky beast. The following tips should help you keep on track:

1. Apply a test config twice. The second apply should yield no change. If it does, there is something wrong.
2. Does your change affect other resources?
3. Could it be useful to split out your change into a separate resource? If yes please do so.
4. Does your change handle manual changes in the oVirt engine or does your code crash if the resource was manually modified?

## Submitting your PR

When your patch is done simply head on over to GitHub and [submit your PR](https://github.com/oVirt/terraform-provider-ovirt/pulls). We will do our best to review your PR within a few weeks.
