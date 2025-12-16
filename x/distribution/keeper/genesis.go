package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// InitGenesis sets distribution information for genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	var moduleHoldings sdk.DecCoins

	err := k.FeePool.Set(ctx, data.FeePool)
	if err != nil {
		return err
	}

	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}

	for _, dwi := range data.DelegatorWithdrawInfos {
		delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(dwi.DelegatorAddress)
		if err != nil {
			return err
		}
		withdrawAddress, err := k.authKeeper.AddressCodec().StringToBytes(dwi.WithdrawAddress)
		if err != nil {
			return err
		}
		err = k.SetDelegatorWithdrawAddr(ctx, delegatorAddress, withdrawAddress)
		if err != nil {
			return err
		}
	}

	var previousProposer sdk.ConsAddress
	if data.PreviousProposer != "" {
		var err error
		previousProposer, err = k.stakingKeeper.ConsensusAddressCodec().StringToBytes(data.PreviousProposer)
		if err != nil {
			return err
		}
	}

	if err = k.SetPreviousProposerConsAddr(ctx, previousProposer); err != nil {
		return err
	}

	for _, rew := range data.OutstandingRewards {
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(rew.ValidatorAddress)
		if err != nil {
			return err
		}
		err = k.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: rew.OutstandingRewards})
		if err != nil {
			return err
		}
		moduleHoldings = moduleHoldings.Add(rew.OutstandingRewards...)
	}
	for _, acc := range data.ValidatorAccumulatedCommissions {
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(acc.ValidatorAddress)
		if err != nil {
			return err
		}
		err = k.SetValidatorAccumulatedCommission(ctx, valAddr, acc.Accumulated)
		if err != nil {
			return err
		}
	}
	for _, his := range data.ValidatorHistoricalRewards {
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(his.ValidatorAddress)
		if err != nil {
			return err
		}
		err = k.SetValidatorHistoricalRewards(ctx, valAddr, his.Period, his.Rewards)
		if err != nil {
			return err
		}
	}
	for _, cur := range data.ValidatorCurrentRewards {
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(cur.ValidatorAddress)
		if err != nil {
			return err
		}
		err = k.SetValidatorCurrentRewards(ctx, valAddr, cur.Rewards)
		if err != nil {
			return err
		}
	}
	for _, del := range data.DelegatorStartingInfos {
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(del.ValidatorAddress)
		if err != nil {
			return err
		}
		delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(del.DelegatorAddress)
		if err != nil {
			return err
		}

		err = k.SetDelegatorStartingInfo(ctx, valAddr, delegatorAddress, del.StartingInfo)
		if err != nil {
			return err
		}
	}
	for _, evt := range data.ValidatorSlashEvents {
		valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(evt.ValidatorAddress)
		if err != nil {
			return err
		}
		err = k.SetValidatorSlashEvent(ctx, valAddr, evt.Height, evt.Period, evt.ValidatorSlashEvent)
		if err != nil {
			return err
		}
	}

	moduleHoldings = moduleHoldings.Add(data.FeePool.CommunityPool...)
	moduleHoldingsInt, _ := moduleHoldings.TruncateDecimal()

	// check if the module account exists
	moduleAcc := k.GetDistributionAccount(ctx)
	if moduleAcc == nil {
		return fmt.Errorf("%s module account has not been set", types.ModuleName)
	}

	balances := k.bankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())
	if balances.IsZero() {
		k.authKeeper.SetModuleAccount(ctx, moduleAcc)
	}
	if !balances.Equal(moduleHoldingsInt) {
		return fmt.Errorf("distribution module balance does not match the module holdings: %s <-> %s", balances, moduleHoldingsInt)
	}

	// set Nakamoto Bonus
	if data.NakamotoBonus.IsZero() {
		// zero is allowed but ADR default is 0.03. If not provided, set default.
		if err := k.SetNakamotoBonusCoefficient(ctx, types.DefaultNakamotoBonus); err != nil {
			return err
		}
	} else {
		if err := k.SetNakamotoBonusCoefficient(ctx, data.NakamotoBonus); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return nil, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	dwi := make([]types.DelegatorWithdrawInfo, 0)
	k.IterateDelegatorWithdrawAddrs(ctx, func(del, addr sdk.AccAddress) (stop bool) {
		dwi = append(dwi, types.DelegatorWithdrawInfo{
			DelegatorAddress: del.String(),
			WithdrawAddress:  addr.String(),
		})
		return false
	})

	pp, err := k.GetPreviousProposerConsAddr(ctx)
	if err != nil {
		return nil, err
	}

	outstanding := make([]types.ValidatorOutstandingRewardsRecord, 0)

	k.IterateValidatorOutstandingRewards(ctx,
		func(addr sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
			outstanding = append(outstanding, types.ValidatorOutstandingRewardsRecord{
				ValidatorAddress:   addr.String(),
				OutstandingRewards: rewards.Rewards,
			})
			return false
		},
	)

	acc := make([]types.ValidatorAccumulatedCommissionRecord, 0)
	k.IterateValidatorAccumulatedCommissions(ctx,
		func(addr sdk.ValAddress, commission types.ValidatorAccumulatedCommission) (stop bool) {
			acc = append(acc, types.ValidatorAccumulatedCommissionRecord{
				ValidatorAddress: addr.String(),
				Accumulated:      commission,
			})
			return false
		},
	)

	his := make([]types.ValidatorHistoricalRewardsRecord, 0)
	k.IterateValidatorHistoricalRewards(ctx,
		func(val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) (stop bool) {
			his = append(his, types.ValidatorHistoricalRewardsRecord{
				ValidatorAddress: val.String(),
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
				ValidatorAddress: val.String(),
				Rewards:          rewards,
			})
			return false
		},
	)

	dels := make([]types.DelegatorStartingInfoRecord, 0)
	k.IterateDelegatorStartingInfos(ctx,
		func(val sdk.ValAddress, del sdk.AccAddress, info types.DelegatorStartingInfo) (stop bool) {
			dels = append(dels, types.DelegatorStartingInfoRecord{
				ValidatorAddress: val.String(),
				DelegatorAddress: del.String(),
				StartingInfo:     info,
			})
			return false
		},
	)

	slashes := make([]types.ValidatorSlashEventRecord, 0)
	k.IterateValidatorSlashEvents(ctx,
		func(val sdk.ValAddress, height uint64, event types.ValidatorSlashEvent) (stop bool) {
			slashes = append(slashes, types.ValidatorSlashEventRecord{
				ValidatorAddress:    val.String(),
				Height:              height,
				Period:              event.ValidatorPeriod,
				ValidatorSlashEvent: event,
			})
			return false
		},
	)

	nb, err := k.GetNakamotoBonusCoefficient(ctx)
	if err != nil {
		return nil, err
	}

	return types.NewGenesisState(params, feePool, dwi, pp, outstanding, acc, his, cur, dels, slashes, nb), nil
}
