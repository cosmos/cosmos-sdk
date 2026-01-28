// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See ./enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package types

import (
	"fmt"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/core/address"
)

// ValidateBasic performs basic validation on the GenesisState.
// It ensures that:
//   - All validators pass basic validation
//   - No duplicate operator addresses exist across validators
//   - All validator pubkeys are non-nil
//   - At least one validator with non-zero power exists
//
// Note: Duplicate consensus addresses are not checked here as they require
// unpacking pubkeys with a codec. The keeper will enforce consensus address
// uniqueness when importing genesis. Parameter validation happens in the
// keeper when params are set.
func (s *GenesisState) ValidateBasic() error {
	// Validate the validator set
	if err := ValidateValidatorSet(s.Validators); err != nil {
		return err
	}

	// Ensure at least one validator with non-zero power exists
	hasNonZeroPower := false
	for _, validator := range s.Validators {
		if validator.Power > 0 {
			hasNonZeroPower = true
			break
		}
	}

	if !hasNonZeroPower {
		return fmt.Errorf("genesis must contain at least one validator with non-zero power")
	}

	return nil
}

// Validate performs full validation on the GenesisState.
// It ensures that:
//   - All basic validations pass (via ValidateBasic)
//   - Params are valid according to Params.Validate()
//
// The address codec is used to validate address formats.
func (s *GenesisState) Validate(ac address.Codec) error {
	// First run basic validation
	if err := s.ValidateBasic(); err != nil {
		return err
	}

	// Validate params with address codec
	if err := s.Params.Validate(ac); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	return nil
}

// DefaultGenesis returns a default genesis state for the POA module.
// It sets the admin address to the governance module address and initializes
// an empty list of validators.
//
// Note: The default genesis state will fail validation as it requires at least
// one validator with non-zero power. Users must provide validators when creating
// the genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     Params{Admin: authtypes.NewModuleAddress(govtypes.ModuleName).String()},
		Validators: []Validator{},
	}
}
