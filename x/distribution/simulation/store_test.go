package simulation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"

	distr "github.com/cosmos/cosmos-sdk/x/distribution"

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
	distr.RegisterCodec(cdc)
	return
}

func TestDecodeDistributionStore(t *testing.T) {
	cdc := makeTestCodec()

	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.OneDec())}
	feePool := distr.InitialFeePool()
	feePool.CommunityPool = decCoins
	info := distr.NewDelegatorStartingInfo(2, sdk.OneDec(), 200)
	outstanding := distr.ValidatorOutstandingRewards{decCoins[0]}
	commission := distr.ValidatorAccumulatedCommission{decCoins[0]}
	historicalRewards := distr.NewValidatorHistoricalRewards(decCoins, 100)
	currentRewards := distr.NewValidatorCurrentRewards(decCoins, 5)
	slashEvent := distr.NewValidatorSlashEvent(10, sdk.OneDec())

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: distr.FeePoolKey, Value: cdc.MustMarshalBinaryLengthPrefixed(feePool)},
		cmn.KVPair{Key: distr.ProposerKey, Value: consAddr1.Bytes()},
		cmn.KVPair{Key: distr.GetValidatorOutstandingRewardsKey(valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(outstanding)},
		cmn.KVPair{Key: distr.GetDelegatorWithdrawAddrKey(delAddr1), Value: delAddr1.Bytes()},
		cmn.KVPair{Key: distr.GetDelegatorStartingInfoKey(valAddr1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(info)},
		cmn.KVPair{Key: distr.GetValidatorHistoricalRewardsKey(valAddr1, 100), Value: cdc.MustMarshalBinaryLengthPrefixed(historicalRewards)},
		cmn.KVPair{Key: distr.GetValidatorCurrentRewardsKey(valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(currentRewards)},
		cmn.KVPair{Key: distr.GetValidatorAccumulatedCommissionKey(valAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(commission)},
		cmn.KVPair{Key: distr.GetValidatorSlashEventKeyPrefix(valAddr1, 13), Value: cdc.MustMarshalBinaryLengthPrefixed(slashEvent)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
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
