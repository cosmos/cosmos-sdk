package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding distribution type.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.FeePoolKey):
			var feePoolA, feePoolB types.FeePool
			cdc.MustUnmarshalBinaryBare(kvA.Value, &feePoolA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &feePoolB)
			return fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

		case bytes.Equal(kvA.Key[:1], types.ProposerKey):
			return fmt.Sprintf("%v\n%v", sdk.ConsAddress(kvA.Value), sdk.ConsAddress(kvB.Value))

		case bytes.Equal(kvA.Key[:1], types.ValidatorOutstandingRewardsPrefix):
			var rewardsA, rewardsB types.ValidatorOutstandingRewards
			cdc.MustUnmarshalBinaryBare(kvA.Value, &rewardsA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &rewardsB)
			return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

		case bytes.Equal(kvA.Key[:1], types.DelegatorWithdrawAddrPrefix):
			return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

		case bytes.Equal(kvA.Key[:1], types.DelegatorStartingInfoPrefix):
			var infoA, infoB types.DelegatorStartingInfo
			cdc.MustUnmarshalBinaryBare(kvA.Value, &infoA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &infoB)
			return fmt.Sprintf("%v\n%v", infoA, infoB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorHistoricalRewardsPrefix):
			var rewardsA, rewardsB types.ValidatorHistoricalRewards
			cdc.MustUnmarshalBinaryBare(kvA.Value, &rewardsA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &rewardsB)
			return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorCurrentRewardsPrefix):
			var rewardsA, rewardsB types.ValidatorCurrentRewards
			cdc.MustUnmarshalBinaryBare(kvA.Value, &rewardsA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &rewardsB)
			return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorAccumulatedCommissionPrefix):
			var commissionA, commissionB types.ValidatorAccumulatedCommission
			cdc.MustUnmarshalBinaryBare(kvA.Value, &commissionA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &commissionB)
			return fmt.Sprintf("%v\n%v", commissionA, commissionB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorSlashEventPrefix):
			var eventA, eventB types.ValidatorSlashEvent
			cdc.MustUnmarshalBinaryBare(kvA.Value, &eventA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &eventB)
			return fmt.Sprintf("%v\n%v", eventA, eventB)

		default:
			panic(fmt.Sprintf("invalid distribution key prefix %X", kvA.Key[:1]))
		}
	}
}
