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

import storetypes "cosmossdk.io/store/types"

// TableExportable defines the methods to import and export a table.
type TableExportable interface {
	// Export stores all the values in the table in the passed
	// ModelSlicePtr. If the table has an associated sequence, then its
	// current value is returned, otherwise 0 is returned by default.
	Export(store storetypes.KVStore, dest ModelSlicePtr) (uint64, error)

	// Import clears the table and initializes it from the given data
	// interface{}. data should be a slice of structs that implement
	// PrimaryKeyed. The seqValue is optional and only
	// used with tables that have an associated sequence.
	Import(store storetypes.KVStore, data any, seqValue uint64) error
}
