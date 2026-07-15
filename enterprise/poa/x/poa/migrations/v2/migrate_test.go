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

package v2_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/collections/indexes"

	v2 "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/migrations/v2"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValIndexes mirrors the keeper's validator power index, the only index the
// migration reads.
type ValIndexes struct {
	Power *indexes.Multi[int64, sdk.ConsAddress, poatypes.Validator]
}

func (v ValIndexes) IndexesList() []collections.Index[sdk.ConsAddress, poatypes.Validator] {
	return []collections.Index[sdk.ConsAddress, poatypes.Validator]{v.Power}
}

func TestMigrate(t *testing.T) {
	t.Run("records active validators", func(t *testing.T) {
		kv, ctx := colltest.MockStore()
		sb := collections.NewSchemaBuilder(kv)

		validators := collections.NewIndexedMap(
			sb,
			poatypes.ValidatorsKey,
			"validators",
			sdk.ConsAddressKey,
			colltest.MockValueCodec[poatypes.Validator](),
			ValIndexes{
				Power: indexes.NewMulti(
					sb,
					poatypes.ValidatorPowerIndex,
					"validator_by_power",
					collections.Int64Key,
					sdk.ConsAddressKey,
					func(_ sdk.ConsAddress, v poatypes.Validator) (int64, error) { return v.Power, nil },
				),
			},
		)
		lastCommittedPower := collections.NewMap(sb, poatypes.LastCommittedPowerKey, "last_committed_power", sdk.ConsAddressKey, collections.Int64Value)

		consA, consB, consZero := sdk.ConsAddress("cons-a"), sdk.ConsAddress("cons-b"), sdk.ConsAddress("cons-zero")
		require.NoError(t, validators.Set(ctx, consA, poatypes.Validator{Power: 100}))
		require.NoError(t, validators.Set(ctx, consB, poatypes.Validator{Power: 50}))
		require.NoError(t, validators.Set(ctx, consZero, poatypes.Validator{Power: 0}))

		require.NoError(t, v2.Migrate(ctx, validators.Indexes.Power, lastCommittedPower))

		gotA, err := lastCommittedPower.Get(ctx, consA)
		require.NoError(t, err)
		require.Equal(t, int64(100), gotA)

		gotB, err := lastCommittedPower.Get(ctx, consB)
		require.NoError(t, err)
		require.Equal(t, int64(50), gotB)

		// zero power validators are not in CometBFT's set, so they must not be seeded.
		hasZero, err := lastCommittedPower.Has(ctx, consZero)
		require.NoError(t, err)
		require.False(t, hasZero)
	})

	t.Run("no validators", func(t *testing.T) {
		kv, ctx := colltest.MockStore()
		sb := collections.NewSchemaBuilder(kv)

		validators := collections.NewIndexedMap(
			sb,
			poatypes.ValidatorsKey,
			"validators",
			sdk.ConsAddressKey,
			colltest.MockValueCodec[poatypes.Validator](),
			ValIndexes{
				Power: indexes.NewMulti(
					sb,
					poatypes.ValidatorPowerIndex,
					"validator_by_power",
					collections.Int64Key,
					sdk.ConsAddressKey,
					func(_ sdk.ConsAddress, v poatypes.Validator) (int64, error) { return v.Power, nil },
				),
			},
		)
		lastCommittedPower := collections.NewMap(sb, poatypes.LastCommittedPowerKey, "last_committed_power", sdk.ConsAddressKey, collections.Int64Value)

		require.NoError(t, v2.Migrate(ctx, validators.Indexes.Power, lastCommittedPower))

		// ensure that we walk over nothing in the last committed power list
		// since there are no active validators
		ranger := new(collections.Range[sdk.ConsAddress]).Descending()
		lastCommittedPower.Walk(ctx, ranger, func(key sdk.ConsAddress, value int64) (stop bool, err error) {
			assert.FailNow(t, "callback should not have been called")
			return false, nil
		})
	})
}
