package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// typeReplacements defines type/struct renames between v53 and v54.
//
// TODO: audit v54 for type renames.
// Known areas to investigate:
// - any ABCI type renames that carried over from CometBFT changes
// - store v2 type changes
// - any renamed config structs
var typeReplacements = []migration.TypeReplacement{
	// Example format:
	// {
	// 	ImportPath: "github.com/cosmos/cosmos-sdk/store/types",
	// 	OldType:    "CommitMultiStore",
	// 	NewType:    "RootStore",
	// },
}
