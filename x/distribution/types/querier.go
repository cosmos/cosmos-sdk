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

type (

	// params for query 'custom/distr/withdraw_addr'
	QueryDelegatorWithdrawAddrParams struct {
		DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	}

	// params for query 'custom/distr/validator_outstanding_rewards'
	QueryValidatorOutstandingRewardsParams struct {
		ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	}

	// params for query 'custom/distr/validator_commission'
	QueryValidatorCommissionParams struct {
		ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	}

	// params for query 'custom/distr/validator_slashes'
	QueryValidatorSlashesParams struct {
		ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
		StartingHeight   uint64         `json:"starting_height" yaml:"starting_height"`
		EndingHeight     uint64         `json:"ending_height" yaml:"ending_height"`
	}

	// params for query 'custom/distr/delegation_rewards'
	QueryDelegationRewardsParams struct {
		DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
		ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
	}

	// params for query 'custom/distr/delegator_total_rewards' and 'custom/distr/delegator_validators'
	QueryDelegatorParams struct {
		DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
	}
)

// creates a new instance of QueryValidatorOutstandingRewardsParams
func NewQueryValidatorOutstandingRewardsParams(validatorAddr sdk.ValAddress) QueryValidatorOutstandingRewardsParams {
	return QueryValidatorOutstandingRewardsParams{
		ValidatorAddress: validatorAddr,
	}
}

// creates a new instance of QueryValidatorCommissionParams
func NewQueryValidatorCommissionParams(validatorAddr sdk.ValAddress) QueryValidatorCommissionParams {
	return QueryValidatorCommissionParams{
		ValidatorAddress: validatorAddr,
	}
}

// creates a new instance of QueryValidatorSlashesParams
func NewQueryValidatorSlashesParams(validatorAddr sdk.ValAddress, startingHeight, endingHeight uint64) QueryValidatorSlashesParams {
	return QueryValidatorSlashesParams{
		ValidatorAddress: validatorAddr,
		StartingHeight:   startingHeight,
		EndingHeight:     endingHeight,
	}
}

// creates a new instance of QueryDelegationRewardsParams
func NewQueryDelegationRewardsParams(delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) QueryDelegationRewardsParams {
	return QueryDelegationRewardsParams{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
	}
}

// creates a new instance of QueryDelegationRewardsParams
func NewQueryDelegatorParams(delegatorAddr sdk.AccAddress) QueryDelegatorParams {
	return QueryDelegatorParams{
		DelegatorAddress: delegatorAddr,
	}
}

// NewQueryDelegatorWithdrawAddrParams creates a new instance of QueryDelegatorWithdrawAddrParams.
func NewQueryDelegatorWithdrawAddrParams(delegatorAddr sdk.AccAddress) QueryDelegatorWithdrawAddrParams {
	return QueryDelegatorWithdrawAddrParams{DelegatorAddress: delegatorAddr}
}
