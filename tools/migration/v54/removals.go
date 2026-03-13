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
	// The v53 simapp has a custom NewAnteHandler wrapper in ante.go that embeds
	// ante.HandlerOptions plus a CircuitKeeper field. After ante.go is deleted,
	// we rewrite the call site to use ante.NewAnteHandler with ante.HandlerOptions
	// directly, removing the outer wrapper struct and the CircuitKeeper argument.
	//
	// v53 pattern (multi-line):
	//   anteHandler, err := NewAnteHandler(
	//       HandlerOptions{
	//           ante.HandlerOptions{
	//               ...
	//           },
	//           &app.CircuitKeeper,
	//       },
	//   )
	//
	// v54 result:
	//   anteHandler, err := ante.NewAnteHandler(
	//       ante.HandlerOptions{
	//           ...
	//       },
	//   )
	{
		Old: "\tanteHandler, err := NewAnteHandler(\n\t\tHandlerOptions{\n\t\t\tante.HandlerOptions{",
		New: "\tanteHandler, err := ante.NewAnteHandler(\n\t\tante.HandlerOptions{",
	},
	// Remove the closing bracket of the inner struct, CircuitKeeper arg, and outer wrapper bracket.
	// Before: \t\t\t},\n\t\t\t&app.CircuitKeeper,\n\t\t},
	// After:  \t\t},
	{
		Old: "\t\t\t},\n\t\t\t&app.CircuitKeeper,\n\t\t},",
		New: "\t\t},",
	},

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

	// --- Import cleanup: remove circuitante import remnants ---
	{Old: "\tcircuitante \"cosmossdk.io/x/circuit/ante\"\n", New: ""},
	{Old: "\tcircuitante \"github.com/cosmos/cosmos-sdk/contrib/x/circuit/ante\"\n", New: ""},

	// --- x/group imports are now removed at the AST level via ImportWarning.AlsoRemove ---
	// (see imports.go). The text replacements below handle non-import type references
	// that would otherwise leave dangling identifiers causing goimports to re-add imports.

	// --- Remove module names from genesisModuleOrder / exportModuleOrder slice literals ---
	// These are local []string variables in app.go whose elements are NOT function call args,
	// so CallArgRemoval cannot handle them. goimports will re-add stripped imports if these
	// dangling identifiers remain.
	{Old: "\t\tnft.ModuleName,\n", New: "", FileMatch: "app.go"},
	{Old: "\t\tgroup.ModuleName,\n", New: "", FileMatch: "app.go"},
	{Old: "\t\tcircuittypes.ModuleName,\n", New: "", FileMatch: "app.go"},

	// --- Remove group type references from test map literals ---
	// app_test.go has a consensus version map with group entries.
	// After imports are stripped, these dangling references would cause
	// goimports to re-add the x/group imports.
	{Old: "\t\t\t\t\tgrouptypes.ModuleName:        group.AppModule{}.ConsensusVersion(),\n", New: ""},

	// ============================================================
	// app_config.go: depinject module configuration cleanup
	// ============================================================
	// These patterns remove circuit, nft, and group references from
	// the depinject config (struct literal arrays, protobuf entries).
	// FileMatch restricts to app_config.go so patterns like "time"
	// import or nft.ModuleName don't accidentally match elsewhere.

	// --- Remove proto API imports (not touched by import rewriter) ---
	// Note: group proto API imports (cosmossdk.io/api/cosmos/group/*) are now removed
	// at the AST level via ImportWarning.AlsoRemove in imports.go.
	{Old: "\tcircuitmodulev1 \"cosmossdk.io/api/cosmos/circuit/module/v1\"\n", New: "", FileMatch: "app_config.go"},
	{Old: "\tnftmodulev1 \"cosmossdk.io/api/cosmos/nft/module/v1\"\n", New: "", FileMatch: "app_config.go"},

	// --- Remove circuit side-effect and named imports (post-rewrite paths) ---
	{Old: "\t_ \"github.com/cosmos/cosmos-sdk/contrib/x/circuit\" // import for side-effects\n", New: "", FileMatch: "app_config.go"},
	{Old: "\tcircuittypes \"github.com/cosmos/cosmos-sdk/contrib/x/circuit/types\"\n", New: "", FileMatch: "app_config.go"},
	// Also match pre-rewrite paths in case import rewriter didn't run
	{Old: "\t_ \"cosmossdk.io/x/circuit\" // import for side-effects\n", New: "", FileMatch: "app_config.go"},
	{Old: "\tcircuittypes \"cosmossdk.io/x/circuit/types\"\n", New: "", FileMatch: "app_config.go"},

	// --- Remove nft imports (post-rewrite paths) ---
	{Old: "\t\"github.com/cosmos/cosmos-sdk/contrib/x/nft\"\n", New: "", FileMatch: "app_config.go"},
	{Old: "\t_ \"github.com/cosmos/cosmos-sdk/contrib/x/nft/module\" // import for side-effects\n", New: "", FileMatch: "app_config.go"},
	// Also match pre-rewrite paths
	{Old: "\t\"cosmossdk.io/x/nft\"\n", New: "", FileMatch: "app_config.go"},
	{Old: "\t_ \"cosmossdk.io/x/nft/module\" // import for side-effects\n", New: "", FileMatch: "app_config.go"},

	// --- Remove group side-effect import: now handled by AST ImportWarning.AlsoRemove ---

	// --- Remove time and durationpb imports (only used by group's MaxExecutionPeriod) ---
	{Old: "\t\"time\"\n", New: "", FileMatch: "app_config.go"},
	{Old: "\t\"google.golang.org/protobuf/types/known/durationpb\"\n", New: "", FileMatch: "app_config.go"},

	// --- Remove nft.ModuleName from moduleAccPerms array ---
	{Old: "\t\t{Account: nft.ModuleName},\n", New: "", FileMatch: "app_config.go"},

	// --- Remove nft.ModuleName from blockAccAddrs array ---
	{Old: "\t\tnft.ModuleName,\n", New: "", FileMatch: "app_config.go"},

	// --- Remove module names from runtime string arrays ---
	// These remove circuit/nft/group from EndBlockers, InitGenesis, ExportGenesis.
	// ReplaceAll handles all occurrences at the 5-tab indent level.
	{Old: "\t\t\t\t\tgroup.ModuleName,\n", New: "", FileMatch: "app_config.go"},
	{Old: "\t\t\t\t\tnft.ModuleName,\n", New: "", FileMatch: "app_config.go"},
	{Old: "\t\t\t\t\tcircuittypes.ModuleName,\n", New: "", FileMatch: "app_config.go"},

	// --- Remove group ModuleConfig entry (multi-line) ---
	{Old: "\t\t{\n\t\t\tName: group.ModuleName,\n\t\t\tConfig: appconfig.WrapAny(&groupmodulev1.Module{\n\t\t\t\tMaxExecutionPeriod: durationpb.New(time.Second * 1209600),\n\t\t\t\tMaxMetadataLen:     255,\n\t\t\t}),\n\t\t},\n", New: "", FileMatch: "app_config.go"},

	// --- Remove nft ModuleConfig entry ---
	{Old: "\t\t{\n\t\t\tName:   nft.ModuleName,\n\t\t\tConfig: appconfig.WrapAny(&nftmodulev1.Module{}),\n\t\t},\n", New: "", FileMatch: "app_config.go"},

	// --- Remove circuit ModuleConfig entry ---
	{Old: "\t\t{\n\t\t\tName:   circuittypes.ModuleName,\n\t\t\tConfig: appconfig.WrapAny(&circuitmodulev1.Module{}),\n\t\t},\n", New: "", FileMatch: "app_config.go"},
}

// fileRemovals deletes files that are no longer needed in v54.
var fileRemovals = []migration.FileRemoval{
	// The custom ante.go wrapper was only needed for the circuit ante decorator.
	// v54 uses ante.NewAnteHandler directly without the circuit wrapper.
	{
		FileName:          "ante.go",
		ContainsMustMatch: "circuitante",
	},
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
