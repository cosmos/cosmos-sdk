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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/group/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/enterprise/group/x/group/errors"
)

func TestSequenceUniqueConstraint(t *testing.T) {
	ctx := NewMockContext()
	store := ctx.KVStore(storetypes.NewKVStoreKey("test"))

	seq := NewSequence(0x1)
	err := seq.InitVal(store, 2)
	require.NoError(t, err)
	err = seq.InitVal(store, 3)
	require.True(t, errors.ErrORMUniqueConstraint.Is(err))
}

func TestSequenceIncrements(t *testing.T) {
	ctx := NewMockContext()
	store := ctx.KVStore(storetypes.NewKVStoreKey("test"))

	seq := NewSequence(0x1)
	var i uint64
	for i = 1; i < 10; i++ {
		autoID := seq.NextVal(store)
		assert.Equal(t, i, autoID)
		assert.Equal(t, i, seq.CurVal(store))
	}

	seq = NewSequence(0x1)
	assert.Equal(t, uint64(10), seq.PeekNextVal(store))
	assert.Equal(t, uint64(9), seq.CurVal(store))
}
