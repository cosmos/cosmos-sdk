# Group Module System Tests

Black-box system tests for the group module. Uses the [testutil/systemtests](../../../testutil/systemtests) framework to run a local multi-node blockchain and exercise the group module via CLI.

## Execution

From `enterprise/group`:

```shell
make test-system
```

Or manually:

```shell
make build
mkdir -p ./tests/systemtests/binaries
cp ./build/simd ./tests/systemtests/binaries/
make -C tests/systemtests test
```

## Manual test run

```shell
cd tests/systemtests
go test -v -mod=readonly -failfast -tags='system_test' ./... --verbose
```

Run a single test:

```shell
go test -v -mod=readonly -failfast -tags='system_test' ./... --run TestGroupQueries --verbose
```
