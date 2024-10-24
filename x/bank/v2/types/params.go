package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new parameter configuration for the bank/v2 module
func NewParams(denomCreationFee sdk.Coins, gasConsume uint64) Params {
	return Params{
		DenomCreationFee:        denomCreationFee,
		DenomCreationGasConsume: gasConsume,
	}
}

// DefaultParams is the default parameter configuration for the bank/v2 module
func DefaultParams() Params {
	return NewParams(sdk.NewCoins(), 1_000_000)
}

// Validate all bank/v2 module parameters
func (p Params) Validate() error {
	if err := validateDenomCreationFee(p.DenomCreationFee); err != nil {
		return err
	}

	if err := validateDenomCreationGasConsume(p.DenomCreationGasConsume); err != nil {
		return err
	}

	return nil
}

func validateDenomCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid denom creation fee: %+v", i)
	}

	return nil
}

func validateDenomCreationGasConsume(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
