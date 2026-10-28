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

package v3

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrate removes allocated fees that exceed the module balance. Before v3,
// per-validator shares were rounded to nearest, so their sum could exceed the
// fees being split, leaving more allocated than the module holds.
//
// For each denom where the per-validator entries sum to more than the balance,
// the excess is taken from the entries in key order until it is gone, and
// totalAllocatedFees is set to the new sum of the entries.
//
// InitGenesis also applies it to imported state, so exports from poa v1.x
// load without the excess.
func Migrate(
	ctx context.Context,
	balance sdk.DecCoins,
	validatorAllocatedFees collections.Map[string, types.ValidatorFees],
	totalAllocatedFees collections.Item[types.ValidatorFees],
) error {
	type entry struct {
		consAddr string
		fees     sdk.DecCoins
	}
	var entries []entry
	allocated := sdk.DecCoins{}
	if err := validatorAllocatedFees.Walk(ctx, nil, func(consAddr string, fees types.ValidatorFees) (bool, error) {
		entries = append(entries, entry{consAddr: consAddr, fees: fees.Fees})
		allocated = allocated.Add(fees.Fees...)
		return false, nil
	}); err != nil {
		return err
	}

	newTotal := allocated.Intersect(balance)
	// excess holds, per denom, how far the entries exceed the balance
	excess := allocated.Sub(newTotal)
	if excess.IsZero() {
		return nil
	}

	for _, e := range entries {
		if excess.IsZero() {
			break
		}

		trim := excess.Intersect(e.fees)
		if trim.IsZero() {
			continue
		}
		excess = excess.Sub(trim)

		if err := validatorAllocatedFees.Set(ctx, e.consAddr, types.ValidatorFees{Fees: e.fees.Sub(trim)}); err != nil {
			return err
		}
	}

	return totalAllocatedFees.Set(ctx, types.ValidatorFees{Fees: newTotal})
}
