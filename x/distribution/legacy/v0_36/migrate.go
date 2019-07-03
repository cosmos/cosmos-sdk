package v0_36

import (
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_34"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.36
// genesis state. All entries are identical except for validator slashing events
// which now include the period.
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

	histRewards := make([]ValidatorHistoricalRewardsRecord, len(oldGenState.ValidatorHistoricalRewards))
	for i, rew := range oldGenState.ValidatorHistoricalRewards {
		histRewards[i] = ValidatorHistoricalRewardsRecord{
			ValidatorAddress: rew.ValidatorAddress,
			Period:           rew.Period,
			Rewards: ValidatorHistoricalRewards{
				CumulativeRewardRatio: rew.Rewards.CumulativeRewardRatio,
				ReferenceCount:        rew.Rewards.ReferenceCount,
			},
		}
	}

	currRewards := make([]ValidatorCurrentRewardsRecord, len(oldGenState.ValidatorCurrentRewards))
	for i, rew := range oldGenState.ValidatorCurrentRewards {
		currRewards[i] = ValidatorCurrentRewardsRecord{
			ValidatorAddress: rew.ValidatorAddress,
			Rewards: ValidatorCurrentRewards{
				Rewards: rew.Rewards.Rewards,
				Period:  rew.Rewards.Period,
			},
		}
	}

	delStartingInfos := make([]DelegatorStartingInfoRecord, len(oldGenState.DelegatorStartingInfos))
	for i, delInfo := range oldGenState.DelegatorStartingInfos {
		delStartingInfos[i] = DelegatorStartingInfoRecord{
			DelegatorAddress: delInfo.DelegatorAddress,
			ValidatorAddress: delInfo.ValidatorAddress,
			StartingInfo: DelegatorStartingInfo{
				PreviousPeriod: delInfo.StartingInfo.PreviousPeriod,
				Stake:          delInfo.StartingInfo.Stake,
				Height:         delInfo.StartingInfo.Height,
			},
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
		feePool, oldGenState.CommunityTax, oldGenState.BaseProposerReward,
		oldGenState.BonusProposerReward, oldGenState.WithdrawAddrEnabled,
		dwifos, oldGenState.PreviousProposer, outRewards, accumComm,
		histRewards, currRewards, delStartingInfos, slashEvents,
	)
}
