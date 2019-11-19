package simapp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/auth"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSimulationLog(t *testing.T) {
	cdc := MakeCodec()

	decoders := make(sdk.StoreDecoderRegistry)
	decoders[auth.StoreKey] = func(cdc *codec.Codec, kvAs, kvBs cmn.KVPair) string { return "10" }

	tests := []struct {
		store       string
		kvPairs     []cmn.KVPair
		expectedLog string
	}{
		{
			"Empty",
			[]cmn.KVPair{{}},
			"",
		},
		{
			auth.StoreKey,
			[]cmn.KVPair{{Key: auth.GlobalAccountNumberKey, Value: cdc.MustMarshalBinaryLengthPrefixed(uint64(10))}},
			"10",
		},
		{
			"OtherStore",
			[]cmn.KVPair{{Key: []byte("key"), Value: []byte("value")}},
			fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", []byte("key"), []byte("value"), []byte("key"), []byte("value")),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.store, func(t *testing.T) {
			require.Equal(t, tt.expectedLog, GetSimulationLog(tt.store, decoders, cdc, tt.kvPairs, tt.kvPairs), tt.store)
		})
	}
}
