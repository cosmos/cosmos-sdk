// DONTCOVER
// nolint
package v0_36

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_34"
)

// ----------------------------------------------------------------------------
// Types and Constants
// ----------------------------------------------------------------------------

const (
	ModuleName = "distribution"
)

type (
	ValidatorAccumulatedCommission = sdk.DecCoins

	ValidatorSlashEventRecord struct {
		ValidatorAddress sdk.ValAddress                `json:"validator_address"`
		Height           uint64                        `json:"height"`
		Period           uint64                        `json:"period"`
		Event            v034distr.ValidatorSlashEvent `json:"validator_slash_event"`
	}

	GenesisState struct {
		FeePool                         v034distr.FeePool                                `json:"fee_pool"`
		CommunityTax                    sdk.Dec                                          `json:"community_tax"`
		BaseProposerReward              sdk.Dec                                          `json:"base_proposer_reward"`
		BonusProposerReward             sdk.Dec                                          `json:"bonus_proposer_reward"`
		WithdrawAddrEnabled             bool                                             `json:"withdraw_addr_enabled"`
		DelegatorWithdrawInfos          []v034distr.DelegatorWithdrawInfo                `json:"delegator_withdraw_infos"`
		PreviousProposer                sdk.ConsAddress                                  `json:"previous_proposer"`
		OutstandingRewards              []v034distr.ValidatorOutstandingRewardsRecord    `json:"outstanding_rewards"`
		ValidatorAccumulatedCommissions []v034distr.ValidatorAccumulatedCommissionRecord `json:"validator_accumulated_commissions"`
		ValidatorHistoricalRewards      []v034distr.ValidatorHistoricalRewardsRecord     `json:"validator_historical_rewards"`
		ValidatorCurrentRewards         []v034distr.ValidatorCurrentRewardsRecord        `json:"validator_current_rewards"`
		DelegatorStartingInfos          []v034distr.DelegatorStartingInfoRecord          `json:"delegator_starting_infos"`
		ValidatorSlashEvents            []ValidatorSlashEventRecord                      `json:"validator_slash_events"`
	}
)

func NewGenesisState(
	feePool v034distr.FeePool, communityTax, baseProposerReward, bonusProposerReward sdk.Dec,
	withdrawAddrEnabled bool, dwis []v034distr.DelegatorWithdrawInfo, pp sdk.ConsAddress,
	r []v034distr.ValidatorOutstandingRewardsRecord, acc []v034distr.ValidatorAccumulatedCommissionRecord,
	historical []v034distr.ValidatorHistoricalRewardsRecord, cur []v034distr.ValidatorCurrentRewardsRecord,
	dels []v034distr.DelegatorStartingInfoRecord, slashes []ValidatorSlashEventRecord,
) GenesisState {

	return GenesisState{
		FeePool:                         feePool,
		CommunityTax:                    communityTax,
		BaseProposerReward:              baseProposerReward,
		BonusProposerReward:             bonusProposerReward,
		WithdrawAddrEnabled:             withdrawAddrEnabled,
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
