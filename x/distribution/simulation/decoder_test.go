package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	delPk1    = ed25519.GenPrivKey().PubKey()
	valAddr1  = sdk.ValAddress(delPk1.Address())
	consAddr1 = sdk.ConsAddress(delPk1.Address().Bytes())
)

func TestDecodeDistributionStore(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	cdc := encodingConfig.Codec

	dec := simulation.NewDecodeStore(cdc)

	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyOneDec())}
	feePool := types.InitialFeePool()
	feePool.CommunityPool = decCoins
	slashEvent := types.NewValidatorSlashEvent(10, math.LegacyOneDec())

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.FeePoolKey, Value: cdc.MustMarshal(&feePool)},
			{Key: types.ProposerKey, Value: consAddr1.Bytes()},
			{Key: types.GetValidatorSlashEventKeyPrefix(valAddr1, 13), Value: cdc.MustMarshal(&slashEvent)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"FeePool", fmt.Sprintf("%v\n%v", feePool, feePool)},
		{"Proposer", fmt.Sprintf("%v\n%v", consAddr1, consAddr1)},
		{"ValidatorSlashEvent", fmt.Sprintf("%v\n%v", slashEvent, slashEvent)},
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
