package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// querier keys
const (
	QueryParams                      = "params"
	QueryValidatorOutstandingRewards = "validator_outstanding_rewards"
	QueryValidatorCommission         = "validator_commission"
	QueryValidatorSlashes            = "validator_slashes"
	QueryDelegationRewards           = "delegation_rewards"
	QueryDelegatorTotalRewards       = "delegator_total_rewards"
	QueryDelegatorValidators         = "delegator_validators"
	QueryWithdrawAddr                = "withdraw_addr"
	QueryCommunityPool               = "community_pool"
)

// QueryValidatorOutstandingRewardsParams params for query 'custom/distr/validator_outstanding_rewards'
type QueryValidatorOutstandingRewardsParams struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
}

// NewQueryValidatorOutstandingRewardsParams creates a new instance of QueryValidatorOutstandingRewardsParams
func NewQueryValidatorOutstandingRewardsParams(validatorAddr sdk.ValAddress) QueryValidatorOutstandingRewardsParams {
	return QueryValidatorOutstandingRewardsParams{
		ValidatorAddress: validatorAddr,
	}
}

// QueryValidatorCommissionParams params for query 'custom/distr/validator_commission'
type QueryValidatorCommissionParams struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
}

// NewQueryValidatorCommissionParams creates a new instance of QueryValidatorCommissionParams
func NewQueryValidatorCommissionParams(validatorAddr sdk.ValAddress) QueryValidatorCommissionParams {
	return QueryValidatorCommissionParams{
		ValidatorAddress: validatorAddr,
	}
}

// QueryValidatorSlashesParams params for query 'custom/distr/validator_slashes'
type QueryValidatorSlashesParams struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	StartingHeight   uint64         `json:"starting_height" yaml:"starting_height"`
	EndingHeight     uint64         `json:"ending_height" yaml:"ending_height"`
}

// NewQueryValidatorSlashesParams creates a new instance of QueryValidatorSlashesParams
func NewQueryValidatorSlashesParams(validatorAddr sdk.ValAddress, startingHeight, endingHeight uint64) QueryValidatorSlashesParams {
	return QueryValidatorSlashesParams{
		ValidatorAddress: validatorAddr,
		StartingHeight:   startingHeight,
		EndingHeight:     endingHeight,
	}
}

// QueryDelegationRewardsParams are params for query 'custom/distr/delegation_rewards'
type QueryDelegationRewardsParams struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
}

// NewQueryDelegationRewardsParams creates a new instance of QueryDelegationRewardsParams
func NewQueryDelegationRewardsParams(delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) QueryDelegationRewardsParams {
	return QueryDelegationRewardsParams{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
	}
}

// QueryDelegatorParams are params for query 'custom/distr/delegator_total_rewards' and 'custom/distr/delegator_validators'
type QueryDelegatorParams struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
}

// NewQueryDelegatorParams creates a new instance of QueryDelegationRewardsParams
func NewQueryDelegatorParams(delegatorAddr sdk.AccAddress) QueryDelegatorParams {
	return QueryDelegatorParams{
		DelegatorAddress: delegatorAddr,
	}
}

// QueryDelegatorWithdrawAddrParams are params for query 'custom/distr/withdraw_addr'
type QueryDelegatorWithdrawAddrParams struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
}

// NewQueryDelegatorWithdrawAddrParams creates a new instance of QueryDelegatorWithdrawAddrParams.
func NewQueryDelegatorWithdrawAddrParams(delegatorAddr sdk.AccAddress) QueryDelegatorWithdrawAddrParams {
	return QueryDelegatorWithdrawAddrParams{DelegatorAddress: delegatorAddr}
}
