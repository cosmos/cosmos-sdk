package simulation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

var (
	delPk1    = ed25519.GenPrivKey().PubKey()
	delAddr1  = sdk.AccAddress(delPk1.Address())
	consAddr1 = sdk.ConsAddress(delPk1.Address().Bytes())
)

func TestDecodeStore(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig(slashing.AppModuleBasic{})
	cdc := encodingConfig.Codec
	dec := simulation.NewDecodeStore(cdc)

	info := types.NewValidatorSigningInfo(consAddr1, 0, 1, time.Now().UTC(), false, 0)
	missed := []byte{1} // we want to display the bytes for simulation diffs
	bz, err := cdc.MarshalInterface(delPk1)
	require.NoError(t, err)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.ValidatorSigningInfoKey(consAddr1), Value: cdc.MustMarshal(&info)},
			{Key: types.ValidatorMissedBlockBitmapKey(consAddr1, 6), Value: missed},
			{Key: types.AddrPubkeyRelationKey(delAddr1), Value: bz},
			{Key: []byte{0x99}, Value: []byte{0x99}}, // This test should panic
		},
	}

	tests := []struct {
		name        string
		expectedLog string
		panics      bool
	}{
		{"ValidatorSigningInfo", fmt.Sprintf("%v\n%v", info, info), false},
		{"ValidatorMissedBlockBitArray", fmt.Sprintf("missedA: %v\nmissedB: %v\n", missed, missed), false},
		{"AddrPubkeyRelation", fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", delPk1, delPk1), false},
		{"other", "", true},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			} else {
				require.Contains(t, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.expectedLog, tt.name)
			}
		})
	}
}
