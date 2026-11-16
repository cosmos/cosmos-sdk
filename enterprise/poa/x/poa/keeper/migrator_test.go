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

	"cosmossdk.io/math"

	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMigrate2to3(t *testing.T) {
	f := setupTest(t)
	_, consAddr1 := createValidator(t, f, 1, 1)
	_, consAddr2 := createValidator(t, f, 2, 1)

	fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1))
	require.NoError(t, f.bankKeeper.MintCoins(f.ctx, poatypes.ModuleName, fees))

	// seed the overshoot that rounding to nearest could leave behind in v2
	share, err := math.LegacyNewDecFromStr("0.500000000000000001")
	require.NoError(t, err)
	for _, consAddr := range []sdk.ConsAddress{consAddr1, consAddr2} {
		require.NoError(t, f.poaKeeper.validatorAllocatedFees.Set(f.ctx, consAddr.String(), poatypes.ValidatorFees{
			Fees: sdk.DecCoins{sdk.NewDecCoinFromDec("stake", share)},
		}))
		require.NoError(t, f.poaKeeper.adjustTotalAllocated(f.ctx, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", share)}))
	}
	require.ErrorIs(t, f.poaKeeper.checkpointAllValidators(f.ctx), poatypes.ErrInvalidAllocatedFees)

	require.NoError(t, NewMigrator(f.poaKeeper).Migrate2to3(f.ctx))

	totalAllocated, err := f.poaKeeper.getTotalAllocated(f.ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDecCoinsFromCoins(fees...), totalAllocated)
	require.NoError(t, f.poaKeeper.checkpointAllValidators(f.ctx))
}
