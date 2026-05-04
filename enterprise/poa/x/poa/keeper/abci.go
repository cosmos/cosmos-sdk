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
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// EndBlocker is called at the end of every block to return pending validator updates to CometBFT.
// It retrieves all queued validator power changes and returns them as ABCI ValidatorUpdates.
// These updates take effect in the next block.
func (k *Keeper) EndBlocker(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	if ctx.BlockHeight() == 1 {
		validateFeeRecipient()
	}
	return k.ReapValidatorUpdates(ctx), nil
}

// validateFeeRecipient panics if the ante handler's fee recipient is not set
// to the POA module. When POA is active, fees must flow to the POA module
// account so it can distribute them to validators.
func validateFeeRecipient() {
	if ante.FeeRecipientModule != types.ModuleName {
		panic(fmt.Sprintf(
			"POA module requires fees to be sent to the %q module account, "+
				"but the ante handler fee recipient is set to %q. "+
				"Use DeductFeeDecorator.WithFeeRecipientModule(%q) when constructing the ante handler",
			types.ModuleName, ante.FeeRecipientModule, types.ModuleName,
		))
	}
}
