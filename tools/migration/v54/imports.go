package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// importReplacements defines import path rewrites for v53 -> v54.
//
// The primary changes are:
// 1. cosmossdk.io/x/* vanity URLs were removed — modules now live under github.com/cosmos/cosmos-sdk/x/*
// 2. cosmossdk.io/log was upgraded to cosmossdk.io/log/v2
// 3. circuit and nft moved to contrib/ (deprecated, not actively maintained)
// 4. group moved to enterprise/ (different license — handled via warning, not auto-rewrite)
var importReplacements = []migration.ImportReplacement{
	// --- log/v2 migration ---
	// cosmossdk.io/log -> cosmossdk.io/log/v2
	{Old: "cosmossdk.io/log", New: "cosmossdk.io/log/v2", AllPackages: false},

	// --- Vanity URL migrations (cosmossdk.io/x/* -> github.com/cosmos/cosmos-sdk/x/*) ---
	{Old: "cosmossdk.io/x/feegrant", New: "github.com/cosmos/cosmos-sdk/x/feegrant", AllPackages: true},
	{Old: "cosmossdk.io/x/evidence", New: "github.com/cosmos/cosmos-sdk/x/evidence", AllPackages: true},
	{Old: "cosmossdk.io/x/upgrade", New: "github.com/cosmos/cosmos-sdk/x/upgrade", AllPackages: true},
	{Old: "cosmossdk.io/x/tx", New: "github.com/cosmos/cosmos-sdk/x/tx", AllPackages: true},

	// --- Modules moved to contrib/ (deprecated) ---
	// Circuit and NFT now live at github.com/cosmos/cosmos-sdk/contrib/x/{circuit,nft}.
	// These are deprecated and not actively maintained.
	{Old: "cosmossdk.io/x/circuit", New: "github.com/cosmos/cosmos-sdk/contrib/x/circuit", AllPackages: true},
	{Old: "cosmossdk.io/x/nft", New: "github.com/cosmos/cosmos-sdk/contrib/x/nft", AllPackages: true},
}

// importWarnings defines import paths that should trigger warnings rather than automatic rewrites.
// AlsoRemove is set to true so the imports are stripped from the AST after warning, preventing
// the code from trying to resolve the old module path. The warnings still inform the user that
// x/group requires a commercial license to use in v54.
var importWarnings = []migration.ImportWarning{
	// group is under github.com/cosmos/cosmos-sdk/x/group (not a vanity URL)
	{
		ImportPrefix: "github.com/cosmos/cosmos-sdk/x/group",
		Message: "The x/group module has been moved to enterprise/group with a commercial license. " +
			"You must contact Cosmos Labs to establish a commercial agreement before using this module. " +
			"See enterprise/README.md for details. This import will NOT be automatically rewritten.",
		AlsoRemove: true,
	},
	// Also catch any cosmossdk.io/x/group references (in case any exist)
	{
		ImportPrefix: "cosmossdk.io/x/group",
		Message: "The x/group module has been moved to enterprise/group with a commercial license. " +
			"You must contact Cosmos Labs to establish a commercial agreement before using this module. " +
			"See enterprise/README.md for details. This import will NOT be automatically rewritten.",
		AlsoRemove: true,
	},
	// Also catch cosmossdk.io/api/cosmos/group (proto API imports used in app_config.go)
	{
		ImportPrefix: "cosmossdk.io/api/cosmos/group",
		Message:      "The x/group module API has been moved to enterprise/group with a commercial license.",
		AlsoRemove:   true,
	},
}
