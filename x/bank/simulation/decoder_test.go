package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestDecodeStore(t *testing.T) {
	app := simapp.Setup(false)
	dec := simulation.NewDecodeStore(app.BankKeeper)

	totalSupply := types.NewSupply(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)))

	supplyBz, err := app.BankKeeper.MarshalSupply(totalSupply)
	require.NoError(t, err)

	kvPairs := tmkv.Pairs{
		tmkv.Pair{Key: types.SupplyKey, Value: supplyBz},
		tmkv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Supply", fmt.Sprintf("%v\n%v", totalSupply, totalSupply)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
