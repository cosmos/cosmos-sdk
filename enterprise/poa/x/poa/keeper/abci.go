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

package keeper

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker is called at the end of every block to return pending validator updates to CometBFT.
// It retrieves all queued validator power changes and returns them as ABCI ValidatorUpdates.
// These updates take effect in the next block.
func (k *Keeper) EndBlocker(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	if ctx.BlockHeight() == 1 {
		k.validateFeeRecipient()
	}

	updates, err := k.ReapValidatorUpdates(ctx)
	if err != nil {
		return nil, err
	}
	if err := k.SetLastCommittedPower(ctx, updates); err != nil {
		return nil, err
	}

	return updates, nil
}

// validateFeeRecipient panics if fees are not sent to the POA module. When POA
// is active, fees must flow to the POA module account so it can distribute them
// to validators. The app must hand its ante.FeeRecipients registry to the keeper
// with SetFeeRecipients, and every DeductFeeDecorator installed through it must
// pay POA (at least one must be installed). It panics if the registry was never
// set, since the check could not see the app's decorators.
func (k *Keeper) validateFeeRecipient() {
	if k.feeRecipients == nil {
		panic(fmt.Sprintf(
			"POA module requires fees to be sent to the %q module account, "+
				"but no FeeRecipients registry was set. "+
				"Pass the ante.HandlerOptions.FeeRecipients registry to Keeper.SetFeeRecipients",
			types.ModuleName,
		))
	}
	validateFeeRecipients(k.feeRecipients.Modules())
}

func validateFeeRecipients(installed []string) {
	if len(installed) == 0 {
		panic(fmt.Sprintf(
			"POA module requires fees to be sent to the %q module account, "+
				"but no DeductFeeDecorator was installed through the FeeRecipients registry",
			types.ModuleName,
		))
	}
	for _, m := range installed {
		if m != types.ModuleName {
			panic(fmt.Sprintf(
				"POA module requires fees to be sent to the %q module account, "+
					"but an installed DeductFeeDecorator sends fees to %q. "+
					"Use DeductFeeDecorator.WithFeeRecipientModule(%q) on every installed decorator",
				types.ModuleName, m, types.ModuleName,
			))
		}
	}
}
