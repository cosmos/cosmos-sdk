// DONTCOVER
package v034

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ----------------------------------------------------------------------------
// Types and Constants
// ----------------------------------------------------------------------------

const (
	ModuleName = "distr"
)

type (
	ValidatorAccumulatedCommission = sdk.DecCoins

	DelegatorStartingInfo struct {
		PreviousPeriod uint64  `json:"previous_period"`
		Stake          sdk.Dec `json:"stake"`
		Height         uint64  `json:"height"`
	}

	DelegatorWithdrawInfo struct {
		DelegatorAddress sdk.AccAddress `json:"delegator_address"`
		WithdrawAddress  sdk.AccAddress `json:"withdraw_address"`
	}

	ValidatorOutstandingRewardsRecord struct {
		ValidatorAddress   sdk.ValAddress `json:"validator_address"`
		OutstandingRewards sdk.DecCoins   `json:"outstanding_rewards"`
	}

	ValidatorAccumulatedCommissionRecord struct {
		ValidatorAddress sdk.ValAddress                 `json:"validator_address"`
		Accumulated      ValidatorAccumulatedCommission `json:"accumulated"`
	}

	ValidatorHistoricalRewardsRecord struct {
		ValidatorAddress sdk.ValAddress             `json:"validator_address"`
		Period           uint64                     `json:"period"`
		Rewards          ValidatorHistoricalRewards `json:"rewards"`
	}

	ValidatorHistoricalRewards struct {
		CumulativeRewardRatio sdk.DecCoins `json:"cumulative_reward_ratio"`
		ReferenceCount        uint16       `json:"reference_count"`
	}

	ValidatorCurrentRewards struct {
		Rewards sdk.DecCoins `json:"rewards"`
		Period  uint64       `json:"period"`
	}

	ValidatorCurrentRewardsRecord struct {
		ValidatorAddress sdk.ValAddress          `json:"validator_address"`
		Rewards          ValidatorCurrentRewards `json:"rewards"`
	}

	DelegatorStartingInfoRecord struct {
		DelegatorAddress sdk.AccAddress        `json:"delegator_address"`
		ValidatorAddress sdk.ValAddress        `json:"validator_address"`
		StartingInfo     DelegatorStartingInfo `json:"starting_info"`
	}

	ValidatorSlashEventRecord struct {
		ValidatorAddress sdk.ValAddress      `json:"validator_address"`
		Height           uint64              `json:"height"`
		Event            ValidatorSlashEvent `json:"validator_slash_event"`
	}

	FeePool struct {
		CommunityPool sdk.DecCoins `json:"community_pool"`
	}

	ValidatorSlashEvent struct {
		ValidatorPeriod uint64  `json:"validator_period"`
		Fraction        sdk.Dec `json:"fraction"`
	}

	GenesisState struct {
		FeePool                         FeePool                                `json:"fee_pool"`
		CommunityTax                    sdk.Dec                                `json:"community_tax"`
		BaseProposerReward              sdk.Dec                                `json:"base_proposer_reward"`
		BonusProposerReward             sdk.Dec                                `json:"bonus_proposer_reward"`
		WithdrawAddrEnabled             bool                                   `json:"withdraw_addr_enabled"`
		DelegatorWithdrawInfos          []DelegatorWithdrawInfo                `json:"delegator_withdraw_infos"`
		PreviousProposer                sdk.ConsAddress                        `json:"previous_proposer"`
		OutstandingRewards              []ValidatorOutstandingRewardsRecord    `json:"outstanding_rewards"`
		ValidatorAccumulatedCommissions []ValidatorAccumulatedCommissionRecord `json:"validator_accumulated_commissions"`
		ValidatorHistoricalRewards      []ValidatorHistoricalRewardsRecord     `json:"validator_historical_rewards"`
		ValidatorCurrentRewards         []ValidatorCurrentRewardsRecord        `json:"validator_current_rewards"`
		DelegatorStartingInfos          []DelegatorStartingInfoRecord          `json:"delegator_starting_infos"`
		ValidatorSlashEvents            []ValidatorSlashEventRecord            `json:"validator_slash_events"`
	}
)
