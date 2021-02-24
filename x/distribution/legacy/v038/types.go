package v038

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v034"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v036"
)

// DONTCOVER

const (
	ModuleName = "distribution"
)

type (
	GenesisState struct {
		Params                          Params                                           `json:"params" yaml:"params"`
		FeePool                         v034distr.FeePool                                `json:"fee_pool"`
		DelegatorWithdrawInfos          []v034distr.DelegatorWithdrawInfo                `json:"delegator_withdraw_infos"`
		PreviousProposer                sdk.ConsAddress                                  `json:"previous_proposer" yaml:"previous_proposer"`
		OutstandingRewards              []v034distr.ValidatorOutstandingRewardsRecord    `json:"outstanding_rewards"`
		ValidatorAccumulatedCommissions []v034distr.ValidatorAccumulatedCommissionRecord `json:"validator_accumulated_commissions"`
		ValidatorHistoricalRewards      []v034distr.ValidatorHistoricalRewardsRecord     `json:"validator_historical_rewards"`
		ValidatorCurrentRewards         []v034distr.ValidatorCurrentRewardsRecord        `json:"validator_current_rewards"`
		DelegatorStartingInfos          []v034distr.DelegatorStartingInfoRecord          `json:"delegator_starting_infos"`
		ValidatorSlashEvents            []v036distr.ValidatorSlashEventRecord            `json:"validator_slash_events" yaml:"validator_slash_events"`
	}

	Params struct {
		CommunityTax        sdk.Dec `json:"community_tax" yaml:"community_tax"`
		BaseProposerReward  sdk.Dec `json:"base_proposer_reward" yaml:"base_proposer_reward"`
		BonusProposerReward sdk.Dec `json:"bonus_proposer_reward" yaml:"bonus_proposer_reward"`
		WithdrawAddrEnabled bool    `json:"withdraw_addr_enabled" yaml:"withdraw_addr_enabled"`
	}
)

func NewGenesisState(
	params Params, feePool v034distr.FeePool, dwis []v034distr.DelegatorWithdrawInfo, pp sdk.ConsAddress,
	r []v034distr.ValidatorOutstandingRewardsRecord, acc []v034distr.ValidatorAccumulatedCommissionRecord,
	historical []v034distr.ValidatorHistoricalRewardsRecord, cur []v034distr.ValidatorCurrentRewardsRecord,
	dels []v034distr.DelegatorStartingInfoRecord, slashes []v036distr.ValidatorSlashEventRecord,
) GenesisState {

	return GenesisState{
		FeePool:                         feePool,
		Params:                          params,
		DelegatorWithdrawInfos:          dwis,
		PreviousProposer:                pp,
		OutstandingRewards:              r,
		ValidatorAccumulatedCommissions: acc,
		ValidatorHistoricalRewards:      historical,
		ValidatorCurrentRewards:         cur,
		DelegatorStartingInfos:          dels,
		ValidatorSlashEvents:            slashes,
	}
}
