package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// the address for where distributions rewards are withdrawn to by default
// this struct is only used at genesis to feed in default withdraw addresses
type DelegatorWithdrawInfo struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	WithdrawAddress  sdk.AccAddress `json:"withdraw_address" yaml:"withdraw_address"`
}

// used for import/export via genesis json
type ValidatorOutstandingRewardsRecord struct {
	ValidatorAddress   sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	OutstandingRewards sdk.DecCoins   `json:"outstanding_rewards" yaml:"outstanding_rewards"`
}

// used for import / export via genesis json
type ValidatorAccumulatedCommissionRecord struct {
	ValidatorAddress sdk.ValAddress                 `json:"validator_address" yaml:"validator_address"`
	Accumulated      ValidatorAccumulatedCommission `json:"accumulated" yaml:"accumulated"`
}

// used for import / export via genesis json
type ValidatorHistoricalRewardsRecord struct {
	ValidatorAddress sdk.ValAddress             `json:"validator_address" yaml:"validator_address"`
	Period           uint64                     `json:"period" yaml:"period"`
	Rewards          ValidatorHistoricalRewards `json:"rewards" yaml:"rewards"`
}

// used for import / export via genesis json
type ValidatorCurrentRewardsRecord struct {
	ValidatorAddress sdk.ValAddress          `json:"validator_address" yaml:"validator_address"`
	Rewards          ValidatorCurrentRewards `json:"rewards" yaml:"rewards"`
}

// used for import / export via genesis json
type DelegatorStartingInfoRecord struct {
	DelegatorAddress sdk.AccAddress        `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress        `json:"validator_address" yaml:"validator_address"`
	StartingInfo     DelegatorStartingInfo `json:"starting_info" yaml:"starting_info"`
}

// used for import / export via genesis json
type ValidatorSlashEventRecord struct {
	ValidatorAddress sdk.ValAddress      `json:"validator_address" yaml:"validator_address"`
	Height           uint64              `json:"height" yaml:"height"`
	Period           uint64              `json:"period" yaml:"period"`
	Event            ValidatorSlashEvent `json:"validator_slash_event" yaml:"validator_slash_event"`
}

// GenesisState - all distribution state that must be provided at genesis
type GenesisState struct {
	Params                          Params                                 `json:"params" yaml:"params"`
	FeePool                         FeePool                                `json:"fee_pool" yaml:"fee_pool"`
	DelegatorWithdrawInfos          []DelegatorWithdrawInfo                `json:"delegator_withdraw_infos" yaml:"delegator_withdraw_infos"`
	PreviousProposer                sdk.ConsAddress                        `json:"previous_proposer" yaml:"previous_proposer"`
	OutstandingRewards              []ValidatorOutstandingRewardsRecord    `json:"outstanding_rewards" yaml:"outstanding_rewards"`
	ValidatorAccumulatedCommissions []ValidatorAccumulatedCommissionRecord `json:"validator_accumulated_commissions" yaml:"validator_accumulated_commissions"`
	ValidatorHistoricalRewards      []ValidatorHistoricalRewardsRecord     `json:"validator_historical_rewards" yaml:"validator_historical_rewards"`
	ValidatorCurrentRewards         []ValidatorCurrentRewardsRecord        `json:"validator_current_rewards" yaml:"validator_current_rewards"`
	DelegatorStartingInfos          []DelegatorStartingInfoRecord          `json:"delegator_starting_infos" yaml:"delegator_starting_infos"`
	ValidatorSlashEvents            []ValidatorSlashEventRecord            `json:"validator_slash_events" yaml:"validator_slash_events"`
}

func NewGenesisState(
	params Params, fp FeePool, dwis []DelegatorWithdrawInfo, pp sdk.ConsAddress, r []ValidatorOutstandingRewardsRecord,
	acc []ValidatorAccumulatedCommissionRecord, historical []ValidatorHistoricalRewardsRecord,
	cur []ValidatorCurrentRewardsRecord, dels []DelegatorStartingInfoRecord, slashes []ValidatorSlashEventRecord,
) GenesisState {

	return GenesisState{
		Params:                          params,
		FeePool:                         fp,
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
		Params:                          DefaultParams(),
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
func ValidateGenesis(gs GenesisState) error {
	if err := gs.Params.ValidateBasic(); err != nil {
		return err
	}
	return gs.FeePool.ValidateGenesis()
}
