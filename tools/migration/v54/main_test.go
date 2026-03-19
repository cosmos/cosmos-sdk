package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	migration "github.com/cosmos/cosmos-sdk/tools/migrate"
)

func TestV54MigrationIsIdempotentForAlreadyMigratedAppSnippets(t *testing.T) {
	files := map[string]string{
		"app.go": `package simapp

import (
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
)

func f() {
	_ = authtx.NewTxConfig
	app.ModuleManager.SetOrderEndBlockers(banktypes.ModuleName, govtypes.ModuleName)
	govkeeper.NewKeeper(cdc, storeService, acctKeeper, bankKeeper, distrKeeper, router, config, authority, govkeeper.NewDefaultCalculateVoteResultsAndVotingPower(stakingKeeper))
}
`,
	}

	dir := writeTestFiles(t, files)
	runV54Migration(t, dir)

	appGo := readTestFile(t, filepath.Join(dir, "app.go"))
	if strings.Contains(appGo, "authauthtx.") {
		t.Fatalf("unexpected double-rewritten auth tx alias:\n%s", appGo)
	}
	if strings.Count(appGo, "banktypes.ModuleName") != 1 {
		t.Fatalf("expected banktypes.ModuleName to appear once, got:\n%s", appGo)
	}
	if strings.Count(appGo, "NewDefaultCalculateVoteResultsAndVotingPower") != 1 {
		t.Fatalf("expected govkeeper tally helper to appear once, got:\n%s", appGo)
	}
}

func TestV54MigrationRewritesAnteWrapperAndParamsProto(t *testing.T) {
	files := map[string]string{
		"app.go": `package simapp

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

func (app *SimApp) setAnteHandler(txConfig client.TxConfig) {
	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			ante.HandlerOptions{
				SignModeHandler: txConfig.SignModeHandler(),
			},
			&app.CircuitKeeper,
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
}
`,
		"params/proto.go": `package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

func MakeTestEncodingConfig() EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          authtx.NewTxConfig(codec, authtx.DefaultSignModes),
		Amino:             cdc,
	}
}
`,
		"simd/cmd/root.go": `package cmd

import (
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
)

func f(initClientCtx client.Context) {
	enabledSignModes := append(authtx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
	txConfigOpts := authtx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: authtxconfig.NewGRPCCoinMetadataQueryFn(initClientCtx),
	}
	_, _ = authtx.NewTxConfigWithOptions(initClientCtx.Codec, txConfigOpts)
}
`,
	}

	dir := writeTestFiles(t, files)
	runV54Migration(t, dir)

	appGo := readTestFile(t, filepath.Join(dir, "app.go"))
	if strings.Contains(appGo, "anteHandler, err := NewAnteHandler(") {
		t.Fatalf("expected custom ante wrapper call to be removed:\n%s", appGo)
	}
	if strings.Contains(appGo, "HandlerOptions{\n\t\t\tante.HandlerOptions{") {
		t.Fatalf("expected custom handler options wrapper to be removed:\n%s", appGo)
	}
	if strings.Contains(appGo, "CircuitKeeper") {
		t.Fatalf("expected circuit keeper ante argument to be removed:\n%s", appGo)
	}
	if !strings.Contains(appGo, "ante.NewAnteHandler(") {
		t.Fatalf("expected direct ante handler call:\n%s", appGo)
	}

	paramsProto := readTestFile(t, filepath.Join(dir, "params/proto.go"))
	if !strings.Contains(paramsProto, `"github.com/cosmos/cosmos-sdk/x/auth/tx"`) {
		t.Fatalf("expected params/proto.go to import x/auth/tx:\n%s", paramsProto)
	}
	if strings.Contains(paramsProto, "authtx.") {
		t.Fatalf("expected params/proto.go to use tx alias, got:\n%s", paramsProto)
	}
	if !strings.Contains(paramsProto, "tx.NewTxConfig(codec, tx.DefaultSignModes)") {
		t.Fatalf("expected params/proto.go tx config usage to be normalized:\n%s", paramsProto)
	}

	rootGo := readTestFile(t, filepath.Join(dir, "simd/cmd/root.go"))
	if !strings.Contains(rootGo, `"github.com/cosmos/cosmos-sdk/x/auth/tx"`) {
		t.Fatalf("expected simd/cmd/root.go to import x/auth/tx:\n%s", rootGo)
	}
	if strings.Contains(rootGo, "authtx.") {
		t.Fatalf("expected simd/cmd/root.go to use tx alias, got:\n%s", rootGo)
	}
	if !strings.Contains(rootGo, "tx.NewTxConfigWithOptions") {
		t.Fatalf("expected simd/cmd/root.go tx config usage to be normalized:\n%s", rootGo)
	}
}

func TestPrepareTargetSDKVersion(t *testing.T) {
	t.Run("accepts v0.53 input", func(t *testing.T) {
		dir := writeTestFiles(t, map[string]string{
			"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.53.6
`,
		})

		if err := prepareTargetSDKVersion(dir); err != nil {
			t.Fatalf("expected v0.53 input to pass validation: %v", err)
		}
	})

	t.Run("bridges pre v0.53 input to v0.53.6", func(t *testing.T) {
		dir := writeTestFiles(t, map[string]string{
			"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.50.11

replace github.com/cosmos/cosmos-sdk => github.com/dydxprotocol/cosmos-sdk v0.50.6
replace github.com/cometbft/cometbft => github.com/dydxprotocol/cometbft v0.38.6
replace github.com/cosmos/iavl => github.com/dydxprotocol/iavl v1.1.1
replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.18.0
replace cosmossdk.io/x/evidence => cosmossdk.io/x/evidence v0.1.0
`,
		})

		if err := prepareTargetSDKVersion(dir); err != nil {
			t.Fatalf("expected v0.50 input to be bridged: %v", err)
		}

		goMod := readTestFile(t, filepath.Join(dir, "go.mod"))
		if !strings.Contains(goMod, "github.com/cosmos/cosmos-sdk v0.53.6") {
			t.Fatalf("expected go.mod to be bridged to v0.53.6, got:\n%s", goMod)
		}
		if !strings.Contains(goMod, "replace github.com/cosmos/cosmos-sdk => github.com/dydxprotocol/cosmos-sdk v0.50.6") {
			t.Fatalf("expected non-local cosmos-sdk replace to be preserved for manual audit, got:\n%s", goMod)
		}
		if !strings.Contains(goMod, "replace github.com/cometbft/cometbft => github.com/dydxprotocol/cometbft v0.38.6") {
			t.Fatalf("expected non-local cometbft replace to be preserved for manual audit, got:\n%s", goMod)
		}
	})

	t.Run("drops stale sdk replace for v0.53 input", func(t *testing.T) {
		dir := writeTestFiles(t, map[string]string{
			"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.53.6

replace github.com/cosmos/cosmos-sdk => github.com/dydxprotocol/cosmos-sdk v0.50.6
replace github.com/cometbft/cometbft => github.com/dydxprotocol/cometbft v0.38.6
replace github.com/cosmos/iavl => github.com/dydxprotocol/iavl v1.1.1
replace github.com/prometheus/common => github.com/prometheus/common v0.47.0
replace cosmossdk.io/x/upgrade => cosmossdk.io/x/upgrade v0.1.1
`,
		})

		if err := prepareTargetSDKVersion(dir); err != nil {
			t.Fatalf("expected v0.53 input with stale replace to pass validation: %v", err)
		}

		goMod := readTestFile(t, filepath.Join(dir, "go.mod"))
		if !strings.Contains(goMod, "replace github.com/cosmos/cosmos-sdk => github.com/dydxprotocol/cosmos-sdk v0.50.6") {
			t.Fatalf("expected non-local cosmos-sdk replace to be preserved for manual audit, got:\n%s", goMod)
		}
		if !strings.Contains(goMod, "replace cosmossdk.io/x/upgrade => cosmossdk.io/x/upgrade v0.1.1") {
			t.Fatalf("expected non-local upgrade replace to be preserved for manual audit, got:\n%s", goMod)
		}
	})

	t.Run("drops local replaces for stale modules", func(t *testing.T) {
		dir := writeTestFiles(t, map[string]string{
			"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.53.6

replace github.com/cosmos/cosmos-sdk => ../cosmos-sdk
replace github.com/cometbft/cometbft => ../cometbft
`,
		})

		if err := prepareTargetSDKVersion(dir); err != nil {
			t.Fatalf("expected v0.53 input with local replaces to pass validation: %v", err)
		}

		goMod := readTestFile(t, filepath.Join(dir, "go.mod"))
		if strings.Contains(goMod, "=> ../cosmos-sdk") || strings.Contains(goMod, "=> ../cometbft") {
			t.Fatalf("expected local stale replaces to be removed, got:\n%s", goMod)
		}
	})

	t.Run("rejects pre v0.50 input", func(t *testing.T) {
		dir := writeTestFiles(t, map[string]string{
			"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.49.8
`,
		})

		err := prepareTargetSDKVersion(dir)
		if err == nil {
			t.Fatal("expected validation error for v0.49 input")
		}
		if !strings.Contains(err.Error(), "supports v0.50.x through v0.53.x inputs") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects already migrated input", func(t *testing.T) {
		dir := writeTestFiles(t, map[string]string{
			"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.54.0-rc.1
`,
		})

		err := prepareTargetSDKVersion(dir)
		if err == nil {
			t.Fatal("expected validation error for v0.54 input")
		}
		if !strings.Contains(err.Error(), "already-migrated targets") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestV54MigrationRemovesCrisisWiring(t *testing.T) {
	dir := writeTestFiles(t, map[string]string{
		"app/app.go": `package app

import (
	"github.com/spf13/cast"
	crisis "github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type App struct {
	CrisisKeeper *crisiskeeper.Keeper
	ModuleManager interface{
		RegisterInvariants(any)
	}
}

func f(app *App) {
	keys := storetypes.NewKVStoreKeys(crisistypes.StoreKey)
	_ = keys
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))
	app.ModuleManager = module.NewManager(crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.getSubspace(crisistypes.ModuleName)))
	app.ModuleManager.RegisterInvariants(app.CrisisKeeper)
	paramsKeeper.Subspace(crisistypes.ModuleName)
}
`,
		"app/basic_manager/basic_manager.go": `package basic_manager

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	crisis "github.com/cosmos/cosmos-sdk/x/crisis"
)

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	crisis.AppModuleBasic{},
)
`,
		"cmd/root.go": `package cmd

import (
	"github.com/spf13/cobra"
	crisis "github.com/cosmos/cosmos-sdk/x/crisis"
)

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}
`,
		"app/msgs/internal_msgs.go": `package msgs

import (
	crisis "github.com/cosmos/cosmos-sdk/x/crisis/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var InternalMsgSamplesDefault = map[string]sdk.Msg{
	"/cosmos.crisis.v1beta1.MsgUpdateParams":         &crisis.MsgUpdateParams{},
	"/cosmos.crisis.v1beta1.MsgUpdateParamsResponse": nil,
}
`,
		"app/msgs/unsupported_msgs.go": `package msgs

import (
	crisis "github.com/cosmos/cosmos-sdk/x/crisis/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var UnsupportedMsgSamples = map[string]sdk.Msg{
	"/cosmos.crisis.v1beta1.MsgVerifyInvariant":         &crisis.MsgVerifyInvariant{},
	"/cosmos.crisis.v1beta1.MsgVerifyInvariantResponse": nil,
}
`,
		"lib/ante/internal_msg.go": `package ante

import crisis "github.com/cosmos/cosmos-sdk/x/crisis/types"

func IsInternalMsg(msg any) bool {
	switch msg.(type) {
	case
		*crisis.MsgUpdateParams:
		return true
	}
	return false
}
`,
		"lib/ante/unsupported_msgs.go": `package ante

import crisis "github.com/cosmos/cosmos-sdk/x/crisis/types"

func IsUnsupportedMsg(msg any) bool {
	switch msg.(type) {
	case
		*crisis.MsgVerifyInvariant:
		return true
	}
	return false
}
`,
		"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.53.6
`,
	})

	runV54Migration(t, dir)

	appGo := readTestFile(t, filepath.Join(dir, "app/app.go"))
	if strings.Contains(appGo, "CrisisKeeper") {
		t.Fatalf("expected crisis keeper field/usage to be removed, got:\n%s", appGo)
	}
	if strings.Contains(appGo, "crisistypes.StoreKey") {
		t.Fatalf("expected crisis store key to be removed, got:\n%s", appGo)
	}
	if strings.Contains(appGo, "crisis.NewAppModule") {
		t.Fatalf("expected crisis app module wiring to be removed, got:\n%s", appGo)
	}
	if strings.Contains(appGo, "crisis.FlagSkipGenesisInvariants") {
		t.Fatalf("expected crisis init flag wiring to be removed, got:\n%s", appGo)
	}

	basicManager := readTestFile(t, filepath.Join(dir, "app/basic_manager/basic_manager.go"))
	if strings.Contains(basicManager, "crisis.AppModuleBasic") {
		t.Fatalf("expected crisis basic module registration to be removed, got:\n%s", basicManager)
	}

	rootGo := readTestFile(t, filepath.Join(dir, "cmd/root.go"))
	if strings.Contains(rootGo, "crisis.AddModuleInitFlags") {
		t.Fatalf("expected crisis init flag registration to be removed, got:\n%s", rootGo)
	}

	internalMsgs := readTestFile(t, filepath.Join(dir, "app/msgs/internal_msgs.go"))
	if strings.Contains(internalMsgs, "/cosmos.crisis.v1beta1.MsgUpdateParams") {
		t.Fatalf("expected crisis internal msg entries to be removed, got:\n%s", internalMsgs)
	}

	unsupportedMsgs := readTestFile(t, filepath.Join(dir, "app/msgs/unsupported_msgs.go"))
	if strings.Contains(unsupportedMsgs, "/cosmos.crisis.v1beta1.MsgVerifyInvariant") {
		t.Fatalf("expected crisis unsupported msg entries to be removed, got:\n%s", unsupportedMsgs)
	}

	anteInternal := readTestFile(t, filepath.Join(dir, "lib/ante/internal_msg.go"))
	if strings.Contains(anteInternal, "crisis.MsgUpdateParams") {
		t.Fatalf("expected crisis ante internal msg case to be removed, got:\n%s", anteInternal)
	}

	anteUnsupported := readTestFile(t, filepath.Join(dir, "lib/ante/unsupported_msgs.go"))
	if strings.Contains(anteUnsupported, "crisis.MsgVerifyInvariant") {
		t.Fatalf("expected crisis ante unsupported msg case to be removed, got:\n%s", anteUnsupported)
	}
}

func TestV54MigrationRewritesCompatibilityHotspots(t *testing.T) {
	dir := writeTestFiles(t, map[string]string{
		"app/msgs/internal_msgs.go": `package msgs

import (
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var InternalMsgSamplesDefault = map[string]sdk.Msg{
	"/cosmos.staking.v1beta1.MsgSetProposers":         &staking.MsgSetProposers{},
	"/cosmos.staking.v1beta1.MsgSetProposersResponse": nil,
}
`,
		"app/msgs/all_msgs.go": `package msgs

var AllTypeMessages = map[string]struct{}{
	"/cosmos.staking.v1beta1.MsgSetProposers":         {},
	"/cosmos.staking.v1beta1.MsgSetProposersResponse": {},
}
`,
		"lib/ante/internal_msg.go": `package ante

import staking "github.com/cosmos/cosmos-sdk/x/staking/types"

func IsInternalMsg(msg any) bool {
	switch msg.(type) {
	case
		*staking.MsgSetProposers,
		*staking.MsgDelegate:
		return true
	}
	return false
}
`,
		"x/foo/client/cli/tx.go": `package cli

import "github.com/cosmos/cosmos-sdk/client/tx"

func f(txf tx.Factory, value any) tx.Factory {
	return txf.WithNonCriticalExtensionOptions(value)
}
`,
		"x/foo/module.go": `package foo

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

func f(ctx context.Context) {
	defer telemetry.ModuleMeasureSince("foo", time.Now(), telemetry.MetricKeyPrecommiter)
	defer telemetry.ModuleMeasureSince("foo", time.Now(), telemetry.MetricKeyPrepareCheckStater)
}
`,
		"testutil/ante/testutil.go": `package ante

import "github.com/golang/mock/gomock"

func f(ctrl *gomock.Controller) {}
`,
		"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.53.6
`,
	})

	runV54Migration(t, dir)

	internalMsgs := readTestFile(t, filepath.Join(dir, "app/msgs/internal_msgs.go"))
	if strings.Contains(internalMsgs, "MsgSetProposers") {
		t.Fatalf("expected staking MsgSetProposers entries to be removed, got:\n%s", internalMsgs)
	}

	allMsgs := readTestFile(t, filepath.Join(dir, "app/msgs/all_msgs.go"))
	if strings.Contains(allMsgs, "MsgSetProposers") {
		t.Fatalf("expected all-msg registry entries to be removed, got:\n%s", allMsgs)
	}

	anteInternal := readTestFile(t, filepath.Join(dir, "lib/ante/internal_msg.go"))
	if strings.Contains(anteInternal, "MsgSetProposers") {
		t.Fatalf("expected staking MsgSetProposers switch cases to be removed, got:\n%s", anteInternal)
	}

	cliTx := readTestFile(t, filepath.Join(dir, "x/foo/client/cli/tx.go"))
	if strings.Contains(cliTx, "WithNonCriticalExtensionOptions") {
		t.Fatalf("expected old tx factory helper to be rewritten, got:\n%s", cliTx)
	}
	if !strings.Contains(cliTx, "WithExtensionOptions") {
		t.Fatalf("expected tx factory helper rewrite, got:\n%s", cliTx)
	}

	moduleGo := readTestFile(t, filepath.Join(dir, "x/foo/module.go"))
	if strings.Contains(moduleGo, "MetricKeyPrecommiter") || strings.Contains(moduleGo, "MetricKeyPrepareCheckStater") {
		t.Fatalf("expected stale telemetry constants to be removed, got:\n%s", moduleGo)
	}
	if !strings.Contains(moduleGo, `"precommitter"`) || !strings.Contains(moduleGo, `"prepare_check_stater"`) {
		t.Fatalf("expected telemetry constants to be rewritten to string keys, got:\n%s", moduleGo)
	}

	testutil := readTestFile(t, filepath.Join(dir, "testutil/ante/testutil.go"))
	if strings.Contains(testutil, "github.com/golang/mock/gomock") {
		t.Fatalf("expected gomock import to be rewritten, got:\n%s", testutil)
	}
	if !strings.Contains(testutil, "go.uber.org/mock/gomock") {
		t.Fatalf("expected gomock import rewrite, got:\n%s", testutil)
	}
}

func TestV54MigrationPreservesCustomAnteHandler(t *testing.T) {
	dir := writeTestFiles(t, map[string]string{
		"app.go": `package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

func (app *App) buildAnteHandler(txConfig client.TxConfig) {
	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				SignModeHandler: txConfig.SignModeHandler(),
			},
			ClobKeeper: app.ClobKeeper,
		},
	)
	if err != nil {
		panic(err)
	}
	_ = anteHandler
}
`,
		"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.53.6
`,
	})

	runV54Migration(t, dir)

	appGo := readTestFile(t, filepath.Join(dir, "app.go"))
	if strings.Contains(appGo, "ante.NewAnteHandler(") {
		t.Fatalf("expected custom ante wrapper to be preserved, got:\n%s", appGo)
	}
	if !strings.Contains(appGo, "NewAnteHandler(") {
		t.Fatalf("expected local ante wrapper call to remain, got:\n%s", appGo)
	}
}

func TestV54MigrationHandlesPointerReturningEpochsKeeper(t *testing.T) {
	dir := writeTestFiles(t, map[string]string{
		"app/app.go": `package app

import epochsmodulekeeper "example.com/chain/x/epochs/keeper"

type App struct {
	EpochsKeeper epochsmodulekeeper.Keeper
}

func f(app *App) {
	app.EpochsKeeper = *epochsmodulekeeper.NewKeeper(storeKey)
}
`,
		"go.mod": `module example.com/app

go 1.25.0

require github.com/cosmos/cosmos-sdk v0.53.6
`,
	})

	runV54Migration(t, dir)

	appGo := readTestFile(t, filepath.Join(dir, "app/app.go"))
	if !strings.Contains(appGo, "EpochsKeeper *epochsmodulekeeper.Keeper") {
		t.Fatalf("expected App.EpochsKeeper to become a pointer, got:\n%s", appGo)
	}
	if strings.Contains(appGo, "= *epochsmodulekeeper.NewKeeper(") {
		t.Fatalf("expected explicit dereference to be removed, got:\n%s", appGo)
	}
	if !strings.Contains(appGo, "app.EpochsKeeper = epochsmodulekeeper.NewKeeper(") {
		t.Fatalf("expected pointer-returning constructor assignment, got:\n%s", appGo)
	}
}

func runV54Migration(t *testing.T, dir string) {
	t.Helper()

	args := migration.MigrateArgs{
		GoModRemoval:           removals,
		GoModAddition:          additions,
		GoModReplacements:      replacements,
		GoModUpdates:           moduleUpdates,
		StripLocalPathReplaces: true,
		ImportUpdates:          importReplacements,
		ImportWarnings:         importWarnings,
		TypeUpdates:            typeReplacements,
		FieldRemovals:          fieldRemovals,
		FieldModifications:     fieldModifications,
		ArgUpdates:             callUpdates,
		ArgSurgeries:           argSurgeries,
		CallArgEdits:           callArgEdits,
		ComplexUpdates:         complexUpdates,
		StatementRemovals:      statementRemovals,
		MapEntryRemovals:       mapEntryRemovals,
		TextReplacements:       textReplacements,
		FileRemovals:           fileRemovals,
	}

	if err := migration.Migrate(dir, args); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
}

func writeTestFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", fullPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", fullPath, err)
		}
	}

	return dir
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()

	bz, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return string(bz)
}
