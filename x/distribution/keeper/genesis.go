package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// InitGenesis sets distribution information for genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	var moduleHoldings sdk.DecCoins

	k.SetFeePool(ctx, data.FeePool)
	k.SetParams(ctx, data.Params)

	for _, dwi := range data.DelegatorWithdrawInfos {
		k.SetDelegatorWithdrawAddr(ctx, dwi.DelegatorAddress, dwi.WithdrawAddress)
	}

	k.SetPreviousProposerConsAddr(ctx, data.PreviousProposer)

	for _, rew := range data.OutstandingRewards {
		k.SetValidatorOutstandingRewards(ctx, rew.ValidatorAddress, types.ValidatorOutstandingRewards{Rewards: rew.OutstandingRewards})
		moduleHoldings = moduleHoldings.Add(rew.OutstandingRewards...)
	}
	for _, acc := range data.ValidatorAccumulatedCommissions {
		k.SetValidatorAccumulatedCommission(ctx, acc.ValidatorAddress, acc.Accumulated)
	}
	for _, his := range data.ValidatorHistoricalRewards {
		k.SetValidatorHistoricalRewards(ctx, his.ValidatorAddress, his.Period, his.Rewards)
	}
	for _, cur := range data.ValidatorCurrentRewards {
		k.SetValidatorCurrentRewards(ctx, cur.ValidatorAddress, cur.Rewards)
	}
	for _, del := range data.DelegatorStartingInfos {
		k.SetDelegatorStartingInfo(ctx, del.ValidatorAddress, del.DelegatorAddress, del.StartingInfo)
	}
	for _, evt := range data.ValidatorSlashEvents {
		k.SetValidatorSlashEvent(ctx, evt.ValidatorAddress, evt.Height, evt.Period, evt.Event)
	}

	moduleHoldings = moduleHoldings.Add(data.FeePool.CommunityPool...)
	moduleHoldingsInt, _ := moduleHoldings.TruncateDecimal()

	// check if the module account exists
	moduleAcc := k.GetDistributionAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	balances := k.bankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())
	if balances.IsZero() {
		if err := k.bankKeeper.SetBalances(ctx, moduleAcc.GetAddress(), moduleHoldingsInt); err != nil {
			panic(err)
		}

		k.authKeeper.SetModuleAccount(ctx, moduleAcc)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	feePool := k.GetFeePool(ctx)
	params := k.GetParams(ctx)

	dwi := make([]types.DelegatorWithdrawInfo, 0)
	k.IterateDelegatorWithdrawAddrs(ctx, func(del sdk.AccAddress, addr sdk.AccAddress) (stop bool) {
		dwi = append(dwi, types.DelegatorWithdrawInfo{
			DelegatorAddress: del,
			WithdrawAddress:  addr,
		})
		return false
	})

	pp := k.GetPreviousProposerConsAddr(ctx)
	outstanding := make([]types.ValidatorOutstandingRewardsRecord, 0)

	k.IterateValidatorOutstandingRewards(ctx,
		func(addr sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
			outstanding = append(outstanding, types.ValidatorOutstandingRewardsRecord{
				ValidatorAddress:   addr,
				OutstandingRewards: rewards.Rewards,
			})
			return false
		},
	)

	acc := make([]types.ValidatorAccumulatedCommissionRecord, 0)
	k.IterateValidatorAccumulatedCommissions(ctx,
		func(addr sdk.ValAddress, commission types.ValidatorAccumulatedCommission) (stop bool) {
			acc = append(acc, types.ValidatorAccumulatedCommissionRecord{
				ValidatorAddress: addr,
				Accumulated:      commission,
			})
			return false
		},
	)

	his := make([]types.ValidatorHistoricalRewardsRecord, 0)
	k.IterateValidatorHistoricalRewards(ctx,
		func(val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) (stop bool) {
			his = append(his, types.ValidatorHistoricalRewardsRecord{
				ValidatorAddress: val,
				Period:           period,
				Rewards:          rewards,
			})
			return false
		},
	)

	cur := make([]types.ValidatorCurrentRewardsRecord, 0)
	k.IterateValidatorCurrentRewards(ctx,
		func(val sdk.ValAddress, rewards types.ValidatorCurrentRewards) (stop bool) {
			cur = append(cur, types.ValidatorCurrentRewardsRecord{
				ValidatorAddress: val,
				Rewards:          rewards,
			})
			return false
		},
	)

	dels := make([]types.DelegatorStartingInfoRecord, 0)
	k.IterateDelegatorStartingInfos(ctx,
		func(val sdk.ValAddress, del sdk.AccAddress, info types.DelegatorStartingInfo) (stop bool) {
			dels = append(dels, types.DelegatorStartingInfoRecord{
				ValidatorAddress: val,
				DelegatorAddress: del,
				StartingInfo:     info,
			})
			return false
		},
	)

	slashes := make([]types.ValidatorSlashEventRecord, 0)
	k.IterateValidatorSlashEvents(ctx,
		func(val sdk.ValAddress, height uint64, event types.ValidatorSlashEvent) (stop bool) {
			slashes = append(slashes, types.ValidatorSlashEventRecord{
				ValidatorAddress: val,
				Height:           height,
				Period:           event.ValidatorPeriod,
				Event:            event,
			})
			return false
		},
	)

	return types.NewGenesisState(params, feePool, dwi, pp, outstanding, acc, his, cur, dels, slashes)
}
