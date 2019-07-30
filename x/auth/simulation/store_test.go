package simulation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/auth"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	delPk1    = ed25519.GenPrivKey().PubKey()
	delAddr1  = sdk.AccAddress(delPk1.Address())
	valAddr1  = sdk.ValAddress(delPk1.Address())
	consAddr1 = sdk.ConsAddress(delPk1.Address().Bytes())
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	auth.RegisterCodec(cdc)
	return
}

func TestDecodeStore(t *testing.T) {
	cdc := makeTestCodec()
	acc := auth.NewBaseAccountWithAddress(delAddr1)
	globalAccNumber := uint64(10)

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: auth.AddressStoreKey(delAddr1), Value: cdc.MustMarshalBinaryBare(acc)},
		cmn.KVPair{Key: auth.GlobalAccountNumberKey, Value: cdc.MustMarshalBinaryLengthPrefixed(globalAccNumber)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}
	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Minter", fmt.Sprintf("%v\n%v", acc, acc)},
		{"GlobalAccNumber", fmt.Sprintf("GlobalAccNumberA: %d\nGlobalAccNumberB: %d", globalAccNumber, globalAccNumber)},
		{"other", ""},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeStore(cdc, cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeStore(cdc, cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
