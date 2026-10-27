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

package v2

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrate seeds the LastCommittedPower snapshot from the current validator set.
// Chains that ran POA before consensus key rotation have no snapshot, so each
// active validator's power is recorded to reflect what CometBFT already holds.
func Migrate(
	ctx context.Context,
	powerIndex *indexes.Multi[int64, sdk.ConsAddress, types.Validator],
	lastCommittedPower collections.Map[sdk.ConsAddress, int64],
) error {
	ranger := new(collections.Range[collections.Pair[int64, sdk.ConsAddress]]).Descending()
	return powerIndex.Walk(ctx, ranger, func(power int64, consAddr sdk.ConsAddress) (bool, error) {
		if power == 0 {
			// since this is a descending walk based on power, all of the rest
			// are zero power as well and therefore not active, so we stop
			// here.
			return true, nil
		}
		if err := lastCommittedPower.Set(ctx, consAddr, power); err != nil {
			return true, err
		}
		return false, nil
	})
}
