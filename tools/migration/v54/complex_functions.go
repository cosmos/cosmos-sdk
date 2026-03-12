package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// complexUpdates defines function call replacements that require multi-statement rewrites.
var complexUpdates = []migration.ComplexFunctionReplacement{
	// No complex function replacements needed for v53 -> v54.
	// The changes that could qualify are handled by other primitives:
	// - govkeeper.NewKeeper → argSurgeries
	// - Circuit/NFT/Group removal → statementRemovals + fieldRemovals
	// - Ante handler → fileRemovals + textReplacements
	// - EpochsKeeper → fieldModifications + textReplacements
}
