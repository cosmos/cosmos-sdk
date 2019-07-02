package v0_36

import (
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_34"
)

func Migrate(oldGenState v034distr.GenesisState) GenesisState {
	feePool := FeePool{CommunityPool: oldGenState.FeePool.CommunityPool}

	dwifos := make([]DelegatorWithdrawInfo, len(oldGenState.DelegatorWithdrawInfos))
	for i, info := range oldGenState.DelegatorWithdrawInfos {
		dwifos[i] = DelegatorWithdrawInfo{
			DelegatorAddress: info.DelegatorAddress,
			WithdrawAddress:  info.WithdrawAddress,
		}
	}

	outRewards := make([]ValidatorOutstandingRewardsRecord, len(oldGenState.OutstandingRewards))
	for i, rew := range oldGenState.OutstandingRewards {
		outRewards[i] = ValidatorOutstandingRewardsRecord{
			ValidatorAddress:   rew.ValidatorAddress,
			OutstandingRewards: rew.OutstandingRewards,
		}
	}

	accumComm := make([]ValidatorAccumulatedCommissionRecord, len(oldGenState.ValidatorAccumulatedCommissions))
	for i, comm := range oldGenState.ValidatorAccumulatedCommissions {
		accumComm[i] = ValidatorAccumulatedCommissionRecord{
			ValidatorAddress: comm.ValidatorAddress,
			Accumulated:      comm.Accumulated,
		}
	}

	// migrate slash events which now have the period included
	slashEvents := make([]ValidatorSlashEventRecord, len(oldGenState.ValidatorSlashEvents))
	for i, se := range oldGenState.ValidatorSlashEvents {
		slashEvents[i] = ValidatorSlashEventRecord{
			ValidatorAddress: se.ValidatorAddress,
			Height:           se.Height,
			Period:           se.Event.ValidatorPeriod,
			Event: ValidatorSlashEvent{
				ValidatorPeriod: se.Event.ValidatorPeriod,
				Fraction:        se.Event.Fraction,
			},
		}
	}

	return NewGenesisState(
		feePool, oldGenState.CommunityTax, oldGenState.BaseProposerReward, oldGenState.BonusProposerReward,
		oldGenState.WithdrawAddrEnabled, dwifos, oldGenState.PreviousProposer, outRewards, accumComm,
		oldGenState.ValidatorHistoricalRewards, oldGenState.ValidatorCurrentRewards,
		oldGenState.DelegatorStartingInfos, slashEvents,
	)
}
