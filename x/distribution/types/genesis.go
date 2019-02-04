package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// the address for where distributions rewards are withdrawn to by default
// this struct is only used at genesis to feed in default withdraw addresses
type DelegatorWithdrawInfo struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	WithdrawAddr  sdk.AccAddress `json:"withdraw_addr"`
}

// used for import / export via genesis json
type ValidatorAccumulatedCommissionRecord struct {
	ValidatorAddr sdk.ValAddress                 `json:"validator_addr"`
	Accumulated   ValidatorAccumulatedCommission `json:"accumulated"`
}

// used for import / export via genesis json
type ValidatorHistoricalRewardsRecord struct {
	ValidatorAddr sdk.ValAddress             `json:"validator_addr"`
	Period        uint64                     `json:"period"`
	Rewards       ValidatorHistoricalRewards `json:"rewards"`
}

// used for import / export via genesis json
type ValidatorCurrentRewardsRecord struct {
	ValidatorAddr sdk.ValAddress          `json:"validator_addr"`
	Rewards       ValidatorCurrentRewards `json:"rewards"`
}

// used for import / export via genesis json
type DelegatorStartingInfoRecord struct {
	DelegatorAddr sdk.AccAddress        `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress        `json:"validator_addr"`
	StartingInfo  DelegatorStartingInfo `json:"starting_info"`
}

// used for import / export via genesis json
type ValidatorSlashEventRecord struct {
	ValidatorAddr sdk.ValAddress      `json:"validator_addr"`
	Height        uint64              `json:"height"`
	Event         ValidatorSlashEvent `json:"validator_slash_event"`
}

// GenesisState - all distribution state that must be provided at genesis
type GenesisState struct {
	FeePool                         FeePool                                `json:"fee_pool"`
	CommunityTax                    sdk.Dec                                `json:"community_tax"`
	ProposerReward                  sdk.Dec                                `json:"proposer_reward"`
	WithdrawAddrEnabled             bool                                   `json:"withdraw_addr_enabled"`
	DelegatorWithdrawInfos          []DelegatorWithdrawInfo                `json:"delegator_withdraw_infos"`
	PreviousProposer                sdk.ConsAddress                        `json:"previous_proposer"`
	OutstandingRewards              sdk.DecCoins                           `json:"outstanding_rewards"`
	ValidatorAccumulatedCommissions []ValidatorAccumulatedCommissionRecord `json:"validator_accumulated_commissions"`
	ValidatorHistoricalRewards      []ValidatorHistoricalRewardsRecord     `json:"validator_historical_rewards"`
	ValidatorCurrentRewards         []ValidatorCurrentRewardsRecord        `json:"validator_current_rewards"`
	DelegatorStartingInfos          []DelegatorStartingInfoRecord          `json:"delegator_starting_infos"`
	ValidatorSlashEvents            []ValidatorSlashEventRecord            `json:"validator_slash_events"`
}

func NewGenesisState(feePool FeePool, communityTax, proposerReward sdk.Dec,
	withdrawAddrEnabled bool, dwis []DelegatorWithdrawInfo, pp sdk.ConsAddress, r OutstandingRewards,
	acc []ValidatorAccumulatedCommissionRecord, historical []ValidatorHistoricalRewardsRecord,
	cur []ValidatorCurrentRewardsRecord, dels []DelegatorStartingInfoRecord,
	slashes []ValidatorSlashEventRecord) GenesisState {

	return GenesisState{
		FeePool:                         feePool,
		CommunityTax:                    communityTax,
		ProposerReward:                  proposerReward,
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
		ProposerReward:                  sdk.NewDecWithPrec(1, 2), // 1%
		WithdrawAddrEnabled:             true,
		DelegatorWithdrawInfos:          []DelegatorWithdrawInfo{},
		PreviousProposer:                nil,
		OutstandingRewards:              sdk.DecCoins{},
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
	if data.ProposerReward.IsNegative() {
		return fmt.Errorf("mint parameter ProposerReward should be positive, is %s",
			data.ProposerReward.String())
	}
	if (data.ProposerReward.Add(data.CommunityTax)).
		GT(sdk.OneDec()) {
		return fmt.Errorf("mint parameters ProposerReward and "+
			"CommunityTax cannot add to be greater than one, "+
			"adds to %s", data.ProposerReward.Add(data.CommunityTax).String())
	}
	return data.FeePool.ValidateGenesis()
}
