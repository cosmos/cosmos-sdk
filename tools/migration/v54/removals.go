package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// statementRemovals removes entire statements (assignments and expression statements)
// that reference deleted modules or deprecated patterns.
var statementRemovals = []migration.StatementRemoval{
	{AssignTarget: "app.CrisisKeeper"},
	{CallPattern: "crisis.AddModuleInitFlags"},
	{CallPattern: "app.ModuleManager.RegisterInvariants"},
}

// mapEntryRemovals removes entries from map literals.
var mapEntryRemovals = []migration.MapEntryRemoval{
	{
		MapVarName: "InternalMsgSamplesDefault",
		KeysToRemove: []string{
			`"/cosmos.crisis.v1beta1.MsgUpdateParams"`,
			`"/cosmos.crisis.v1beta1.MsgUpdateParamsResponse"`,
			`"/cosmos.staking.v1beta1.MsgSetProposers"`,
			`"/cosmos.staking.v1beta1.MsgSetProposersResponse"`,
		},
	},
	{
		MapVarName: "UnsupportedMsgSamples",
		KeysToRemove: []string{
			`"/cosmos.crisis.v1beta1.MsgVerifyInvariant"`,
			`"/cosmos.crisis.v1beta1.MsgVerifyInvariantResponse"`,
		},
	},
	{
		MapVarName: "AllTypeMessages",
		KeysToRemove: []string{
			`"/cosmos.crisis.v1beta1.MsgUpdateParams"`,
			`"/cosmos.crisis.v1beta1.MsgUpdateParamsResponse"`,
			`"/cosmos.crisis.v1beta1.MsgVerifyInvariant"`,
			`"/cosmos.crisis.v1beta1.MsgVerifyInvariantResponse"`,
			`"/cosmos.staking.v1beta1.MsgSetProposers"`,
			`"/cosmos.staking.v1beta1.MsgSetProposersResponse"`,
		},
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

	// --- tx.Factory API cleanup ---
	// The non-critical variant was folded into WithExtensionOptions.
	{Old: ".WithNonCriticalExtensionOptions(", New: ".WithExtensionOptions("},
	{Old: "telemetry.MetricKeyPrecommiter", New: `"precommitter"`},
	{Old: "telemetry.MetricKeyPrepareCheckStater", New: `"prepare_check_stater"`},

	// --- EpochsKeeper value -> pointer init pattern ---
	// Convert field assignment to local variable declaration
	{Old: "app.EpochsKeeper = epochskeeper.NewKeeper(", New: "epochsKeeper := epochskeeper.NewKeeper("},
	// Insert pointer assignment before SetHooks call
	{Old: "\tapp.EpochsKeeper.SetHooks(", New: "\tapp.EpochsKeeper = &epochsKeeper\n\n\tapp.EpochsKeeper.SetHooks("},
	// Some chains use a chain-local epochs keeper constructor that already returns *Keeper.
	// Once the field type is migrated to a pointer, only the explicit dereference must be removed.
	{Old: "app.EpochsKeeper = *epochskeeper.NewKeeper(", New: "app.EpochsKeeper = epochskeeper.NewKeeper(", FileMatch: "app.go"},
	{Old: "app.EpochsKeeper = *epochsmodulekeeper.NewKeeper(", New: "app.EpochsKeeper = epochsmodulekeeper.NewKeeper(", FileMatch: "app.go"},

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
		Old:       "import (\n\t\"github.com/cosmos/cosmos-sdk/codec\"\n\t\"github.com/cosmos/cosmos-sdk/codec/types\"\n)\n",
		New:       "import (\n\t\"github.com/cosmos/cosmos-sdk/codec\"\n\t\"github.com/cosmos/cosmos-sdk/codec/types\"\n\t\"github.com/cosmos/cosmos-sdk/x/auth/tx\"\n)\n",
		FileMatch: "params/proto.go",
	},
	{Old: "authtx.NewTxConfig(", New: "tx.NewTxConfig(", FileMatch: "params/proto.go"},
	{Old: "authtx.DefaultSignModes", New: "tx.DefaultSignModes", FileMatch: "params/proto.go"},

	// --- simd/cmd/root.go: normalize tx import/use for curated v53 fixture ---
	{
		Old:       "\t\"github.com/cosmos/cosmos-sdk/types/tx/signing\"\n\tauthtxconfig \"github.com/cosmos/cosmos-sdk/x/auth/tx/config\"\n",
		New:       "\t\"github.com/cosmos/cosmos-sdk/types/tx/signing\"\n\t\"github.com/cosmos/cosmos-sdk/x/auth/tx\"\n\tauthtxconfig \"github.com/cosmos/cosmos-sdk/x/auth/tx/config\"\n",
		FileMatch: "simd/cmd/root.go",
	},
	{Old: "authtx.DefaultSignModes", New: "tx.DefaultSignModes", FileMatch: "simd/cmd/root.go"},
	{Old: "authtx.ConfigOptions", New: "tx.ConfigOptions", FileMatch: "simd/cmd/root.go"},
	{Old: "authtx.NewTxConfigWithOptions(", New: "tx.NewTxConfigWithOptions(", FileMatch: "simd/cmd/root.go"},

	// --- app.go: rewrite custom ante wrapper to direct SDK ante handler ---
	{
		Old:              "anteHandler, err := NewAnteHandler(",
		New:              "anteHandler, err := ante.NewAnteHandler(",
		FileMatch:        "app.go",
		RequiresContains: []string{"&app.CircuitKeeper"},
	},
	{
		Old:              "HandlerOptions{\n\t\t\tante.HandlerOptions{\n",
		New:              "ante.HandlerOptions{\n",
		FileMatch:        "app.go",
		RequiresContains: []string{"&app.CircuitKeeper"},
	},
	{
		Old:              "\t\t\t},\n\t\t\t&app.CircuitKeeper,\n\t\t},\n\t)\n",
		New:              "\t\t\t},\n\t)\n",
		FileMatch:        "app.go",
		RequiresContains: []string{"&app.CircuitKeeper"},
	},

	// --- app.go: strip leftover contrib module order entries ---
	{Old: "\t\tnft.ModuleName,\n", New: "", FileMatch: "app.go"},
	{Old: "\t\tcircuittypes.ModuleName,\n", New: "", FileMatch: "app.go"},
	// --- app.go: remove crisis-specific module wiring ---
	{Old: "\tvar skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))\n", New: "", FileMatch: "app.go"},
	{Old: "\tparamsKeeper.Subspace(crisistypes.ModuleName)\n", New: "", FileMatch: "app.go"},

	// --- Switch/map cleanup after crisis removals ---
	{
		Old:       "\n\t\t// crisis\n\t\t*crisis.MsgUpdateParams,\n",
		New:       "\n",
		FileMatch: "lib/ante/internal_msg.go",
	},
	{
		Old:       "\n\t\t*crisis.MsgUpdateParams,\n",
		New:       "\n",
		FileMatch: "lib/ante/internal_msg.go",
	},
	{
		Old:       "\n\t\t*crisis.MsgUpdateParams:\n",
		New:       "\n",
		FileMatch: "lib/ante/internal_msg.go",
	},
	{
		Old:       "\t\t*crisis.MsgVerifyInvariant:\n",
		New:       "",
		FileMatch: "lib/ante/unsupported_msgs.go",
	},
	{
		Old:       "\t\t*vaulttypes.MsgUpdateParams,\n\t\treturn true\n",
		New:       "\t\t*vaulttypes.MsgUpdateParams:\n\t\treturn true\n",
		FileMatch: "lib/ante/unsupported_msgs.go",
	},
	{
		Old:       "\t\t// crisis\n",
		New:       "",
		FileMatch: "app/msgs/internal_msgs.go",
	},
	{
		Old:       "\n\t\t// Disable MsgVerifyInvariant in the crisis module, since:\n\t\t// 1. We currently do not rely on crisis module for any invariant assertion.\n\t\t// 2. MsgVerifyInvariant can potentially be abused to consume massive compute.\n",
		New:       "\n",
		FileMatch: "app/msgs/unsupported_msgs.go",
	},

	// --- Remove stale staking MsgSetProposers references ---
	{
		Old:       "\t\t\"/cosmos.staking.v1beta1.MsgSetProposers\",\n",
		New:       "",
		FileMatch: "app/msgs/internal_msgs_test.go",
	},
	{
		Old:       "\t\t\"/cosmos.staking.v1beta1.MsgSetProposersResponse\",\n",
		New:       "",
		FileMatch: "app/msgs/internal_msgs_test.go",
	},
	{
		Old:       "\t\t*staking.MsgSetProposers,\n",
		New:       "",
		FileMatch: "lib/ante/internal_msg.go",
	},
	{
		Old:       "\t\t\"/cosmos.crisis.v1beta1.MsgUpdateParams\",\n",
		New:       "",
		FileMatch: "app/msgs/internal_msgs_test.go",
	},
	{
		Old:       "\t\t\"/cosmos.crisis.v1beta1.MsgUpdateParamsResponse\",\n",
		New:       "",
		FileMatch: "app/msgs/internal_msgs_test.go",
	},
	{
		Old:       "\t\t\"/cosmos.crisis.v1beta1.MsgVerifyInvariant\",\n",
		New:       "",
		FileMatch: "app/msgs/unsupported_msgs_test.go",
	},
	{
		Old:       "\t\t\"/cosmos.crisis.v1beta1.MsgVerifyInvariantResponse\",\n",
		New:       "",
		FileMatch: "app/msgs/unsupported_msgs_test.go",
	},

	// --- curated simapp fixture: remove traceStore plumbing dropped in v54 ---
	{
		Old:       "\tdb dbm.DB,\n\ttraceStore io.Writer,\n\tloadLatest bool,\n",
		New:       "\tdb dbm.DB,\n\tloadLatest bool,\n",
		FileMatch: "app.go",
	},
	{Old: "\tbApp.SetCommitMultiStoreTracer(traceStore)\n", New: "", FileMatch: "app.go"},
	{Old: "\tbApp.SetCommitMultiStoreTracer(nil)\n", New: "", FileMatch: "app_test.go"},
	{
		Old:       "\tlogger log.Logger,\n\tdb dbm.DB,\n\ttraceStore io.Writer,\n\tappOpts servertypes.AppOptions,\n",
		New:       "\tlogger log.Logger,\n\tdb dbm.DB,\n\tappOpts servertypes.AppOptions,\n",
		FileMatch: "simd/cmd/commands.go",
	},
	{
		Old:       "\t\tlogger, db, traceStore, true,\n",
		New:       "\t\tlogger, db, true,\n",
		FileMatch: "simd/cmd/commands.go",
	},
	{
		Old:       "\tlogger log.Logger,\n\tdb dbm.DB,\n\ttraceStore io.Writer,\n\theight int64,\n",
		New:       "\tlogger log.Logger,\n\tdb dbm.DB,\n\theight int64,\n",
		FileMatch: "simd/cmd/commands.go",
	},
	{
		Old:       "\t\tsimApp = simapp.NewSimApp(logger, db, traceStore, false, appOpts)\n",
		New:       "\t\tsimApp = simapp.NewSimApp(logger, db, false, appOpts)\n",
		FileMatch: "simd/cmd/commands.go",
	},
	{
		Old:       "\t\tsimApp = simapp.NewSimApp(logger, db, traceStore, true, appOpts)\n",
		New:       "\t\tsimApp = simapp.NewSimApp(logger, db, true, appOpts)\n",
		FileMatch: "simd/cmd/commands.go",
	},
	{Old: "NewSimApp(log.NewNopLogger(), db, nil, true, appOptions)", New: "NewSimApp(log.NewNopLogger(), db, true, appOptions)", FileMatch: "test_helpers.go"},
	{Old: "NewSimApp(options.Logger, options.DB, nil, true, options.AppOpts)", New: "NewSimApp(options.Logger, options.DB, true, options.AppOpts)", FileMatch: "test_helpers.go"},
	{Old: "NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(dir))", New: "NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(dir))", FileMatch: "test_helpers.go"},
	{
		Old:       "NewSimApp(\n\t\t\tlog.NewNopLogger(), dbm.NewMemDB(), nil, true,",
		New:       "NewSimApp(\n\t\t\tlog.NewNopLogger(), dbm.NewMemDB(), true,",
		FileMatch: "test_helpers.go",
	},
	{
		Old:       "return NewSimApp(\n\t\t\tval.GetCtx().Logger, dbm.NewMemDB(), nil, true,\n",
		New:       "return NewSimApp(\n\t\t\tval.GetCtx().Logger, dbm.NewMemDB(), true,\n",
		FileMatch: "test_helpers.go",
	},
	{Old: "NewSimApp(logger, db, nil, true, appOpts, append(baseAppOptions, interBlockCacheOpt())...)", New: "NewSimApp(logger, db, true, appOpts, append(baseAppOptions, interBlockCacheOpt())...)", FileMatch: "sim_test.go"},
	{Old: "NewSimApp(logger.With(\"instance\", \"second\"), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))", New: "NewSimApp(logger.With(\"instance\", \"second\"), db, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))", FileMatch: "app_test.go"},
	{Old: "NewSimApp(logger.With(\"instance\", \"simapp\"), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))", New: "NewSimApp(logger.With(\"instance\", \"simapp\"), db, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))", FileMatch: "app_test.go"},
	{Old: "NewSimApp(log.NewTestLogger(t), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))", New: "NewSimApp(log.NewTestLogger(t), db, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))", FileMatch: "app_test.go"},
	{Old: "NewSimApp(logger, db, nil, true, appOptions, interBlockCacheOpt(), baseapp.SetChainID(simsx.SimAppChainID))", New: "NewSimApp(logger, db, true, appOptions, interBlockCacheOpt(), baseapp.SetChainID(simsx.SimAppChainID))", FileMatch: "sim_bench_test.go"},
	{Old: "simapp.NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(simapp.DefaultNodeHome))", New: "simapp.NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(simapp.DefaultNodeHome))", FileMatch: "simd/cmd/root.go"},

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
