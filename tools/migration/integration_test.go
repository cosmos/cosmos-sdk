package migration

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationV53ToV54AppGo(t *testing.T) {
	v53AppGo := `package simapp

import (
	"cosmossdk.io/log"
	"cosmossdk.io/x/circuit"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	circuittypes "cosmossdk.io/x/circuit/types"
	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	nftmodule "cosmossdk.io/x/nft/module"
	"cosmossdk.io/x/upgrade"
	epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var maccPerms = map[string][]string{
	"auth":         {"minter", "burner"},
	nft.ModuleName: nil,
}

type SimApp struct {
	CircuitKeeper circuitkeeper.Keeper
	NFTKeeper     nftkeeper.Keeper
	EpochsKeeper  epochskeeper.Keeper
	GovKeeper     govkeeper.Keeper
}

func NewSimApp() *SimApp {
	app := &SimApp{}

	keys := storetypes.NewKVStoreKeys(
		"auth",
		circuittypes.StoreKey,
		nftkeeper.StoreKey,
		"bank",
	)

	app.CircuitKeeper = circuitkeeper.NewKeeper(env, cdc, authority)
	app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)
	app.NFTKeeper = nftkeeper.NewKeeper(storeService, cdc)
	app.GovKeeper = govkeeper.NewKeeper(cdc, storeService, acctKeeper, bankKeeper, stakingKeeper, distrKeeper, router, config, authority)
	app.EpochsKeeper = epochskeeper.NewKeeper(storeService, cdc)
	app.EpochsKeeper.SetHooks(hooks)

	app.ModuleManager = module.NewManager(
		circuit.NewAppModule(cdc, app.CircuitKeeper),
		nftmodule.NewAppModule(cdc, app.NFTKeeper),
	)

	app.ModuleManager.SetOrderBeginBlockers(
		circuittypes.ModuleName,
		nft.ModuleName,
		"bank",
	)

	app.ModuleManager.SetOrderEndBlockers(
		circuittypes.ModuleName,
		nft.ModuleName,
		"staking",
	)

	return app
}
`

	importReplacements := []ImportReplacement{
		{Old: "cosmossdk.io/log", New: "cosmossdk.io/log/v2", AllPackages: false},
		{Old: "cosmossdk.io/x/feegrant", New: "github.com/cosmos/cosmos-sdk/x/feegrant", AllPackages: true},
		{Old: "cosmossdk.io/x/evidence", New: "github.com/cosmos/cosmos-sdk/x/evidence", AllPackages: true},
		{Old: "cosmossdk.io/x/upgrade", New: "github.com/cosmos/cosmos-sdk/x/upgrade", AllPackages: true},
		{Old: "cosmossdk.io/x/circuit", New: "github.com/cosmos/cosmos-sdk/contrib/x/circuit", AllPackages: true},
		{Old: "cosmossdk.io/x/nft", New: "github.com/cosmos/cosmos-sdk/contrib/x/nft", AllPackages: true},
	}

	fieldMods := []StructFieldModification{
		{StructName: "SimApp", FieldName: "EpochsKeeper", MakePointer: true},
	}

	callEdits := []CallArgRemoval{
		{
			MethodName: "SetOrderEndBlockers",
			ArgsToAdd:  []ArgAddition{{Position: 0, Expr: "banktypes.ModuleName"}},
		},
	}

	argSurgeries := []ArgSurgeryWithAST{
		{
			ImportPath:  "github.com/cosmos/cosmos-sdk/x/gov/keeper",
			FuncName:    "NewKeeper",
			OldArgCount: -1,
			Transform: func(pkgAlias string, args []ast.Expr) []ast.Expr {
				if len(args) < 9 {
					return args
				}
				stakingKeeper := args[4]
				newArgs := make([]ast.Expr, 0, 9)
				newArgs = append(newArgs, args[0:4]...)
				newArgs = append(newArgs, args[5:9]...)
				newArgs = append(newArgs, &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: pkgAlias},
						Sel: &ast.Ident{Name: "NewDefaultCalculateVoteResultsAndVotingPower"},
					},
					Args: []ast.Expr{stakingKeeper},
				})
				return newArgs
			},
		},
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "app.go", v53AppGo, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if _, err := updateImports(node, importReplacements); err != nil {
		t.Fatalf("updateImports error: %v", err)
	}
	if _, err := updateStructFieldModifications(node, fieldMods); err != nil {
		t.Fatalf("updateStructFieldModifications error: %v", err)
	}
	if _, err := updateCallArgRemovals(node, callEdits); err != nil {
		t.Fatalf("updateCallArgRemovals error: %v", err)
	}
	if _, err := updateArgSurgeryAST(node, argSurgeries); err != nil {
		t.Fatalf("updateArgSurgeryAST error: %v", err)
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		t.Fatalf("print error: %v", err)
	}
	output := buf.String()

	textReplacements := []TextReplacement{
		{Old: "app.EpochsKeeper = epochskeeper.NewKeeper(", New: "epochsKeeper := epochskeeper.NewKeeper("},
		{Old: "\tapp.EpochsKeeper.SetHooks(", New: "\tapp.EpochsKeeper = &epochsKeeper\n\n\tapp.EpochsKeeper.SetHooks("},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "app.go")
	if err := os.WriteFile(tmpFile, []byte(output), 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if _, err := applyTextReplacements(tmpFile, textReplacements); err != nil {
		t.Fatalf("applyTextReplacements error: %v", err)
	}

	result, _ := os.ReadFile(tmpFile)
	output = string(result)

	assertContains(t, output, `"cosmossdk.io/log/v2"`, "log import should be rewritten to v2")
	assertContains(t, output, `"github.com/cosmos/cosmos-sdk/contrib/x/circuit"`, "circuit import should be rewritten to contrib")
	assertContains(t, output, `"github.com/cosmos/cosmos-sdk/contrib/x/nft"`, "nft import should be rewritten to contrib")
	assertMissing(t, output, `"cosmossdk.io/x/circuit"`, "old circuit import should be gone")
	assertMissing(t, output, `"cosmossdk.io/x/nft"`, "old nft import should be gone")

	assertContains(t, output, "CircuitKeeper", "CircuitKeeper field should be preserved")
	assertContains(t, output, "circuitkeeper.Keeper", "CircuitKeeper type should be preserved")
	assertContains(t, output, "NFTKeeper", "NFTKeeper field should be preserved")
	assertContains(t, output, "nftkeeper.Keeper", "NFTKeeper type should be preserved")
	assertContains(t, output, "circuitkeeper.NewKeeper", "circuit keeper init should be preserved")
	assertContains(t, output, "SetCircuitBreaker", "SetCircuitBreaker should be preserved")
	assertContains(t, output, "nftkeeper.NewKeeper", "nft keeper init should be preserved")
	assertContains(t, output, "nft.ModuleName", "nft.ModuleName should be preserved")
	assertContains(t, output, "circuittypes.StoreKey", "circuit store key should be preserved")
	assertContains(t, output, "nftkeeper.StoreKey", "nft store key should be preserved")
	assertContains(t, output, "circuit.NewAppModule", "circuit AppModule should be preserved")
	assertContains(t, output, "nftmodule.NewAppModule", "nft AppModule should be preserved")

	assertContains(t, output, "banktypes.ModuleName", "banktypes.ModuleName should be added to EndBlockers")
	assertContains(t, output, "NewDefaultCalculateVoteResultsAndVotingPower", "govkeeper should wrap stakingKeeper")
	assertContains(t, output, "stakingKeeper", "wrapped stakingKeeper should still be present as the new helper argument")
	assertMissing(t, output, "stakingKeeper, distrKeeper", "stakingKeeper should no longer be in main args before distrKeeper")
	assertContains(t, output, "*epochskeeper.Keeper", "EpochsKeeper should be pointer")
	assertContains(t, output, "epochsKeeper := epochskeeper.NewKeeper(", "EpochsKeeper should use local var")
	assertContains(t, output, "app.EpochsKeeper = &epochsKeeper", "EpochsKeeper pointer assignment should be inserted")
}

func TestIntegrationV53ToV54AppConfig(t *testing.T) {
	postASTContent := `package simapp

import (
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	circuitmodulev1 "cosmossdk.io/api/cosmos/circuit/module/v1"
	nftmodulev1 "cosmossdk.io/api/cosmos/nft/module/v1"
	"cosmossdk.io/core/appconfig"
	_ "github.com/cosmos/cosmos-sdk/contrib/x/circuit" // import for side-effects
	circuittypes "github.com/cosmos/cosmos-sdk/contrib/x/circuit/types"
	"github.com/cosmos/cosmos-sdk/contrib/x/nft"
	_ "github.com/cosmos/cosmos-sdk/contrib/x/nft/module" // import for side-effects
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	moduleAccPerms = []*authmodulev1.ModuleAccountPermission{
		{Account: authtypes.FeeCollectorName},
		{Account: nft.ModuleName},
		{Account: govtypes.ModuleName},
	}

	blockAccAddrs = []string{
		authtypes.FeeCollectorName,
		nft.ModuleName,
		stakingtypes.BondedPoolName,
	}

	ModuleConfig = []*appv1alpha1.ModuleConfig{
		{
			Name: "runtime",
			Config: appconfig.WrapAny(&runtimev1alpha1.Module{
				InitGenesis: []string{
					authtypes.ModuleName,
					banktypes.ModuleName,
					nft.ModuleName,
					circuittypes.ModuleName,
				},
			}),
		},
		{
			Name:   nft.ModuleName,
			Config: appconfig.WrapAny(&nftmodulev1.Module{}),
		},
		{
			Name:   circuittypes.ModuleName,
			Config: appconfig.WrapAny(&circuitmodulev1.Module{}),
		},
	}
)
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "app_config.go")
	if err := os.WriteFile(tmpFile, []byte(postASTContent), 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	modified, err := applyTextReplacements(tmpFile, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if modified {
		t.Error("expected no app_config.go text replacements for preserved contrib modules")
	}

	result, _ := os.ReadFile(tmpFile)
	output := string(result)

	assertContains(t, output, "circuitmodulev1", "circuitmodulev1 API import should be preserved")
	assertContains(t, output, "nftmodulev1", "nftmodulev1 API import should be preserved")
	assertContains(t, output, "contrib/x/circuit", "circuit contrib imports should be preserved")
	assertContains(t, output, "contrib/x/nft", "nft contrib imports should be preserved")
	assertContains(t, output, "nft.ModuleName", "nft.ModuleName should be preserved")
	assertContains(t, output, "circuittypes.ModuleName", "circuittypes.ModuleName should be preserved")
	assertContains(t, output, "nftmodulev1.Module", "nft ModuleConfig should be preserved")
	assertContains(t, output, "circuitmodulev1.Module", "circuit ModuleConfig should be preserved")
}

func assertContains(t *testing.T, output, substr, msg string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("FAIL: %s — output should contain %q", msg, substr)
	}
}

func assertMissing(t *testing.T, output, substr, msg string) {
	t.Helper()
	if strings.Contains(output, substr) {
		t.Errorf("FAIL: %s — output should NOT contain %q", msg, substr)
	}
}
