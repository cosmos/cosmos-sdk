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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package types

import (
	"fmt"
	"math"

	"cosmossdk.io/core/address"
	sdkerrors "cosmossdk.io/errors"
)

// MaxValidatorPower is the maximum allowed validator power (CometBFT uses int64 for voting power).
const MaxValidatorPower = math.MaxInt64

// MinPubKeyLength is the minimum expected pubkey serialized length (ed25519 ~34 bytes, secp256k1 ~35 bytes with proto overhead).
const MinPubKeyLength = 30

// MaxPubKeyLength is the maximum allowed pubkey serialized length (includes proto overhead).
const MaxPubKeyLength = 128

// ValidateBasic performs basic validation on a Validator.
// It ensures that:
//   - Power is non-negative and does not overflow
//   - Metadata passes validation (operator address, moniker, and description are valid)
//   - PubKey is not nil and has valid length
func (v *Validator) ValidateBasic() error {
	if v.Power < 0 {
		return ErrNegativeValidatorPower
	}
	if v.Power > MaxValidatorPower {
		return ErrValidatorPowerOverflow
	}

	if err := v.Metadata.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(ErrInvalidMetadata, err.Error())
	}

	// Check that pubkey is not nil
	if v.PubKey == nil {
		return fmt.Errorf("validator pubkey cannot be nil")
	}
	// Validate pubkey length (Value contains serialized pubkey bytes)
	if len(v.PubKey.Value) < MinPubKeyLength {
		return sdkerrors.Wrapf(ErrInvalidPubKeyLength, "pubkey length %d is below minimum %d", len(v.PubKey.Value), MinPubKeyLength)
	}
	if len(v.PubKey.Value) > MaxPubKeyLength {
		return sdkerrors.Wrapf(ErrInvalidPubKeyLength, "pubkey length %d exceeds maximum %d", len(v.PubKey.Value), MaxPubKeyLength)
	}

	return nil
}

// ValidateValidatorSet validates a set of validators.
// It ensures that:
//   - All validators pass basic validation
//   - No duplicate operator addresses exist across validators
//   - Total power does not overflow int64
//   - Total power is greater than zero (at least one validator with non-zero power)
//
// Returns an error with the validator index if validation fails.
func ValidateValidatorSet(vs []Validator) error {
	// Track operator addresses to detect duplicates
	operatorAddresses := make(map[string]struct{})
	var totalPower int64

	// Validate each validator
	for i, validator := range vs {
		// Validate basic validator fields
		if err := validator.ValidateBasic(); err != nil {
			return fmt.Errorf("validator at index %d: %w", i, err)
		}

		// Check for duplicate operator addresses
		operatorAddr := validator.Metadata.OperatorAddress
		if _, found := operatorAddresses[operatorAddr]; found {
			return fmt.Errorf("duplicate operator address %s found in validators", operatorAddr)
		}
		operatorAddresses[operatorAddr] = struct{}{}

		// Check for total power overflow when summing
		if totalPower > math.MaxInt64-validator.Power {
			return ErrTotalPowerOverflow
		}
		totalPower += validator.Power
	}

	if totalPower <= 0 {
		return sdkerrors.Wrap(ErrInvalidTotalPower, "total power must be greater than zero")
	}

	return nil
}

// ValidateBasic performs basic validation on ValidatorMetadata.
// It ensures that:
//   - OperatorAddress is not empty
//   - Moniker is not empty and does not exceed 256 characters
//   - Description is not empty and does not exceed 256 characters
func (m *ValidatorMetadata) ValidateBasic() error {
	if m.OperatorAddress == "" {
		return ErrMissingOperatorAddress
	}

	if len(m.Moniker) > 256 {
		return sdkerrors.Wrap(ErrInvalidMetadata, "moniker too long") // todo: err
	}

	if len(m.Moniker) == 0 {
		return sdkerrors.Wrap(ErrInvalidMetadata, "moniker cannot be empty")
	}

	if len(m.Description) > 256 {
		return sdkerrors.Wrap(ErrInvalidMetadata, "description too long") // todo: err
	}

	return nil
}

// Validate performs validation on Params.
// It ensures that:
//   - Admin address is a valid address format according to the address codec
//
// The address codec is used to validate the admin address format.
func (p *Params) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(p.Admin); err != nil {
		return sdkerrors.Wrap(ErrInvalidAdminAddress, err.Error())
	}

	return nil
}

// Validate performs validation on MsgUpdateParams.
// It ensures that:
//   - Admin (signer) is a valid address format
//   - Params are valid according to Params.Validate()
//
// The address codec is used to validate address formats.
func (m *MsgUpdateParams) Validate(ac address.Codec) error {
	if _, err := ac.StringToBytes(m.Admin); err != nil {
		return sdkerrors.Wrap(ErrInvalidAdminAddress, "invalid signer address: "+err.Error())
	}

	if err := m.Params.Validate(ac); err != nil {
		return err
	}

	return nil
}

// ValidateBasic performs basic validation on MsgUpdateValidators.
// It delegates to ValidateValidatorSet() to validate the validators list.
// This ensures that:
//   - All validators pass basic validation
//   - No duplicate operator addresses exist across validators
//
// Returns an error with the validator index if validation fails.
func (m *MsgUpdateValidators) ValidateBasic() error {
	return ValidateValidatorSet(m.Validators)
}

// Validate performs validation on MsgCreateValidator.
// It ensures that:
//   - Metadata passes basic validation (operator address, moniker, and description are valid)
//   - Operator address is a valid address format according to the address codec
//   - PubKey is not nil
//
// The address codec is used to validate the operator address format.
func (m *MsgCreateValidator) Validate(ac address.Codec) error {
	md := ValidatorMetadata{
		Moniker:         m.Moniker,
		Description:     m.Description,
		OperatorAddress: m.OperatorAddress,
	}
	if err := md.ValidateBasic(); err != nil {
		return err
	}

	if _, err := ac.StringToBytes(m.OperatorAddress); err != nil {
		return sdkerrors.Wrap(ErrInvalidMetadata, "operator address is invalid")
	}

	// Check that pubkey is not nil
	if m.PubKey == nil {
		return fmt.Errorf("validator pubkey cannot be nil")
	}
	// Validate pubkey length
	if len(m.PubKey.Value) < MinPubKeyLength {
		return sdkerrors.Wrapf(ErrInvalidPubKeyLength, "pubkey length %d is below minimum %d", len(m.PubKey.Value), MinPubKeyLength)
	}
	if len(m.PubKey.Value) > MaxPubKeyLength {
		return sdkerrors.Wrapf(ErrInvalidPubKeyLength, "pubkey length %d exceeds maximum %d", len(m.PubKey.Value), MaxPubKeyLength)
	}

	return nil
}

// ValidateBasic performs basic validation on MsgWithdrawFees.
// It ensures that the Operator field is not empty.
func (m *MsgWithdrawFees) ValidateBasic() error {
	if m.Operator == "" {
		return ErrMissingOperatorAddress
	}
	return nil
}
