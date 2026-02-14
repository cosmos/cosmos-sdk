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

## Part 1: Writing the first system test

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
In this example we want to execute `simd q bank total-supply --output json --node tcp://localhost:26657` which queries
the bank module.
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

## Part 2: Working with json

When we have a json response, the [gjson](https://github.com/tidwall/gjson) lib can shine. It comes with jquery like
syntax that makes it easy to navigation within the document.

For example `gjson.Get(raw, "supply").Array()` gives us all the children to `supply` as an array.
Or `gjson.Get("supply.#(denom==stake).amount").Int()` for the amount of the stake token as int64 type.

In order to test our assumptions in the system test, we modify the code to use `gjson` to fetch the data:

```go
	raw := cli.CustomQuery("q", "bank", "total-supply")

	exp := map[string]int64{
        "stake":     int64(500000000 * sut.nodesCount),
        "testtoken": int64(1000000000 * sut.nodesCount),
	}
	require.Len(t, gjson.Get(raw, "supply").Array(), len(exp), raw)

	for k, v := range exp {
		got := gjson.Get(raw, fmt.Sprintf("supply.#(denom==%q).amount", k)).Int()
		assert.Equal(t, v, got, raw)
	}
```

The assumption on the staking token usually fails due to inflation minted on the staking token. Let's fix this in the next step 

### Run the test

```shell
go test -mod=readonly -tags='system_test' -v ./...  --run TestQueryTotalSupply --verbose 
```

### Tips

* Putting the `raw` json response to the assert/require statements helps with debugging on failures. You are usually lacking
  context when you look at the values only.


## Part 3: Setting state via genesis

First step is to disable inflation. This can be done via the `ModifyGenesisJSON` helper. But to add some complexity, 
we also introduce a new token and update the balance of the account for key `node0`.
The setup code looks quite big and unreadable now. Usually a good time to think about extracting helper functions for
common operations. The `genesis_io.go` file contains some examples already. I would skip this and take this to showcase the mix
of `gjson`, `sjson` and stdlib json operations.

```go
	sut.ResetChain(t)
    cli := NewCLIWrapper(t, sut, verbose)

	sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		// disable inflation
		genesis, err := sjson.SetRawBytes(genesis, "app_state.mint.minter.inflation", []byte(`"0.000000000000000000"`))
		require.NoError(t, err)

		// add new token to supply
		var supply []json.RawMessage
		rawSupply := gjson.Get(string(genesis), "app_state.bank.supply").String()
		require.NoError(t, json.Unmarshal([]byte(rawSupply), &supply))
		supply = append(supply, json.RawMessage(`{"denom": "mytoken","amount": "1000000"}`))
		newSupply, err := json.Marshal(supply)
		require.NoError(t, err)
		genesis, err = sjson.SetRawBytes(genesis, "app_state.bank.supply", newSupply)
		require.NoError(t, err)

		// add amount to any balance
		anyAddr := cli.GetKeyAddr("node0")
		newBalances := GetGenesisBalance(genesis, anyAddr).Add(sdk.NewInt64Coin("mytoken", 1000000))
		newBalancesBz, err := newBalances.MarshalJSON()
		require.NoError(t, err)
		newState, err := sjson.SetRawBytes(genesis, fmt.Sprintf("app_state.bank.balances.#[address==%q]#.coins", anyAddr), newBalancesBz)
		require.NoError(t, err)
		return newState
	})
    sut.StartChain(t)
```

Next step is to add the new token to the assert map. But we can also make it more resilient to different node counts.

```go
	exp := map[string]int64{
		"stake":     int64(500000000 * sut.nodesCount),
		"testtoken": int64(1000000000 * sut.nodesCount),
		"mytoken":   1000000,
	}
```

```shell
go test -mod=readonly -tags='system_test' -v ./...  --run TestQueryTotalSupply --verbose --nodes-count=1 
```

## Part 4: Set state via TX

Complexer workflows and tests require modifying state on a running chain. This works only with builtin logic and operations.
If we want to burn some of our new tokens, we need to submit a bank burn message to do this.
The CLI wrapper works similar to the query. Just pass the parameters. It uses the `node0` key as *default*:

```go
	// and when
	txHash := cli.Run("tx", "bank", "burn", "node0", "400000mytoken")
	RequireTxSuccess(t, txHash)
```

`RequireTxSuccess` or `RequireTxFailure` can be used to ensure the expected result of the operation.
Next, check that the changes are applied.

```go
	exp["mytoken"] = 600_000 // update expected state
	raw = cli.CustomQuery("q", "bank", "total-supply")
	for k, v := range exp {
		got := gjson.Get(raw, fmt.Sprintf("supply.#(denom==%q).amount", k)).Int()
		assert.Equal(t, v, got, raw)
	}
	assert.Equal(t, int64(600_000), cli.QueryBalance(cli.GetKeyAddr("node0"), "mytoken"))
```

While tests are still more or less readable, it can gets harder the longer they are. I found it helpful to add
some comments at the beginning to describe what the intention is. For example:

```go
	// scenario:
	// given a chain with a custom token on genesis
	// when an amount is burned
	// then this is reflected in the total supply
```
