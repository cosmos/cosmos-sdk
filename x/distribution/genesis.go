package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// InitGenesis sets distribution information for genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	keeper.SetFeePool(ctx, data.FeePool)
	keeper.SetCommunityTax(ctx, data.CommunityTax)
	keeper.SetBaseProposerReward(ctx, data.BaseProposerReward)
	keeper.SetBonusProposerReward(ctx, data.BonusProposerReward)
	keeper.SetWithdrawAddrEnabled(ctx, data.WithdrawAddrEnabled)
	for _, dwi := range data.DelegatorWithdrawInfos {
		keeper.SetDelegatorWithdrawAddr(ctx, dwi.DelegatorAddr, dwi.WithdrawAddr)
	}
	keeper.SetPreviousProposerConsAddr(ctx, data.PreviousProposer)
	keeper.SetOutstandingRewards(ctx, data.OutstandingRewards)
	for _, acc := range data.ValidatorAccumulatedCommissions {
		keeper.SetValidatorAccumulatedCommission(ctx, acc.ValidatorAddr, acc.Accumulated)
	}
	for _, his := range data.ValidatorHistoricalRewards {
		keeper.SetValidatorHistoricalRewards(ctx, his.ValidatorAddr, his.Period, his.Rewards)
	}
	for _, cur := range data.ValidatorCurrentRewards {
		keeper.SetValidatorCurrentRewards(ctx, cur.ValidatorAddr, cur.Rewards)
	}
	for _, del := range data.DelegatorStartingInfos {
		keeper.SetDelegatorStartingInfo(ctx, del.ValidatorAddr, del.DelegatorAddr, del.StartingInfo)
	}
	for _, evt := range data.ValidatorSlashEvents {
		keeper.SetValidatorSlashEvent(ctx, evt.ValidatorAddr, evt.Height, evt.Event)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	feePool := keeper.GetFeePool(ctx)
	communityTax := keeper.GetCommunityTax(ctx)
	baseProposerRewards := keeper.GetBaseProposerReward(ctx)
	bonusProposerRewards := keeper.GetBonusProposerReward(ctx)
	withdrawAddrEnabled := keeper.GetWithdrawAddrEnabled(ctx)
	dwi := make([]types.DelegatorWithdrawInfo, 0)
	keeper.IterateDelegatorWithdrawAddrs(ctx, func(del sdk.AccAddress, addr sdk.AccAddress) (stop bool) {
		dwi = append(dwi, types.DelegatorWithdrawInfo{
			DelegatorAddr: del,
			WithdrawAddr:  addr,
		})
		return false
	})
	pp := keeper.GetPreviousProposerConsAddr(ctx)
	outstanding := keeper.GetOutstandingRewards(ctx)
	acc := make([]types.ValidatorAccumulatedCommissionRecord, 0)
	keeper.IterateValidatorAccumulatedCommissions(ctx,
		func(addr sdk.ValAddress, commission types.ValidatorAccumulatedCommission) (stop bool) {
			acc = append(acc, types.ValidatorAccumulatedCommissionRecord{
				ValidatorAddr: addr,
				Accumulated:   commission,
			})
			return false
		},
	)
	his := make([]types.ValidatorHistoricalRewardsRecord, 0)
	keeper.IterateValidatorHistoricalRewards(ctx,
		func(val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) (stop bool) {
			his = append(his, types.ValidatorHistoricalRewardsRecord{
				ValidatorAddr: val,
				Period:        period,
				Rewards:       rewards,
			})
			return false
		},
	)
	cur := make([]types.ValidatorCurrentRewardsRecord, 0)
	keeper.IterateValidatorCurrentRewards(ctx,
		func(val sdk.ValAddress, rewards types.ValidatorCurrentRewards) (stop bool) {
			cur = append(cur, types.ValidatorCurrentRewardsRecord{
				ValidatorAddr: val,
				Rewards:       rewards,
			})
			return false
		},
	)
	dels := make([]types.DelegatorStartingInfoRecord, 0)
	keeper.IterateDelegatorStartingInfos(ctx,
		func(val sdk.ValAddress, del sdk.AccAddress, info types.DelegatorStartingInfo) (stop bool) {
			dels = append(dels, types.DelegatorStartingInfoRecord{
				ValidatorAddr: val,
				DelegatorAddr: del,
				StartingInfo:  info,
			})
			return false
		},
	)
	slashes := make([]types.ValidatorSlashEventRecord, 0)
	keeper.IterateValidatorSlashEvents(ctx,
		func(val sdk.ValAddress, height uint64, event types.ValidatorSlashEvent) (stop bool) {
			slashes = append(slashes, types.ValidatorSlashEventRecord{
				ValidatorAddr: val,
				Height:        height,
				Event:         event,
			})
			return false
		},
	)
	return types.NewGenesisState(feePool, communityTax, baseProposerRewards, bonusProposerRewards, withdrawAddrEnabled,
		dwi, pp, outstanding, acc, his, cur, dels, slashes)
}
