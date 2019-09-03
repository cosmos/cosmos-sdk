package simulation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	types.RegisterCodec(cdc)
	return
}
func TestDecodeStore(t *testing.T) {
	cdc := makeTestCodec()

	totalSupply := types.NewSupply(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)))

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: keeper.SupplyKey, Value: cdc.MustMarshalBinaryLengthPrefixed(totalSupply)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Supply", fmt.Sprintf("%v\n%v", totalSupply, totalSupply)},
		{"other", ""},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeStore(cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeStore(cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
