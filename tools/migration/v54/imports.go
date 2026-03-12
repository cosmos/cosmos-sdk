package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// importReplacements defines import path rewrites for v53 -> v54.
// The primary change is that cosmossdk.io/x/* vanity URLs were removed in v54.
// Modules that lived under separate go.mods at cosmossdk.io/x/<module> are now
// imported from github.com/cosmos/cosmos-sdk/x/<module>.
var importReplacements = []migration.ImportReplacement{
	// --- Vanity URL migrations (cosmossdk.io/x/* -> github.com/cosmos/cosmos-sdk/x/*) ---
	{Old: "cosmossdk.io/x/feegrant", New: "github.com/cosmos/cosmos-sdk/x/feegrant", AllPackages: true},
	{Old: "cosmossdk.io/x/circuit", New: "github.com/cosmos/cosmos-sdk/x/circuit", AllPackages: true},
	{Old: "cosmossdk.io/x/nft", New: "github.com/cosmos/cosmos-sdk/x/nft", AllPackages: true},
	{Old: "cosmossdk.io/x/evidence", New: "github.com/cosmos/cosmos-sdk/x/evidence", AllPackages: true},
	{Old: "cosmossdk.io/x/upgrade", New: "github.com/cosmos/cosmos-sdk/x/upgrade", AllPackages: true},

	// TODO: verify if any other cosmossdk.io/x/* modules moved.
	// Candidates to check:
	// - cosmossdk.io/x/auth (still separate? or folded in?)
	// - cosmossdk.io/x/bank
	// - cosmossdk.io/x/staking
	// - cosmossdk.io/x/gov
	// - cosmossdk.io/x/distribution
	// - cosmossdk.io/x/slashing
	// - cosmossdk.io/x/mint
	// - cosmossdk.io/x/params
	// - cosmossdk.io/x/group (moved to enterprise, needs special handling)

	// --- log/v2 migration ---
	// TODO: add log -> log/v2 import migration once we confirm the exact path change.
	// {Old: "cosmossdk.io/log", New: "cosmossdk.io/log/v2"},
}
