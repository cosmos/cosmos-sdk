package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// moduleUpdates defines go.mod dependency version bumps for v53 -> v54.
// These are pinned to the versions used by simapp on the main branch.
//
// NOTE: github.com/cosmos/cosmos-sdk itself uses a pseudo version derived from main.
// Once v0.54.0 is tagged, this should be updated to the stable release tag.
var moduleUpdates = migration.GoModUpdate{
	// Core SDK — pseudo version from main (commit 2c527014f3ee, 2026-03-11)
	"github.com/cosmos/cosmos-sdk": "v0.54.0-rc.0.20260311165803-2c527014f3ee",

	// SDK companion modules
	"cosmossdk.io/api":       "v1.0.0",
	"cosmossdk.io/client/v2": "v2.0.0-beta.11",
	"cosmossdk.io/core":      "v1.1.0",
	"cosmossdk.io/depinject": "v1.2.1",
	"cosmossdk.io/store":     "v1.3.0-beta.0",
	"cosmossdk.io/math":      "v1.5.3",

	// CometBFT
	"github.com/cometbft/cometbft": "v0.39.0-beta.3",

	// Other direct dependencies commonly bumped
	"github.com/cosmos/cosmos-db": "v1.1.3",
	"github.com/cosmos/gogoproto": "v1.7.2",
}

// additions defines new go.mod dependencies that must be added.
// In v54, cosmossdk.io/log was replaced by cosmossdk.io/log/v2.
var additions = migration.GoModAddition{
	"cosmossdk.io/log/v2": "v2.0.1",
}

// removals defines go.mod dependencies that should be dropped.
// In v54, several cosmossdk.io/x/* vanity URL modules were folded back under the SDK monorepo,
// so their separate go.mod entries should be removed.
var removals = migration.GoModRemoval{
	// Vanity URL modules folded into SDK monorepo
	"cosmossdk.io/x/circuit",
	"cosmossdk.io/x/evidence",
	"cosmossdk.io/x/upgrade",
	"cosmossdk.io/x/nft",
	"cosmossdk.io/x/feegrant",
	"cosmossdk.io/x/tx",

	// Old log module replaced by log/v2
	"cosmossdk.io/log",
}

// replacements defines go.mod replace directives needed during migration.
var replacements = []migration.GoModReplacement{}
