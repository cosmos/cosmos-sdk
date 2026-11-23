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
	"testing"

	"github.com/stretchr/testify/require"

	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

func TestEndBlockerFeeRecipientValidation(t *testing.T) {
	tests := []struct {
		name         string
		blockHeight  int64
		feeRecipient string
		shouldPanic  bool
	}{
		{"panics when fee recipient is fee_collector at block 1", 1, "fee_collector", true},
		{"does not panic when fee recipient is poa at block 1", 1, poatypes.ModuleName, false},
		{"skips check at block 2", 2, "fee_collector", false},
		{"skips check at block 0", 0, "fee_collector", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := setupTest(t)
			ctx := f.ctx.WithBlockHeight(tc.blockHeight)

			registry := ante.NewFeeRecipients()
			registry.Install(ante.NewDeductFeeDecorator(nil, nil, nil, nil).WithFeeRecipientModule(tc.feeRecipient))
			f.poaKeeper.SetFeeRecipients(registry)

			run := func() { _, _ = f.poaKeeper.EndBlocker(ctx) }
			if tc.shouldPanic {
				require.Panics(t, run)
			} else {
				require.NotPanics(t, run)
			}
		})
	}
}

// TestValidateFeeRecipients covers an app that never calls SetFeeRecipients:
// PoA cannot see its decorators, so a plain decorator paying fee_collector
// would go unnoticed. EndBlocker must panic at block 1 rather than pass.
func TestValidateFeeRecipients(t *testing.T) {
	tests := []struct {
		name        string
		blockHeight int64
		// poaThenPlain builds a poa-configured decorator and then a plain one,
		// as an app that never wires the registry would.
		poaThenPlain bool
		setRegistry  bool
		shouldPanic  bool
	}{
		{"no registry, poa then plain decorator, panics at block 1", 1, true, false, true},
		{"no registry, poa decorator only, panics at block 1", 1, false, false, true},
		{"no registry, check skipped at block 2", 2, true, false, false},
		{"registry with only poa decorator passes", 1, false, true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := setupTest(t)
			ctx := f.ctx.WithBlockHeight(tc.blockHeight)

			var registry *ante.FeeRecipients
			if tc.setRegistry {
				registry = ante.NewFeeRecipients()
				f.poaKeeper.SetFeeRecipients(registry)
			}
			poaDfd := ante.NewDeductFeeDecorator(nil, nil, nil, nil).WithFeeRecipientModule(poatypes.ModuleName)
			registry.Install(poaDfd) // no-op on a nil registry
			if tc.poaThenPlain {
				_ = ante.NewDeductFeeDecorator(nil, nil, nil, nil)
			}

			run := func() { _, _ = f.poaKeeper.EndBlocker(ctx) }
			if tc.shouldPanic {
				require.Panics(t, run)
			} else {
				require.NotPanics(t, run)
			}
		})
	}
}

func TestEndBlockerFeeRecipientRegistry(t *testing.T) {
	// Decorators built but never installed (e.g. a default ante handler that
	// the app replaced, or another app in the same process) do not count.
	f := setupTest(t)
	ctx := f.ctx.WithBlockHeight(1)

	registry := ante.NewFeeRecipients()
	f.poaKeeper.SetFeeRecipients(registry)
	_ = ante.NewDeductFeeDecorator(nil, nil, nil, nil)
	registry.Install(ante.NewDeductFeeDecorator(nil, nil, nil, nil).WithFeeRecipientModule(poatypes.ModuleName))
	require.NotPanics(t, func() { _, _ = f.poaKeeper.EndBlocker(ctx) })

	// A copy redirected to poa does not redirect the original: installing both
	// leaves fee_collector in the registry, so PoA rejects it.
	f = setupTest(t)
	ctx = f.ctx.WithBlockHeight(1)
	registry = ante.NewFeeRecipients()
	f.poaKeeper.SetFeeRecipients(registry)
	base := ante.NewDeductFeeDecorator(nil, nil, nil, nil)
	redirected := base.WithFeeRecipientModule(poatypes.ModuleName)
	registry.Install(redirected)
	registry.Install(base)
	require.Panics(t, func() { _, _ = f.poaKeeper.EndBlocker(ctx) })

	// An empty registry is rejected too.
	f = setupTest(t)
	ctx = f.ctx.WithBlockHeight(1)
	f.poaKeeper.SetFeeRecipients(ante.NewFeeRecipients())
	require.Panics(t, func() { _, _ = f.poaKeeper.EndBlocker(ctx) })
}
