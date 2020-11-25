package simulation_test

import (
	"fmt"
	"testing"
	"time"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// nolint:deadcode,unused,varcheck
var (
	delPk1    = ed25519.GenPrivKey().PubKey()
	delAddr1  = sdk.AccAddress(delPk1.Address())
	valAddr1  = sdk.ValAddress(delPk1.Address())
	consAddr1 = sdk.ConsAddress(delPk1.Address().Bytes())
)

func TestDecodeStore(t *testing.T) {
	cdc, _ := simapp.MakeCodecs()
	dec := simulation.NewDecodeStore(cdc)

	info := types.NewValidatorSigningInfo(consAddr1, 0, 1, time.Now().UTC(), false, 0)
	bechPK := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, delPk1)
	missed := gogotypes.BoolValue{Value: true}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.ValidatorSigningInfoKey(consAddr1), Value: cdc.MustMarshalBinaryBare(&info)},
			{Key: types.ValidatorMissedBlockBitArrayKey(consAddr1, 6), Value: cdc.MustMarshalBinaryBare(&missed)},
			{Key: types.AddrPubkeyRelationKey(delAddr1), Value: cdc.MustMarshalBinaryBare(&gogotypes.StringValue{Value: bechPK})},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"ValidatorSigningInfo", fmt.Sprintf("%v\n%v", info, info)},
		{"ValidatorMissedBlockBitArray", fmt.Sprintf("missedA: %v\nmissedB: %v", missed.Value, missed.Value)},
		{"AddrPubkeyRelation", fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", bechPK, bechPK)},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
