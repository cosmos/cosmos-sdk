package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
	valAddr1 = sdk.ValAddress(delPk1.Address())
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

	bondTime := time.Now().UTC()

	val := types.NewValidator(valAddr1, delPk1, types.NewDescription("test", "test", "test", "test", "test"))
	del := types.NewDelegation(delAddr1, valAddr1, sdk.OneDec())
	ubd := types.NewUnbondingDelegation(delAddr1, valAddr1, 15, bondTime, sdk.OneInt())
	red := types.NewRedelegation(delAddr1, valAddr1, valAddr1, 12, bondTime, sdk.OneInt(), sdk.OneDec())

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: types.LastTotalPowerKey, Value: cdc.MustMarshalBinaryLengthPrefixed(sdk.OneInt())},
		cmn.KVPair{Key: types.GetValidatorKey(valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(val)},
		cmn.KVPair{Key: types.LastValidatorPowerKey, Value: valAddr1.Bytes()},
		cmn.KVPair{Key: types.GetDelegationKey(delAddr1, valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(del)},
		cmn.KVPair{Key: types.GetUBDKey(delAddr1, valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(ubd)},
		cmn.KVPair{Key: types.GetREDKey(delAddr1, valAddr1, valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(red)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"LastTotalPower", fmt.Sprintf("%v\n%v", sdk.OneInt(), sdk.OneInt())},
		{"Validator", fmt.Sprintf("%v\n%v", val, val)},
		{"LastValidatorPower/ValidatorsByConsAddr/ValidatorsByPowerIndex", fmt.Sprintf("%v\n%v", valAddr1, valAddr1)},
		{"Delegation", fmt.Sprintf("%v\n%v", del, del)},
		{"UnbondingDelegation", fmt.Sprintf("%v\n%v", ubd, ubd)},
		{"Redelegation", fmt.Sprintf("%v\n%v", red, red)},
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
