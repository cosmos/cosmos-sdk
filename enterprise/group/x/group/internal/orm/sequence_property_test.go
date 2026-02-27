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

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	storetypes "cosmossdk.io/store/types"
)

func TestSequence(t *testing.T) {
	rapid.Check(t, testSequenceMachine)
}

func testSequenceMachine(t *rapid.T) {
	// Init sets up the real Sequence, including choosing a random initial value,
	// and initializes the model state
	ctx := NewMockContext()
	store := ctx.KVStore(storetypes.NewKVStoreKey("test"))

	// Create primary key table
	seq := NewSequence(0x1)

	// Choose initial sequence value
	initSeqVal := rapid.Uint64().Draw(t, "initSeqVal")
	err := seq.InitVal(store, initSeqVal)
	require.NoError(t, err)

	// Create model state
	state := initSeqVal

	t.Repeat(map[string]func(*rapid.T){
		// NextVal is one of the model commands. It checks that the next value of the
		// sequence matches the model and increments the model state.
		"NextVal": func(t *rapid.T) {
			// Check that the next value in the sequence matches the model
			require.Equal(t, state+1, seq.NextVal(store))
			// Increment the model state
			state++
		},
		// CurVal is one of the model commands. It checks that the current value of the
		// sequence matches the model.
		"CurVal": func(t *rapid.T) {
			// Check the current value matches the model
			require.Equal(t, state, seq.CurVal(store))
		},
		// PeekNextVal is one of the model commands. It checks that the next value of
		// the sequence matches the model without modifying the state.
		"PeekNextVal": func(t *rapid.T) {
			// Check that the next value in the sequence matches the model
			require.Equal(t, state+1, seq.PeekNextVal(store))
		},
	})
}
