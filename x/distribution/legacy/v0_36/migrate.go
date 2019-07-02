package v0_36

import (
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_34"
)

func Migrate(oldGenState v034distr.GenesisState) GenesisState {
	// migrate slash events which now have the period included
	slashEvents := make([]ValidatorSlashEventRecord, len(oldGenState.ValidatorSlashEvents), len(oldGenState.ValidatorSlashEvents))
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
		oldGenState.FeePool, oldGenState.CommunityTax, oldGenState.BaseProposerReward, oldGenState.BonusProposerReward,
		oldGenState.WithdrawAddrEnabled, oldGenState.DelegatorWithdrawInfos, oldGenState.PreviousProposer,
		oldGenState.OutstandingRewards, oldGenState.ValidatorAccumulatedCommissions,
		oldGenState.ValidatorHistoricalRewards, oldGenState.ValidatorCurrentRewards,
		oldGenState.DelegatorStartingInfos, slashEvents,
	)
}
