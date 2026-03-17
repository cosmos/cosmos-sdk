package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// statementRemovals removes entire statements (assignments and expression statements)
// that reference deleted modules or deprecated patterns.
var statementRemovals = []migration.StatementRemoval{}

// mapEntryRemovals removes entries from map literals.
var mapEntryRemovals = []migration.MapEntryRemoval{}

// textReplacements defines text-level find-and-replace operations.
// These run AFTER AST transformations and operate on the file as plain text.
// Used for patterns that are multi-line or too deeply nested for AST manipulation
// but have reliable textual patterns.
var textReplacements = []migration.TextReplacement{
	// --- BaseApp method simplification ---
	{Old: "app.BaseApp.GRPCQueryRouter()", New: "app.GRPCQueryRouter()"},
	{Old: "app.BaseApp.Simulate", New: "app.Simulate"},

	// --- EpochsKeeper value -> pointer init pattern ---
	// Convert field assignment to local variable declaration
	{Old: "app.EpochsKeeper = epochskeeper.NewKeeper(", New: "epochsKeeper := epochskeeper.NewKeeper("},
	// Insert pointer assignment before SetHooks call
	{Old: "\tapp.EpochsKeeper.SetHooks(", New: "\tapp.EpochsKeeper = &epochsKeeper\n\n\tapp.EpochsKeeper.SetHooks("},

	// --- nodeservice.RegisterNodeService: add 4th argument ---
	// v54 adds a func() int64 parameter for earliest version.
	// v53 already uses app.GRPCQueryRouter() (not app.BaseApp) as the 2nd arg.
	{
		Old: "nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)",
		New: "nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg, func() int64 { return app.CommitMultiStore().EarliestVersion() })",
	},

	// --- SetModuleVersionMap: add error handling ---
	// v53 discarded the error; v54 requires it
	{
		Old: "\tapp.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())\n\treturn app.ModuleManager.InitGenesis(",
		New: "\terr := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn app.ModuleManager.InitGenesis(",
	},

	// --- auth/tx import dedup: remove unaliased import, rename tx. -> authtx. ---
	// v53 app.go has BOTH unaliased and aliased imports for the same package:
	//   "github.com/cosmos/cosmos-sdk/x/auth/tx"        (unaliased)
	//   authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"  (aliased)
	// We remove the unaliased one and rewrite tx.X -> authtx.X, but ONLY in app.go.
	// Other files (e.g., params/proto.go) only have the unaliased import and use tx.X
	// directly — those should be left untouched.
	{Old: "\t\"github.com/cosmos/cosmos-sdk/x/auth/tx\"\n", New: "", FileMatch: "app.go"},
	// Replace unaliased usages with the alias (only in app.go)
	{Old: "tx.NewTxConfig(", New: "authtx.NewTxConfig(", FileMatch: "app.go"},
	{Old: "tx.DefaultSignModes", New: "authtx.DefaultSignModes", FileMatch: "app.go"},
	{Old: "tx.ConfigOptions", New: "authtx.ConfigOptions", FileMatch: "app.go"},
	{Old: "tx.NewTxConfigWithOptions(", New: "authtx.NewTxConfigWithOptions(", FileMatch: "app.go"},
	// Repair already-aliased `authtx.*` usages after the naive `tx.` rewrite above.
	{Old: "authauthtx.", New: "authtx.", FileMatch: "app.go"},

	// --- params/proto.go: normalize tx import/use for curated v53 fixture ---
	{
		Old: "import (\n\t\"github.com/cosmos/cosmos-sdk/codec\"\n\t\"github.com/cosmos/cosmos-sdk/codec/types\"\n)\n",
		New: "import (\n\t\"github.com/cosmos/cosmos-sdk/codec\"\n\t\"github.com/cosmos/cosmos-sdk/codec/types\"\n\t\"github.com/cosmos/cosmos-sdk/x/auth/tx\"\n)\n",
		FileMatch: "params/proto.go",
	},
	{Old: "authtx.NewTxConfig(", New: "tx.NewTxConfig(", FileMatch: "params/proto.go"},
	{Old: "authtx.DefaultSignModes", New: "tx.DefaultSignModes", FileMatch: "params/proto.go"},

	// --- simd/cmd/root.go: normalize tx import/use for curated v53 fixture ---
	{
		Old: "\t\"github.com/cosmos/cosmos-sdk/types/tx/signing\"\n\tauthtxconfig \"github.com/cosmos/cosmos-sdk/x/auth/tx/config\"\n",
		New: "\t\"github.com/cosmos/cosmos-sdk/types/tx/signing\"\n\t\"github.com/cosmos/cosmos-sdk/x/auth/tx\"\n\tauthtxconfig \"github.com/cosmos/cosmos-sdk/x/auth/tx/config\"\n",
		FileMatch: "simd/cmd/root.go",
	},
	{Old: "authtx.DefaultSignModes", New: "tx.DefaultSignModes", FileMatch: "simd/cmd/root.go"},
	{Old: "authtx.ConfigOptions", New: "tx.ConfigOptions", FileMatch: "simd/cmd/root.go"},
	{Old: "authtx.NewTxConfigWithOptions(", New: "tx.NewTxConfigWithOptions(", FileMatch: "simd/cmd/root.go"},

	// --- app.go: rewrite custom ante wrapper to direct SDK ante handler ---
	{Old: "anteHandler, err := NewAnteHandler(", New: "anteHandler, err := ante.NewAnteHandler(", FileMatch: "app.go"},
	{
		Old: "HandlerOptions{\n\t\t\tante.HandlerOptions{\n",
		New: "ante.HandlerOptions{\n",
		FileMatch: "app.go",
	},
	{
		Old: "\t\t\t},\n\t\t\t&app.CircuitKeeper,\n\t\t},\n\t)\n",
		New: "\t\t\t},\n\t)\n",
		FileMatch: "app.go",
	},

	// --- app.go: strip leftover contrib module order entries ---
	{Old: "\t\tnft.ModuleName,\n", New: "", FileMatch: "app.go"},
	{Old: "\t\tcircuittypes.ModuleName,\n", New: "", FileMatch: "app.go"},

	// ============================================================
	// app_config.go: depinject module configuration cleanup
	// ============================================================
	// No contrib-specific cleanup is needed. group is handled as a fatal warning
	// before file updates run.
}

// fileRemovals deletes files that are no longer needed in v54.
var fileRemovals = []migration.FileRemoval{
	// DI-variant files (build tag !app_v1) — v54 simapp only supports the
	// non-DI variant. These files import x/group and other removed modules;
	// leaving them would cause goimports to re-add stripped imports.
	{
		FileName:          "app_di.go",
		ContainsMustMatch: "!app_v1",
	},
	{
		FileName:          "root_di.go",
		ContainsMustMatch: "!app_v1",
	},
}
