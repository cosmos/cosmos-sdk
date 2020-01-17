package simulation

import (
	"bytes"
	"fmt"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding distribution type
func DecodeStore(cdc *codec.Codec, kvA, kvB tmkv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.FeePoolKey):
		var feePoolA, feePoolB types.FeePool
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &feePoolA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &feePoolB)
		return fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

	case bytes.Equal(kvA.Key[:1], types.ProposerKey):
		return fmt.Sprintf("%v\n%v", sdk.ConsAddress(kvA.Value), sdk.ConsAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], types.ValidatorOutstandingRewardsPrefix):
		var rewardsA, rewardsB types.ValidatorOutstandingRewards
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], types.DelegatorWithdrawAddrPrefix):
		return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1], types.DelegatorStartingInfoPrefix):
		var infoA, infoB types.DelegatorStartingInfo
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &infoA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorHistoricalRewardsPrefix):
		var rewardsA, rewardsB types.ValidatorHistoricalRewards
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorCurrentRewardsPrefix):
		var rewardsA, rewardsB types.ValidatorCurrentRewards
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorAccumulatedCommissionPrefix):
		var commissionA, commissionB types.ValidatorAccumulatedCommission
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &commissionA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &commissionB)
		return fmt.Sprintf("%v\n%v", commissionA, commissionB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorSlashEventPrefix):
		var eventA, eventB types.ValidatorSlashEvent
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &eventA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &eventB)
		return fmt.Sprintf("%v\n%v", eventA, eventB)

	default:
		panic(fmt.Sprintf("invalid distribution key prefix %X", kvA.Key[:1]))
	}
}
