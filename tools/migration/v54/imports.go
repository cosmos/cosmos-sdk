package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// importReplacements defines import path rewrites for v53 -> v54.
//
// The primary changes are:
// 1. cosmossdk.io/x/* vanity URLs were removed — modules now live under github.com/cosmos/cosmos-sdk/x/*
// 2. cosmossdk.io/log was upgraded to cosmossdk.io/log/v2
// 3. circuit and nft moved to contrib/ (deprecated, not actively maintained)
// 4. crisis should be removed rather than migrated forward
var importReplacements = []migration.ImportReplacement{
	// --- log/v2 migration ---
	// cosmossdk.io/log -> cosmossdk.io/log/v2
	{Old: "cosmossdk.io/log", New: "cosmossdk.io/log/v2", AllPackages: false},

	// --- mock package migration ---
	// SDK v54-generated mocks use go.uber.org/mock/gomock.
	{Old: "github.com/golang/mock/gomock", New: "go.uber.org/mock/gomock", AllPackages: true},

	// --- Vanity URL migrations (cosmossdk.io/x/* -> github.com/cosmos/cosmos-sdk/x/*) ---
	{Old: "cosmossdk.io/x/feegrant", New: "github.com/cosmos/cosmos-sdk/x/feegrant", AllPackages: true},
	{Old: "cosmossdk.io/x/evidence", New: "github.com/cosmos/cosmos-sdk/x/evidence", AllPackages: true},
	{Old: "cosmossdk.io/x/upgrade", New: "github.com/cosmos/cosmos-sdk/x/upgrade", AllPackages: true},
	{Old: "cosmossdk.io/systemtests", New: "github.com/cosmos/cosmos-sdk/testutil/systemtests", AllPackages: true},

	// --- Modules moved to contrib/ (deprecated) ---
	// Circuit and NFT now live at github.com/cosmos/cosmos-sdk/contrib/x/{circuit,nft}.
	// These are deprecated and not actively maintained.
	{Old: "cosmossdk.io/x/circuit", New: "github.com/cosmos/cosmos-sdk/contrib/x/circuit", AllPackages: true},
	{Old: "cosmossdk.io/x/nft", New: "github.com/cosmos/cosmos-sdk/contrib/x/nft", AllPackages: true},
}

// importWarnings defines import paths that should trigger warnings.
var importWarnings = []migration.ImportWarning{
	{
		ImportPrefix: "cosmossdk.io/x/circuit",
		Message:      "this module has been moved to contrib and will not be maintained by the Cosmos SDK team.",
	},
	{
		ImportPrefix: "github.com/cosmos/cosmos-sdk/contrib/x/circuit",
		Message:      "this module has been moved to contrib and will not be maintained by the Cosmos SDK team.",
	},
	{
		ImportPrefix: "cosmossdk.io/x/crisis",
		Message:      "this module is deprecated and should be removed during migration instead of being carried forward.",
	},
	{
		ImportPrefix: "github.com/cosmos/cosmos-sdk/x/crisis",
		Message:      "this module is deprecated and should be removed during migration instead of being carried forward.",
	},
	{
		ImportPrefix: "cosmossdk.io/x/nft",
		Message:      "this module has been moved to contrib and will not be maintained by the Cosmos SDK team.",
	},
	{
		ImportPrefix: "github.com/cosmos/cosmos-sdk/contrib/x/nft",
		Message:      "this module has been moved to contrib and will not be maintained by the Cosmos SDK team.",
	},
	{
		ImportPrefix: "github.com/cosmos/cosmos-sdk/x/group",
		Message: "the group module is not supported by the v54 migration tool. " +
			"It requires a manual move to enterprise/group.",
		Fatal: true,
	},
	{
		ImportPrefix: "cosmossdk.io/x/group",
		Message: "the group module is not supported by the v54 migration tool. " +
			"It requires a manual move to enterprise/group.",
		Fatal: true,
	},
	{
		ImportPrefix: "cosmossdk.io/api/cosmos/group",
		Message:      "the group module is not supported by the v54 migration tool. It requires a manual move to enterprise/group.",
		Fatal:        true,
	},
}
