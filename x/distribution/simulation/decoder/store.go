package decoder

import (
	"bytes"
	"fmt"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding distribution type
func DecodeStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], keeper.FeePoolKey):
		var feePoolA, feePoolB types.FeePool
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &feePoolA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &feePoolB)
		return fmt.Sprintf("%v\n%v", feePoolA, feePoolB)

	case bytes.Equal(kvA.Key[:1],  keeper.ProposerKey):
		return fmt.Sprintf("%v\n%v", sdk.ConsAddress(kvA.Value), sdk.ConsAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1],  keeper.ValidatorOutstandingRewardsPrefix):
		var rewardsA, rewardsB types.ValidatorOutstandingRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1],  keeper.DelegatorWithdrawAddrPrefix):
		return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

	case bytes.Equal(kvA.Key[:1],  keeper.DelegatorStartingInfoPrefix):
		var infoA, infoB types.DelegatorStartingInfo
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &infoA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1],  keeper.ValidatorHistoricalRewardsPrefix):
		var rewardsA, rewardsB types.ValidatorHistoricalRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1],  keeper.ValidatorCurrentRewardsPrefix):
		var rewardsA, rewardsB types.ValidatorCurrentRewards
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardsA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardsB)
		return fmt.Sprintf("%v\n%v", rewardsA, rewardsB)

	case bytes.Equal(kvA.Key[:1],  keeper.ValidatorAccumulatedCommissionPrefix):
		var commissionA, commissionB types.ValidatorAccumulatedCommission
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &commissionA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &commissionB)
		return fmt.Sprintf("%v\n%v", commissionA, commissionB)

	case bytes.Equal(kvA.Key[:1],  keeper.ValidatorSlashEventPrefix):
		var eventA, eventB types.ValidatorSlashEvent
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &eventA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &eventB)
		return fmt.Sprintf("%v\n%v", eventA, eventB)

	default:
		panic(fmt.Sprintf("invalid distribution key prefix %X", kvA.Key[:1]))
	}
}
