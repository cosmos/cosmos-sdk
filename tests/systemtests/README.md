# System tests

Go black box tests that setup and interact with a local blockchain. The system test [framework](../../systemtests) 
works with the compiled binary of the chain artifact only.
To get up to speed, checkout the [getting started guide](../../systemtests/getting_started.md).

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

### Working with macOS

Most tests should function seamlessly. However, the file [upgrade_test.go](upgrade_test.go) includes a **build annotation** for Linux only.

For the system upgrade test, an older version of the binary is utilized to perform a chain upgrade. This artifact is retrieved from a Docker container built for Linux.

To circumvent this limitation locally:
1. Checkout and build the older version of the artifact from a specific tag for your OS.
2. Place the built artifact into the `binaries` folder.
3. Ensure that the filename, including the version, is correct.

With the cached artifact in place, the test will use this file instead of attempting to pull it from Docker.