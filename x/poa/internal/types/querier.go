package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// query endpoints supported by the staking Querier
const (
	QueryValidators = "validators"
	QueryValidator  = "validator"
	QueryParameters = "parameters"
)

// defines the params for the following queries:
// - 'custom/staking/validator'
type QueryValidatorParams struct {
	ValidatorAddr sdk.ValAddress
}

func NewQueryValidatorParams(validatorAddr sdk.ValAddress) QueryValidatorParams {
	return QueryValidatorParams{
		ValidatorAddr: validatorAddr,
	}
}

// QueryValidatorsParams defines the params for the following queries:
// - 'custom/staking/validators'
type QueryValidatorsParams struct {
	Page, Limit int
	Status      string
}

func NewQueryValidatorsParams(page, limit int, status string) QueryValidatorsParams {
	return QueryValidatorsParams{page, limit, status}
}
