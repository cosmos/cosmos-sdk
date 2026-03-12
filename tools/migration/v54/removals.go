package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// statementRemovals removes entire statements (assignments and expression statements)
// that reference deleted modules or deprecated patterns.
var statementRemovals = []migration.StatementRemoval{
	// --- Circuit module removal ---
	// Remove: app.CircuitKeeper = circuitkeeper.NewKeeper(...)
	// Also removes the following: app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)
	{
		AssignTarget:     "app.CircuitKeeper",
		IncludeFollowing: 1, // app.BaseApp.SetCircuitBreaker(...)
	},
	// Safety net: also match SetCircuitBreaker independently in case ordering varies
	{
		CallPattern: "app.BaseApp.SetCircuitBreaker",
	},

	// --- NFT module removal ---
	// Remove: app.NFTKeeper = nftkeeper.NewKeeper(...)
	{
		AssignTarget: "app.NFTKeeper",
	},

	// --- Group module removal ---
	// Remove: groupConfig := group.DefaultConfig()  (preceding)
	//         app.GroupKeeper = groupkeeper.NewKeeper(...)
	{
		AssignTarget:           "app.GroupKeeper",
		IncludePrecedingAssign: "groupConfig",
	},
}

// mapEntryRemovals removes entries from map literals.
var mapEntryRemovals = []migration.MapEntryRemoval{
	// Remove nft.ModuleName from maccPerms map
	{
		MapVarName:   "maccPerms",
		KeysToRemove: []string{"nft.ModuleName"},
	},
}

// textReplacements defines text-level find-and-replace operations.
// These run AFTER AST transformations and operate on the file as plain text.
// Used for patterns that are multi-line or too deeply nested for AST manipulation
// but have reliable textual patterns.
var textReplacements = []migration.TextReplacement{
	// --- BaseApp method simplification ---
	{Old: "app.BaseApp.GRPCQueryRouter()", New: "app.GRPCQueryRouter()"},
	{Old: "app.BaseApp.Simulate", New: "app.Simulate"},

	// --- Ante handler: after ante.go is deleted, fix references in app.go ---
	// The custom wrapper used local HandlerOptions type — replace with ante.HandlerOptions
	{Old: "NewAnteHandler(HandlerOptions{", New: "ante.NewAnteHandler(ante.HandlerOptions{"},
	// Remove the CircuitKeeper field from HandlerOptions struct literal
	{Old: "\t\tCircuitKeeper: &app.CircuitKeeper,\n", New: ""},

	// --- EpochsKeeper value -> pointer init pattern ---
	// Convert field assignment to local variable declaration
	{Old: "app.EpochsKeeper = epochskeeper.NewKeeper(", New: "epochsKeeper := epochskeeper.NewKeeper("},
	// Insert pointer assignment before SetHooks call
	{Old: "\tapp.EpochsKeeper.SetHooks(", New: "\tapp.EpochsKeeper = &epochsKeeper\n\n\tapp.EpochsKeeper.SetHooks("},

	// --- nodeservice.RegisterNodeService: add 4th argument ---
	{
		Old: "nodeservice.RegisterNodeService(clientCtx, app.BaseApp, cfg)",
		New: "nodeservice.RegisterNodeService(clientCtx, app.BaseApp, cfg, func() int64 { return app.CommitMultiStore().EarliestVersion() })",
	},

	// --- SetModuleVersionMap: add error handling ---
	// v53 discarded the error; v54 requires it
	{
		Old: "\tapp.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())\n\treturn app.ModuleManager.InitGenesis(",
		New: "\terr := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn app.ModuleManager.InitGenesis(",
	},

	// --- auth/tx import dedup: remove unaliased import, rename tx. -> authtx. ---
	// v53 has both unaliased and aliased imports for the same package
	{Old: "\t\"github.com/cosmos/cosmos-sdk/x/auth/tx\"\n", New: ""},
	// Replace unaliased usages with the alias
	{Old: "tx.NewTxConfig(", New: "authtx.NewTxConfig("},
	{Old: "tx.DefaultSignModes", New: "authtx.DefaultSignModes"},
	{Old: "tx.ConfigOptions", New: "authtx.ConfigOptions"},
	{Old: "tx.NewTxConfigWithOptions(", New: "authtx.NewTxConfigWithOptions("},

	// --- Import cleanup: remove circuitante import remnants ---
	{Old: "\tcircuitante \"cosmossdk.io/x/circuit/ante\"\n", New: ""},
	{Old: "\tcircuitante \"github.com/cosmos/cosmos-sdk/contrib/x/circuit/ante\"\n", New: ""},
}

// fileRemovals deletes files that are no longer needed in v54.
var fileRemovals = []migration.FileRemoval{
	// The custom ante.go wrapper was only needed for the circuit ante decorator.
	// v54 uses ante.NewAnteHandler directly without the circuit wrapper.
	{
		FileName:          "ante.go",
		ContainsMustMatch: "circuitante",
	},
}
