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
