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
		{"panics when fee recipient is empty at block 1", 1, "", true},
		{"does not panic when fee recipient is poa at block 1", 1, poatypes.ModuleName, false},
		{"skips check at block 2", 2, "fee_collector", false},
		{"skips check at block 0", 0, "fee_collector", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := setupTest(t)
			ctx := f.ctx.WithBlockHeight(tc.blockHeight)
			ante.FeeRecipientModule = tc.feeRecipient

			run := func() { _, _ = f.poaKeeper.EndBlocker(ctx) }
			if tc.shouldPanic {
				require.Panics(t, run)
			} else {
				require.NotPanics(t, run)
			}
		})
	}
}
