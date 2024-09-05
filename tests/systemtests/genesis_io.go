package systemtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetConsensusMaxGas max gas that can be consumed in a block
func SetConsensusMaxGas(t *testing.T, max int) GenesisMutator {
	t.Helper()
	return func(genesis []byte) []byte {
		state, err := sjson.SetRawBytes(genesis, "consensus.params.block.max_gas", []byte(fmt.Sprintf(`"%d"`, max)))
		require.NoError(t, err)
		return state
	}
}

func SetGovVotingPeriod(t *testing.T, period time.Duration) GenesisMutator {
	t.Helper()
	return func(genesis []byte) []byte {
		state, err := sjson.SetRawBytes(genesis, "app_state.gov.params.voting_period", []byte(fmt.Sprintf("%q", period.String())))
		require.NoError(t, err)
		return state
	}
}

// GetGenesisBalance return the balance amount for an address from the given genesis json
func GetGenesisBalance(rawGenesis []byte, addr string) sdk.Coins {
	var r []sdk.Coin
	balances := gjson.GetBytes(rawGenesis, fmt.Sprintf(`app_state.bank.balances.#[address==%q]#.coins`, addr)).Array()
	for _, coins := range balances {
		for _, coin := range coins.Array() {
			r = append(r, sdk.NewCoin(coin.Get("denom").String(), sdkmath.NewInt(coin.Get("amount").Int())))
		}
	}
	return r
}
