# Getting started with a new system test

## Preparation

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

## Writing the first system test

Switch to the `tests/systemtests` folder to work from here.

If there is no test file matching your use case, start a new test file here.
for example `bank_test.go` to begin with:

```go
//go:build system_test

package systemtests

import (
	"testing"
)

func TestQueryTotalSupply(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	raw := cli.CustomQuery("q", "bank", "total-supply")
	t.Log("### got: " + raw)
}
```
The file begins with a Go build tag to exclude it from regular go test runs.
All tests in the `systemtests` folder build upon the *test runner* initialized in `main_test.go`.
This gives you a multi node chain started on your box.
It is a good practice to reset state in the beginning so that you have a stable base.

The system tests framework comes with a CLI wrapper that makes it easier to interact or parse results.
In this example we want to execute `simd q bank total-supply --output json --node tcp://localhost:26657` which queries the bank module.
Then print the result to for the next steps

### Run the test

```shell
go test -mod=readonly -tags='system_test' -v ./...  --run TestQueryTotalSupply --verbose 
```

This give very verbose output. You would see all simd CLI commands used for starting the server or by the client to interact.
In the example code, we just log the output. Watch out for 
```shell
    bank_test.go:15: ### got: {
          "supply": [
            {
              "denom": "stake",
              "amount": "2000000190"
            },
            {
              "denom": "testtoken",
              "amount": "4000000000"
            }
          ],
          "pagination": {
            "total": "2"
          }
        }
```

At the end is a tail from the server log printed. This can sometimes be handy when debugging issues.


### Tips

* Passing `--nodes-count=1` overwrites the default node count and can speed up your test for local runs