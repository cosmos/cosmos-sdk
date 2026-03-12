# System tests

Go black box tests that setup and interact with a local blockchain. The system test [framework](../../testutil/systemtests) 
works with the compiled binary of the chain artifact only.
To get up to speed, checkout the [getting started guide](../../testutil/systemtests/GETTING_STARTED.md).

Besides the Go tests and testdata files, this directory can contain the following directories: 

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

From project root, build first (see above), then:

```shell
cd tests/systemtests
go test -v -mod=readonly -failfast -tags='system_test' -run TestStakeUnstake ./... --verbose
```

### Running specific tests

Run from `tests/systemtests` (this package has its own go.mod):

| Test | Command | Notes |
|------|---------|------|
| TestStakeUnstake | `go test -mod=readonly -tags='system_test' -v -run TestStakeUnstake ./... --verbose` | |
| TestHeavyLoadLight | `go test -mod=readonly -tags='system_test' -v -short=false -run TestHeavyLoadLight ./... --nodes-count=4 --verbose` | Requires `-short=false` (skipped in short mode) |
| TestHeavyLoad | `COSMOS_RUN_HEAVY_LOAD_TEST=1 go test -mod=readonly -tags='system_test' -v -run TestHeavyLoad ./... -timeout=15m --nodes-count=4 --verbose` | Gated by env var; use `-timeout=15m` or longer |

CLI flags: `-verbose`, `-nodes-count` (default 4), `-wait-time`, `-block-time`
