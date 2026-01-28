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

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
)

func TestGovHooks_ValidateVoter(t *testing.T) {
	t.Run("non-validator cannot vote", func(t *testing.T) {
		f := setupTest(t)

		// Create a validator
		validatorAddr, _ := createValidator(t, f, 1, 100)
		validatorAddrBech32, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Create a proposal (using validator as proposer)
		proposalID := createProposal(t, f, validatorAddrBech32)

		// Try to vote as a non-validator
		nonValidatorAddr := sdk.AccAddress("nonvalidator")
		options := govv1.WeightedVoteOptions{
			{Option: govv1.OptionYes, Weight: "1.0"},
		}

		// Attempt to submit a vote (this should fail via the hook)
		err = f.govKeeper.AddVote(f.ctx, proposalID, nonValidatorAddr, options, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not an active POA validator")

		// Verify that a validator can vote successfully
		err = f.govKeeper.AddVote(f.ctx, proposalID, validatorAddrBech32, options, "")
		require.NoError(t, err)
	})

	t.Run("address not registered as validator cannot vote", func(t *testing.T) {
		f := setupTest(t)

		// Create a validator for proposing
		validatorAddr, _ := createValidator(t, f, 1, 100)
		validatorAddrBech32, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Create a completely unregistered address (not in validatorMetadata at all)
		unregisteredAddr := sdk.AccAddress("unregistered")

		// Create a proposal (using validator as proposer)
		proposalID := createProposal(t, f, validatorAddrBech32)

		// Try to vote as an unregistered address
		options := govv1.WeightedVoteOptions{
			{Option: govv1.OptionYes, Weight: "1.0"},
		}

		err = f.govKeeper.AddVote(f.ctx, proposalID, unregisteredAddr, options, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not an active POA validator")
	})

	t.Run("validator with zero power cannot vote", func(t *testing.T) {
		f := setupTest(t)

		// Create a validator with power for proposing
		proposerAddr, _ := createValidator(t, f, 1, 100)
		proposerAddrBech32, err := sdk.AccAddressFromBech32(proposerAddr)
		require.NoError(t, err)

		// Create a validator with zero power
		validatorAddr, consAddr := createValidator(t, f, 2, 0)

		// Verify the validator has zero power
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		if err != nil {
			require.ErrorIs(t, err, collections.ErrNotFound)
		} else {
			require.Equal(t, int64(0), power)
		}

		// Create a proposal (using validator with power as proposer)
		proposalID := createProposal(t, f, proposerAddrBech32)

		// Try to vote as a validator with zero power
		validatorAddrBech32, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		options := govv1.WeightedVoteOptions{
			{Option: govv1.OptionYes, Weight: "1.0"},
		}

		err = f.govKeeper.AddVote(f.ctx, proposalID, validatorAddrBech32, options, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not an active POA validator")
	})

	t.Run("multiple validators can vote", func(t *testing.T) {
		f := setupTest(t)

		// Create multiple validators
		validator1Addr, _ := createValidator(t, f, 1, 100)
		validator2Addr, _ := createValidator(t, f, 2, 200)

		// Use validator1 as proposer
		validator1AddrBech32, err := sdk.AccAddressFromBech32(validator1Addr)
		require.NoError(t, err)

		// Create a proposal
		proposalID := createProposal(t, f, validator1AddrBech32)

		// Both validators should be able to vote
		options := govv1.WeightedVoteOptions{
			{Option: govv1.OptionYes, Weight: "1.0"},
		}

		err = f.govKeeper.AddVote(f.ctx, proposalID, validator1AddrBech32, options, "")
		require.NoError(t, err)

		validator2AddrBech32, err := sdk.AccAddressFromBech32(validator2Addr)
		require.NoError(t, err)
		err = f.govKeeper.AddVote(f.ctx, proposalID, validator2AddrBech32, options, "")
		require.NoError(t, err)
	})
}

func TestGovHooks_ValidateProposer(t *testing.T) {
	t.Run("non-validator cannot submit proposal", func(t *testing.T) {
		f := setupTest(t)

		// Create a validator (but don't use it as proposer)
		createValidator(t, f, 1, 100)

		// Try to submit proposal as a non-validator
		nonValidatorAddr := sdk.AccAddress("nonvalidator")
		_, err := f.govKeeper.SubmitProposal(f.ctx, []sdk.Msg{}, "", "Test Proposal", "Description", nonValidatorAddr, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not an active POA validator")
	})

	t.Run("validator can submit proposal", func(t *testing.T) {
		f := setupTest(t)

		// Create a validator
		validatorAddr, _ := createValidator(t, f, 1, 100)
		validatorAddrBech32, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Submit proposal as validator
		proposal, err := f.govKeeper.SubmitProposal(f.ctx, []sdk.Msg{}, "", "Test Proposal", "Description", validatorAddrBech32, false)
		require.NoError(t, err)
		require.NotNil(t, proposal)
	})

	t.Run("validator with zero power cannot submit proposal", func(t *testing.T) {
		f := setupTest(t)

		// Create a validator with zero power
		validatorAddr, _ := createValidator(t, f, 1, 0)
		validatorAddrBech32, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Try to submit proposal as validator with zero power
		_, err = f.govKeeper.SubmitProposal(f.ctx, []sdk.Msg{}, "", "Test Proposal", "Description", validatorAddrBech32, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not an active POA validator")
	})
}

func TestGovHooks_ValidateDepositor(t *testing.T) {
	t.Run("non-validator deposit validation fails", func(t *testing.T) {
		f := setupTest(t)

		// Create a validator
		createValidator(t, f, 1, 100)

		// Test the hook directly
		hooks := f.poaKeeper.NewGovHooks()
		nonValidatorAddr := sdk.AccAddress("nonvalidator")

		err := hooks.AfterProposalDeposit(f.ctx, 1, nonValidatorAddr)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not an active POA validator")
	})

	t.Run("validator deposit validation passes", func(t *testing.T) {
		f := setupTest(t)

		// Create a validator
		validatorAddr, _ := createValidator(t, f, 1, 100)
		validatorAddrBech32, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Test the hook directly
		hooks := f.poaKeeper.NewGovHooks()
		err = hooks.AfterProposalDeposit(f.ctx, 1, validatorAddrBech32)
		require.NoError(t, err)
	})

	t.Run("validator with zero power deposit validation fails", func(t *testing.T) {
		f := setupTest(t)

		// Create a validator with zero power
		validatorAddr, _ := createValidator(t, f, 1, 0)
		validatorAddrBech32, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Test the hook directly
		hooks := f.poaKeeper.NewGovHooks()
		err = hooks.AfterProposalDeposit(f.ctx, 1, validatorAddrBech32)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not an active POA validator")
	})
}
