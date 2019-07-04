package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// the address for where distributions rewards are withdrawn to by default
// this struct is only used at genesis to feed in default withdraw addresses
type DelegatorWithdrawInfo struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address"`
	WithdrawAddress  sdk.AccAddress `json:"withdraw_address"`
}

// used for import/export via genesis json
type ValidatorOutstandingRewardsRecord struct {
	ValidatorAddress   sdk.ValAddress `json:"validator_address"`
	OutstandingRewards sdk.DecCoins   `json:"outstanding_rewards"`
}

// used for import / export via genesis json
type ValidatorAccumulatedCommissionRecord struct {
	ValidatorAddress sdk.ValAddress                 `json:"validator_address"`
	Accumulated      ValidatorAccumulatedCommission `json:"accumulated"`
}

// used for import / export via genesis json
type ValidatorHistoricalRewardsRecord struct {
	ValidatorAddress sdk.ValAddress             `json:"validator_address"`
	Period           uint64                     `json:"period"`
	Rewards          ValidatorHistoricalRewards `json:"rewards"`
}

// used for import / export via genesis json
type ValidatorCurrentRewardsRecord struct {
	ValidatorAddress sdk.ValAddress          `json:"validator_address"`
	Rewards          ValidatorCurrentRewards `json:"rewards"`
}

// used for import / export via genesis json
type DelegatorStartingInfoRecord struct {
	DelegatorAddress sdk.AccAddress        `json:"delegator_address"`
	ValidatorAddress sdk.ValAddress        `json:"validator_address"`
	StartingInfo     DelegatorStartingInfo `json:"starting_info"`
}

// used for import / export via genesis json
type ValidatorSlashEventRecord struct {
	ValidatorAddress sdk.ValAddress      `json:"validator_address"`
	Height           uint64              `json:"height"`
	Period           uint64              `json:"period"`
	Event            ValidatorSlashEvent `json:"validator_slash_event"`
}

// GenesisState - all distribution state that must be provided at genesis
type GenesisState struct {
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

func NewGenesisState(feePool FeePool, communityTax, baseProposerReward, bonusProposerReward sdk.Dec,
	withdrawAddrEnabled bool, dwis []DelegatorWithdrawInfo, pp sdk.ConsAddress, r []ValidatorOutstandingRewardsRecord,
	acc []ValidatorAccumulatedCommissionRecord, historical []ValidatorHistoricalRewardsRecord,
	cur []ValidatorCurrentRewardsRecord, dels []DelegatorStartingInfoRecord,
	slashes []ValidatorSlashEventRecord) GenesisState {

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

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		FeePool:                         InitialFeePool(),
		CommunityTax:                    sdk.NewDecWithPrec(2, 2), // 2%
		BaseProposerReward:              sdk.NewDecWithPrec(1, 2), // 1%
		BonusProposerReward:             sdk.NewDecWithPrec(4, 2), // 4%
		WithdrawAddrEnabled:             true,
		DelegatorWithdrawInfos:          []DelegatorWithdrawInfo{},
		PreviousProposer:                nil,
		OutstandingRewards:              []ValidatorOutstandingRewardsRecord{},
		ValidatorAccumulatedCommissions: []ValidatorAccumulatedCommissionRecord{},
		ValidatorHistoricalRewards:      []ValidatorHistoricalRewardsRecord{},
		ValidatorCurrentRewards:         []ValidatorCurrentRewardsRecord{},
		DelegatorStartingInfos:          []DelegatorStartingInfoRecord{},
		ValidatorSlashEvents:            []ValidatorSlashEventRecord{},
	}
}

// ValidateGenesis validates the genesis state of distribution genesis input
func ValidateGenesis(data GenesisState) error {
	if data.CommunityTax.IsNegative() || data.CommunityTax.GT(sdk.OneDec()) {
		return fmt.Errorf("mint parameter CommunityTax should non-negative and "+
			"less than one, is %s", data.CommunityTax.String())
	}
	if data.BaseProposerReward.IsNegative() {
		return fmt.Errorf("mint parameter BaseProposerReward should be positive, is %s",
			data.BaseProposerReward.String())
	}
	if data.BonusProposerReward.IsNegative() {
		return fmt.Errorf("mint parameter BonusProposerReward should be positive, is %s",
			data.BonusProposerReward.String())
	}
	if (data.BaseProposerReward.Add(data.BonusProposerReward)).
		GT(sdk.OneDec()) {
		return fmt.Errorf("mint parameters BaseProposerReward and "+
			"BonusProposerReward cannot add to be greater than one, "+
			"adds to %s", data.BaseProposerReward.Add(data.BonusProposerReward).String())
	}
	return data.FeePool.ValidateGenesis()
}
