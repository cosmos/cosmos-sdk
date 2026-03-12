package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// moduleUpdates defines go.mod dependency version bumps for v53 -> v54.
// These versions should be pinned to stable release tags once available.
//
// TODO: pin to stable v0.54.0 release tags (currently using pre-release commits).
var moduleUpdates = migration.GoModUpdate{
	"github.com/cosmos/cosmos-sdk": "v0.54.0",
	// TODO: add updated versions for the following once v54 release is tagged:
	// "cosmossdk.io/store":               "v1.10.0",
	// "cosmossdk.io/x/tx":                "<version>",
	// "cosmossdk.io/client/v2":           "<version>",
	// "cosmossdk.io/core":                "<version>",
	// "cosmossdk.io/api":                 "<version>",
	// "cosmossdk.io/tools/confix":        "<version>",
	// "github.com/cosmos/cosmos-db":      "v1.1.3",
	// "google.golang.org/grpc":           "v1.73.0",
}

// replacements defines go.mod replace directives needed during migration.
//
// TODO: add any necessary replace directives for pre-release or forked dependencies.
var replacements = []migration.GoModReplacement{}

// additions defines new go.mod dependencies that must be added.
//
// TODO: add any net-new dependencies required by v54 that wouldn't exist in a v53 go.mod.
var additions = migration.GoModAddition{}

// removals defines go.mod dependencies that should be dropped.
// In v54, several cosmossdk.io/x/* vanity URL modules were folded back under the SDK monorepo.
var removals = migration.GoModRemoval{
	"cosmossdk.io/x/circuit",
	"cosmossdk.io/x/evidence",
	"cosmossdk.io/x/upgrade",
	"cosmossdk.io/x/nft",
	"cosmossdk.io/x/feegrant",
	// TODO: verify complete list of removed vanity modules.
	// x/group was moved to enterprise — chains using it will need separate handling.
}
