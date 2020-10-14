package v040

import (
	v038distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v038"
	v040distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Migrate accepts exported x/distribution genesis state from v0.38 and migrates it
// to v0.40 x/distribution genesis state. The migration includes:
//
// - Convert addresses from bytes to bech32 strings.
// - Re-encode in v0.40 GenesisState.
func Migrate(oldDistributionState v038distribution.GenesisState) *v040distribution.GenesisState {
	newDelegatorWithdrawInfos := make([]v040distribution.DelegatorWithdrawInfo, len(oldDistributionState.DelegatorWithdrawInfos))
	for i, oldDelegatorWithdrawInfo := range oldDistributionState.DelegatorWithdrawInfos {
		newDelegatorWithdrawInfos[i] = v040distribution.DelegatorWithdrawInfo{
			DelegatorAddress: oldDelegatorWithdrawInfo.DelegatorAddress.String(),
			WithdrawAddress:  oldDelegatorWithdrawInfo.WithdrawAddress.String(),
		}
	}

	newValidatorOutstandingRewards := make([]v040distribution.ValidatorOutstandingRewardsRecord, len(oldDistributionState.OutstandingRewards))
	for i, oldValidatorOutstandingReward := range oldDistributionState.OutstandingRewards {
		newValidatorOutstandingRewards[i] = v040distribution.ValidatorOutstandingRewardsRecord{
			ValidatorAddress:   oldValidatorOutstandingReward.ValidatorAddress.String(),
			OutstandingRewards: oldValidatorOutstandingReward.OutstandingRewards,
		}
	}

	newValidatorAccumulatedCommissions := make([]v040distribution.ValidatorAccumulatedCommissionRecord, len(oldDistributionState.ValidatorAccumulatedCommissions))
	for i, oldValidatorAccumulatedCommission := range oldDistributionState.ValidatorAccumulatedCommissions {
		newValidatorAccumulatedCommissions[i] = v040distribution.ValidatorAccumulatedCommissionRecord{
			ValidatorAddress: oldValidatorAccumulatedCommission.ValidatorAddress.String(),
			Accumulated: v040distribution.ValidatorAccumulatedCommission{
				Commission: oldValidatorAccumulatedCommission.Accumulated,
			},
		}
	}

	newValidatorHistoricalRewards := make([]v040distribution.ValidatorHistoricalRewardsRecord, len(oldDistributionState.ValidatorHistoricalRewards))
	for i, oldValidatorHistoricalReward := range oldDistributionState.ValidatorHistoricalRewards {
		newValidatorHistoricalRewards[i] = v040distribution.ValidatorHistoricalRewardsRecord{
			ValidatorAddress: oldValidatorHistoricalReward.ValidatorAddress.String(),
			Period:           oldValidatorHistoricalReward.Period,
			Rewards: v040distribution.ValidatorHistoricalRewards{
				CumulativeRewardRatio: oldValidatorHistoricalReward.Rewards.CumulativeRewardRatio,
				ReferenceCount:        uint32(oldValidatorHistoricalReward.Rewards.ReferenceCount),
			},
		}
	}

	newValidatorCurrentRewards := make([]v040distribution.ValidatorCurrentRewardsRecord, len(oldDistributionState.ValidatorCurrentRewards))
	for i, oldValidatorCurrentReward := range oldDistributionState.ValidatorCurrentRewards {
		newValidatorCurrentRewards[i] = v040distribution.ValidatorCurrentRewardsRecord{
			ValidatorAddress: oldValidatorCurrentReward.ValidatorAddress.String(),
			Rewards: v040distribution.ValidatorCurrentRewards{
				Rewards: oldValidatorCurrentReward.Rewards.Rewards,
				Period:  oldValidatorCurrentReward.Rewards.Period,
			},
		}
	}

	newDelegatorStartingInfos := make([]v040distribution.DelegatorStartingInfoRecord, len(oldDistributionState.DelegatorStartingInfos))
	for i, oldDelegatorStartingInfo := range oldDistributionState.DelegatorStartingInfos {
		newDelegatorStartingInfos[i] = v040distribution.DelegatorStartingInfoRecord{
			DelegatorAddress: oldDelegatorStartingInfo.DelegatorAddress.String(),
			ValidatorAddress: oldDelegatorStartingInfo.ValidatorAddress.String(),
			StartingInfo: v040distribution.DelegatorStartingInfo{
				PreviousPeriod: oldDelegatorStartingInfo.StartingInfo.PreviousPeriod,
				Stake:          oldDelegatorStartingInfo.StartingInfo.Stake,
				Height:         oldDelegatorStartingInfo.StartingInfo.Height,
			},
		}
	}

	newValidatorSlashEvents := make([]v040distribution.ValidatorSlashEventRecord, len(oldDistributionState.ValidatorSlashEvents))
	for i, oldValidatorSlashEvent := range oldDistributionState.ValidatorSlashEvents {
		newValidatorSlashEvents[i] = v040distribution.ValidatorSlashEventRecord{
			ValidatorAddress: oldValidatorSlashEvent.ValidatorAddress.String(),
			Height:           oldValidatorSlashEvent.Height,
			Period:           oldValidatorSlashEvent.Period,
			ValidatorSlashEvent: v040distribution.ValidatorSlashEvent{
				ValidatorPeriod: oldValidatorSlashEvent.Event.ValidatorPeriod,
				Fraction:        oldValidatorSlashEvent.Event.Fraction,
			},
		}
	}

	return &v040distribution.GenesisState{
		Params: v040distribution.Params{
			CommunityTax:        oldDistributionState.Params.CommunityTax,
			BaseProposerReward:  oldDistributionState.Params.BaseProposerReward,
			BonusProposerReward: oldDistributionState.Params.BonusProposerReward,
			WithdrawAddrEnabled: oldDistributionState.Params.WithdrawAddrEnabled,
		},
		FeePool: v040distribution.FeePool{
			CommunityPool: oldDistributionState.FeePool.CommunityPool,
		},
		DelegatorWithdrawInfos:          newDelegatorWithdrawInfos,
		PreviousProposer:                oldDistributionState.PreviousProposer.String(),
		OutstandingRewards:              newValidatorOutstandingRewards,
		ValidatorAccumulatedCommissions: newValidatorAccumulatedCommissions,
		ValidatorHistoricalRewards:      newValidatorHistoricalRewards,
		ValidatorCurrentRewards:         newValidatorCurrentRewards,
		DelegatorStartingInfos:          newDelegatorStartingInfos,
		ValidatorSlashEvents:            newValidatorSlashEvents,
	}
}
