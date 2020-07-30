package simapp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkkv "github.com/cosmos/cosmos-sdk/types/kv"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestGetSimulationLog(t *testing.T) {
	cdc := std.MakeCodec(ModuleBasics)

	decoders := make(sdk.StoreDecoderRegistry)
	decoders[authtypes.StoreKey] = func(kvAs, kvBs sdkkv.Pair) string { return "10" }

	tests := []struct {
		store       string
		kvPairs     []sdkkv.Pair
		expectedLog string
	}{
		{
			"Empty",
			[]sdkkv.Pair{{}},
			"",
		},
		{
			authtypes.StoreKey,
			[]sdkkv.Pair{{Key: authtypes.GlobalAccountNumberKey, Value: cdc.MustMarshalBinaryBare(uint64(10))}},
			"10",
		},
		{
			"OtherStore",
			[]sdkkv.Pair{{Key: []byte("key"), Value: []byte("value")}},
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
