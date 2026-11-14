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

package v3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/math"

	v3 "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/migrations/v3"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupStores(t *testing.T) (context.Context, collections.Map[string, poatypes.ValidatorFees], collections.Item[poatypes.ValidatorFees]) {
	t.Helper()
	kv, ctx := colltest.MockStore()
	sb := collections.NewSchemaBuilder(kv)
	validatorAllocatedFees := collections.NewMap(sb, poatypes.ValidatorAllocatedFeesKey, "validator_allocated_fees", collections.StringKey, colltest.MockValueCodec[poatypes.ValidatorFees]())
	totalAllocatedFees := collections.NewItem(sb, poatypes.TotalAllocatedKey, "total_allocated", colltest.MockValueCodec[poatypes.ValidatorFees]())
	return ctx, validatorAllocatedFees, totalAllocatedFees
}

func decCoin(t *testing.T, denom, amount string) sdk.DecCoin {
	t.Helper()
	dec, err := math.LegacyNewDecFromStr(amount)
	require.NoError(t, err)
	return sdk.NewDecCoinFromDec(denom, dec)
}

func seed(t *testing.T, ctx context.Context, validatorAllocatedFees collections.Map[string, poatypes.ValidatorFees], totalAllocatedFees collections.Item[poatypes.ValidatorFees], entries map[string]sdk.DecCoins) {
	t.Helper()
	total := sdk.DecCoins{}
	for consAddr, fees := range entries {
		require.NoError(t, validatorAllocatedFees.Set(ctx, consAddr, poatypes.ValidatorFees{Fees: fees}))
		total = total.Add(fees...)
	}
	require.NoError(t, totalAllocatedFees.Set(ctx, poatypes.ValidatorFees{Fees: total}))
}

func TestMigrate(t *testing.T) {
	t.Run("leaves allocations within the balance untouched", func(t *testing.T) {
		ctx, validatorAllocatedFees, totalAllocatedFees := setupStores(t)
		entries := map[string]sdk.DecCoins{
			"cons-a": sdk.NewDecCoins(decCoin(t, "stake", "0.5")),
			"cons-b": sdk.NewDecCoins(decCoin(t, "stake", "0.5")),
		}
		seed(t, ctx, validatorAllocatedFees, totalAllocatedFees, entries)

		balance := sdk.NewDecCoins(decCoin(t, "stake", "1"))
		require.NoError(t, v3.Migrate(ctx, balance, validatorAllocatedFees, totalAllocatedFees))

		for consAddr, fees := range entries {
			got, err := validatorAllocatedFees.Get(ctx, consAddr)
			require.NoError(t, err)
			require.Equal(t, fees, got.Fees)
		}
		total, err := totalAllocatedFees.Get(ctx)
		require.NoError(t, err)
		require.Equal(t, balance, total.Fees)
	})

	t.Run("trims the excess from entries in key order", func(t *testing.T) {
		ctx, validatorAllocatedFees, totalAllocatedFees := setupStores(t)
		// sums to 1.000000000000000003 against a balance of 1, and cons-a alone
		// is smaller than the excess
		seed(t, ctx, validatorAllocatedFees, totalAllocatedFees, map[string]sdk.DecCoins{
			"cons-a": sdk.NewDecCoins(decCoin(t, "stake", "0.000000000000000001")),
			"cons-b": sdk.NewDecCoins(decCoin(t, "stake", "0.5")),
			"cons-c": sdk.NewDecCoins(decCoin(t, "stake", "0.500000000000000002")),
		})

		balance := sdk.NewDecCoins(decCoin(t, "stake", "1"))
		require.NoError(t, v3.Migrate(ctx, balance, validatorAllocatedFees, totalAllocatedFees))

		gotA, err := validatorAllocatedFees.Get(ctx, "cons-a")
		require.NoError(t, err)
		require.True(t, gotA.Fees.IsZero())

		gotB, err := validatorAllocatedFees.Get(ctx, "cons-b")
		require.NoError(t, err)
		require.Equal(t, sdk.NewDecCoins(decCoin(t, "stake", "0.499999999999999998")), gotB.Fees)

		gotC, err := validatorAllocatedFees.Get(ctx, "cons-c")
		require.NoError(t, err)
		require.Equal(t, sdk.NewDecCoins(decCoin(t, "stake", "0.500000000000000002")), gotC.Fees)

		total, err := totalAllocatedFees.Get(ctx)
		require.NoError(t, err)
		require.Equal(t, balance, total.Fees)
	})

	t.Run("trims each denom independently, including denoms absent from the balance", func(t *testing.T) {
		ctx, validatorAllocatedFees, totalAllocatedFees := setupStores(t)
		seed(t, ctx, validatorAllocatedFees, totalAllocatedFees, map[string]sdk.DecCoins{
			"cons-a": sdk.NewDecCoins(decCoin(t, "atom", "0.000000000000000002"), decCoin(t, "stake", "0.5")),
			"cons-b": sdk.NewDecCoins(decCoin(t, "stake", "0.500000000000000001")),
		})

		balance := sdk.NewDecCoins(decCoin(t, "stake", "1"))
		require.NoError(t, v3.Migrate(ctx, balance, validatorAllocatedFees, totalAllocatedFees))

		gotA, err := validatorAllocatedFees.Get(ctx, "cons-a")
		require.NoError(t, err)
		require.Equal(t, sdk.NewDecCoins(decCoin(t, "stake", "0.499999999999999999")), gotA.Fees)

		gotB, err := validatorAllocatedFees.Get(ctx, "cons-b")
		require.NoError(t, err)
		require.Equal(t, sdk.NewDecCoins(decCoin(t, "stake", "0.500000000000000001")), gotB.Fees)

		total, err := totalAllocatedFees.Get(ctx)
		require.NoError(t, err)
		require.Equal(t, balance, total.Fees)
	})
}
