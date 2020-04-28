package simulation_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/std"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	delPk1    = ed25519.GenPrivKey().PubKey()
	delAddr1  = sdk.AccAddress(delPk1.Address())
	valAddr1  = sdk.ValAddress(delPk1.Address())
	consAddr1 = sdk.ConsAddress(delPk1.Address().Bytes())
)

func TestDecodeDistributionStore(t *testing.T) {
	cdc := std.NewAppCodec(std.MakeCodec(simapp.ModuleBasics))
	dec := simulation.NewDecodeStore(cdc)

	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.OneDec())}
	feePool := types.InitialFeePool()
	feePool.CommunityPool = decCoins
	info := types.NewDelegatorStartingInfo(2, sdk.OneDec(), 200)
	outstanding := types.ValidatorOutstandingRewards{Rewards: decCoins}
	commission := types.ValidatorAccumulatedCommission{Commission: decCoins}
	historicalRewards := types.NewValidatorHistoricalRewards(decCoins, 100)
	currentRewards := types.NewValidatorCurrentRewards(decCoins, 5)
	slashEvent := types.NewValidatorSlashEvent(10, sdk.OneDec())

	kvPairs := tmkv.Pairs{
		tmkv.Pair{Key: types.FeePoolKey, Value: cdc.MustMarshalBinaryBare(&feePool)},
		tmkv.Pair{Key: types.ProposerKey, Value: consAddr1.Bytes()},
		tmkv.Pair{Key: types.GetValidatorOutstandingRewardsKey(valAddr1), Value: cdc.MustMarshalBinaryBare(&outstanding)},
		tmkv.Pair{Key: types.GetDelegatorWithdrawAddrKey(delAddr1), Value: delAddr1.Bytes()},
		tmkv.Pair{Key: types.GetDelegatorStartingInfoKey(valAddr1, delAddr1), Value: cdc.MustMarshalBinaryBare(&info)},
		tmkv.Pair{Key: types.GetValidatorHistoricalRewardsKey(valAddr1, 100), Value: cdc.MustMarshalBinaryBare(&historicalRewards)},
		tmkv.Pair{Key: types.GetValidatorCurrentRewardsKey(valAddr1), Value: cdc.MustMarshalBinaryBare(&currentRewards)},
		tmkv.Pair{Key: types.GetValidatorAccumulatedCommissionKey(valAddr1), Value: cdc.MustMarshalBinaryBare(&commission)},
		tmkv.Pair{Key: types.GetValidatorSlashEventKeyPrefix(valAddr1, 13), Value: cdc.MustMarshalBinaryBare(&slashEvent)},
		tmkv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"FeePool", fmt.Sprintf("%v\n%v", feePool, feePool)},
		{"Proposer", fmt.Sprintf("%v\n%v", consAddr1, consAddr1)},
		{"ValidatorOutstandingRewards", fmt.Sprintf("%v\n%v", outstanding, outstanding)},
		{"DelegatorWithdrawAddr", fmt.Sprintf("%v\n%v", delAddr1, delAddr1)},
		{"DelegatorStartingInfo", fmt.Sprintf("%v\n%v", info, info)},
		{"ValidatorHistoricalRewards", fmt.Sprintf("%v\n%v", historicalRewards, historicalRewards)},
		{"ValidatorCurrentRewards", fmt.Sprintf("%v\n%v", currentRewards, currentRewards)},
		{"ValidatorAccumulatedCommission", fmt.Sprintf("%v\n%v", commission, commission)},
		{"ValidatorSlashEvent", fmt.Sprintf("%v\n%v", slashEvent, slashEvent)},
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
