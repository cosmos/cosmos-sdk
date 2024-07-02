//go:build system_test

package systemtests

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tidwall/sjson"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestQueryTotalSupply(t *testing.T) {
	// scenario:
	// given a chain with a custom token on genesis
	// when an amount is burned
	// then this is reflected in the total supply

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	sut.ModifyGenesisJSON(t, setupGenesis(t, cli))
	sut.StartChain(t)

	raw := cli.CustomQuery("q", "bank", "total-supply")

	exp := map[string]int64{
		"stake":     int64(500000000 * sut.nodesCount),
		"testtoken": int64(1000000000 * sut.nodesCount),
		"mytoken":   1_000_000,
	}
	require.Len(t, gjson.Get(raw, "supply").Array(), len(exp), raw)

	for k, v := range exp {
		got := gjson.Get(raw, fmt.Sprintf("supply.#(denom==%q).amount", k)).Int()
		assert.Equal(t, v, got, raw)
	}

	// and when
	txHash := cli.Run("tx", "bank", "burn", "node0", "400000mytoken")
	RequireTxSuccess(t, txHash)
	// then
	exp["mytoken"] = 600_000 // update expected state
	raw = cli.CustomQuery("q", "bank", "total-supply")
	for k, v := range exp {
		got := gjson.Get(raw, fmt.Sprintf("supply.#(denom==%q).amount", k)).Int()
		assert.Equal(t, v, got, raw)
	}
	assert.Equal(t, int64(600_000), cli.QueryBalance(cli.GetKeyAddr("node0"), "mytoken"))
}

func setupGenesis(t *testing.T, cli *CLIWrapper) func(genesis []byte) []byte {
	return func(genesis []byte) []byte {
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
	}
}
