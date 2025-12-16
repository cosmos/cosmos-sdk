# System tests

Go black box tests that setup and interact with a local blockchain. The system test [framework](../../systemtests) 
works with the compiled binary of the chain artifact only.
To get up to speed, checkout the [getting started guide](../../systemtests/GETTING_STARTED.md).

Beside the Go tests and testdata files, this directory can contain the following directories:  

* `binaries` - cache for binary
* `testnet` - node files

Please make sure to not add or push them to git. 

## Execution

Build a new binary from current branch and copy it to the `tests/systemtests/binaries` folder by running system tests.
In project root:

```shell
make test-system
```

Or via manual steps

```shell
make build
mkdir -p ./tests/systemtests/binaries
cp ./build/simd ./tests/systemtests/binaries/
```

### Manual test run

```shell
go test -v -mod=readonly -failfast -tags='system_test' --run TestStakeUnstake    ./... --verbose
```
