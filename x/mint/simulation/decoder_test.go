package simulation_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/std"

	"github.com/stretchr/testify/require"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/simulation"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

func TestDecodeStore(t *testing.T) {
	cdc := std.NewAppCodec(std.MakeCodec(simapp.ModuleBasics))
	dec := simulation.NewDecodeStore(cdc)

	minter := types.NewMinter(sdk.OneDec(), sdk.NewDec(15))

	kvPairs := tmkv.Pairs{
		tmkv.Pair{Key: types.MinterKey, Value: cdc.MustMarshalBinaryBare(&minter)},
		tmkv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Minter", fmt.Sprintf("%v\n%v", minter, minter)},
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
