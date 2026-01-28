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
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Governance Hooks implementation
var _ govtypes.GovHooks = &GovHooks{}

// GovHooks implements governance hooks for the POA module.
type GovHooks struct {
	k *Keeper
}

// NewGovHooks creates new governance hooks
func (k *Keeper) NewGovHooks() govtypes.GovHooks {
	return &GovHooks{k: k}
}

// AfterProposalSubmission validates that only POA validators can submit proposals
func (h GovHooks) AfterProposalSubmission(ctx context.Context, _ uint64, proposerAddr sdk.AccAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return h.k.validateAuthorizedValidator(sdkCtx, proposerAddr)
}

// AfterProposalDeposit validates that only POA validators can deposit on proposals
func (h GovHooks) AfterProposalDeposit(ctx context.Context, _ uint64, depositorAddr sdk.AccAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return h.k.validateAuthorizedValidator(sdkCtx, depositorAddr)
}

// AfterProposalVote validates that only POA validators can vote
func (h GovHooks) AfterProposalVote(ctx context.Context, _ uint64, voterAddr sdk.AccAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return h.k.validateAuthorizedValidator(sdkCtx, voterAddr)
}

// AfterProposalFailedMinDeposit is called after a proposal fails to meet minimum deposit.
func (h GovHooks) AfterProposalFailedMinDeposit(_ context.Context, _ uint64) error {
	return nil
}

// AfterProposalVotingPeriodEnded is called after a proposal's voting period ends.
func (h GovHooks) AfterProposalVotingPeriodEnded(_ context.Context, _ uint64) error {
	return nil
}
