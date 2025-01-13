package simulation

import (
	"bytes"
	"fmt"

	"cosmossdk.io/core/codec"
	"cosmossdk.io/x/distribution/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding distribution type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.FeePoolKey):
			var feePoolA, feePoolB types.FeePool
			if err := cdc.Unmarshal(kvA.Value, &feePoolA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &feePoolB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorOutstandingRewardsPrefix):
			var rewardsA, rewardsB types.ValidatorOutstandingRewards
			if err := cdc.Unmarshal(kvA.Value, &rewardsA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &rewardsB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

		case bytes.Equal(kvA.Key[:1], types.DelegatorWithdrawAddrPrefix):
			return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

		case bytes.Equal(kvA.Key[:1], types.DelegatorStartingInfoPrefix):
			var infoA, infoB types.DelegatorStartingInfo
			if err := cdc.Unmarshal(kvA.Value, &infoA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &infoB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", infoA, infoB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorHistoricalRewardsPrefix):
			var rewardsA, rewardsB types.ValidatorHistoricalRewards
			if err := cdc.Unmarshal(kvA.Value, &rewardsA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &rewardsB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorCurrentRewardsPrefix):
			var rewardsA, rewardsB types.ValidatorCurrentRewards
			if err := cdc.Unmarshal(kvA.Value, &rewardsA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &rewardsB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorAccumulatedCommissionPrefix):
			var commissionA, commissionB types.ValidatorAccumulatedCommission
			if err := cdc.Unmarshal(kvA.Value, &commissionA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &commissionB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", commissionA, commissionB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorSlashEventPrefix):
			var eventA, eventB types.ValidatorSlashEvent
			if err := cdc.Unmarshal(kvA.Value, &eventA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &eventB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", eventA, eventB)

		default:
			panic(fmt.Sprintf("invalid distribution key prefix %X", kvA.Key[:1]))
		}
	}
}
