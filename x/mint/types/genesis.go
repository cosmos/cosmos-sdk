package types

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/math"
)

type MintFn func(ctx context.Context, env appmodule.Environment, minter *Minter) error

// DefaultInflationCalculationFn is the default function used to calculate inflation.
func DefaultInflationCalculationFn(_ context.Context, minter Minter, params Params, bondedRatio math.LegacyDec) math.LegacyDec {
	return minter.NextInflationRate(params, bondedRatio)
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter Minter, params Params) *GenesisState {
	return &GenesisState{
		Minter: minter,
		Params: params,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Minter: DefaultInitialMinter(),
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return ValidateMinter(data.Minter)
}
