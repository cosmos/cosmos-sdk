package simapp

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/std"

	"github.com/stretchr/testify/require"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestGetSimulationLog(t *testing.T) {
	cdc := std.MakeCodec(ModuleBasics)

	decoders := make(sdk.StoreDecoderRegistry)
	decoders[authtypes.StoreKey] = func(kvAs, kvBs tmkv.Pair) string { return "10" }

	tests := []struct {
		store       string
		kvPairs     []tmkv.Pair
		expectedLog string
	}{
		{
			"Empty",
			[]tmkv.Pair{{}},
			"",
		},
		{
			authtypes.StoreKey,
			[]tmkv.Pair{{Key: authtypes.GlobalAccountNumberKey, Value: cdc.MustMarshalBinaryBare(uint64(10))}},
			"10",
		},
		{
			"OtherStore",
			[]tmkv.Pair{{Key: []byte("key"), Value: []byte("value")}},
			fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", []byte("key"), []byte("value"), []byte("key"), []byte("value")),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.store, func(t *testing.T) {
			require.Equal(t, tt.expectedLog, GetSimulationLog(tt.store, decoders, tt.kvPairs, tt.kvPairs), tt.store)
		})
	}
}
