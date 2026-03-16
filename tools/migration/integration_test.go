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

// TestIntegrationV53ToV54AppGo simulates running all AST transformations on a
// simplified v53-style app.go and verifies the output matches v54 expectations.
func TestIntegrationV53ToV54AppGo(t *testing.T) {
	// Simplified v53-style app.go with all the patterns we need to transform
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
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var maccPerms = map[string][]string{
	"auth":           {"minter", "burner"},
	nft.ModuleName:   nil,
	"staking":        {"burner"},
}

type SimApp struct {
	BankKeeper    int
	CircuitKeeper circuitkeeper.Keeper
	NFTKeeper     nftkeeper.Keeper
	GroupKeeper   groupkeeper.Keeper
	EpochsKeeper  epochskeeper.Keeper
	GovKeeper     govkeeper.Keeper
}

func NewSimApp() *SimApp {
	app := &SimApp{}

	keys := storetypes.NewKVStoreKeys(
		"auth",
		circuittypes.StoreKey,
		nftkeeper.StoreKey,
		group.StoreKey,
		"bank",
	)

	app.CircuitKeeper = circuitkeeper.NewKeeper(env, cdc, authority)
	app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)

	app.NFTKeeper = nftkeeper.NewKeeper(storeService, cdc)

	groupConfig := group.DefaultConfig()
	app.GroupKeeper = groupkeeper.NewKeeper(storeService, cdc, groupConfig)

	app.GovKeeper = govkeeper.NewKeeper(cdc, storeService, acctKeeper, bankKeeper, stakingKeeper, distrKeeper, router, config, authority)

	app.EpochsKeeper = epochskeeper.NewKeeper(storeService, cdc)

	app.EpochsKeeper.SetHooks(hooks)

	app.ModuleManager = module.NewManager(
		circuit.NewAppModule(cdc, app.CircuitKeeper),
		nftmodule.NewAppModule(cdc, app.NFTKeeper),
		groupmodule.NewAppModule(cdc, app.GroupKeeper),
	)

	app.ModuleManager.SetOrderBeginBlockers(
		circuittypes.ModuleName,
		nft.ModuleName,
		group.ModuleName,
		"bank",
	)

	app.ModuleManager.SetOrderEndBlockers(
		group.ModuleName,
		circuittypes.ModuleName,
		nft.ModuleName,
		"staking",
	)

	app.ModuleManager.SetOrderInitGenesis(
		circuittypes.ModuleName,
		nft.ModuleName,
		group.ModuleName,
		"bank",
	)

	return app
}
`

	// --- Import rewrites ---
	importReplacements := []ImportReplacement{
		{Old: "cosmossdk.io/log", New: "cosmossdk.io/log/v2", AllPackages: false},
		{Old: "cosmossdk.io/x/feegrant", New: "github.com/cosmos/cosmos-sdk/x/feegrant", AllPackages: true},
		{Old: "cosmossdk.io/x/evidence", New: "github.com/cosmos/cosmos-sdk/x/evidence", AllPackages: true},
		{Old: "cosmossdk.io/x/upgrade", New: "github.com/cosmos/cosmos-sdk/x/upgrade", AllPackages: true},
		{Old: "cosmossdk.io/x/circuit", New: "github.com/cosmos/cosmos-sdk/contrib/x/circuit", AllPackages: true},
		{Old: "cosmossdk.io/x/nft", New: "github.com/cosmos/cosmos-sdk/contrib/x/nft", AllPackages: true},
	}

	// --- Struct field removals ---
	fieldRemovals := []StructFieldRemoval{
		{StructName: "SimApp", FieldName: "CircuitKeeper"},
		{StructName: "SimApp", FieldName: "NFTKeeper"},
		{StructName: "SimApp", FieldName: "GroupKeeper"},
	}

	// --- Struct field modifications ---
	fieldMods := []StructFieldModification{
		{StructName: "SimApp", FieldName: "EpochsKeeper", MakePointer: true},
	}

	// --- Statement removals ---
	stmtRemovals := []StatementRemoval{
		{AssignTarget: "app.CircuitKeeper", IncludeFollowing: 1},
		{CallPattern: "app.BaseApp.SetCircuitBreaker"},
		{AssignTarget: "app.NFTKeeper"},
		{AssignTarget: "app.GroupKeeper", IncludePrecedingAssign: "groupConfig"},
	}

	// --- Map entry removals ---
	mapRemovals := []MapEntryRemoval{
		{MapVarName: "maccPerms", KeysToRemove: []string{"nft.ModuleName"}},
	}

	// --- Call arg removals ---
	callEdits := []CallArgRemoval{
		{FuncPattern: "storetypes.NewKVStoreKeys", ArgsToRemove: []string{"circuittypes.StoreKey", "nftkeeper.StoreKey", "group.StoreKey"}},
		{FuncPattern: "module.NewManager", ArgsToRemove: []string{"circuit.NewAppModule(...)", "nftmodule.NewAppModule(...)", "groupmodule.NewAppModule(...)"}},
		{MethodName: "SetOrderBeginBlockers", ArgsToRemove: []string{"circuittypes.ModuleName", "nft.ModuleName", "group.ModuleName"}},
		{
			MethodName:   "SetOrderEndBlockers",
			ArgsToRemove: []string{"group.ModuleName", "circuittypes.ModuleName", "nft.ModuleName"},
			ArgsToAdd:    []ArgAddition{{Position: 0, Expr: "banktypes.ModuleName"}},
		},
		{MethodName: "SetOrderInitGenesis", ArgsToRemove: []string{"circuittypes.ModuleName", "nft.ModuleName", "group.ModuleName"}},
	}

	// --- Arg surgery ---
	argSurgeries := []ArgSurgeryWithAST{
		{
			ImportPath:  "github.com/cosmos/cosmos-sdk/x/gov/keeper",
			FuncName:    "NewKeeper",
			OldArgCount: -1,
			Transform: func(args []ast.Expr) []ast.Expr {
				if len(args) < 9 {
					return args
				}
				stakingKeeper := args[4]
				newArgs := make([]ast.Expr, 0, 9)
				newArgs = append(newArgs, args[0:4]...)
				newArgs = append(newArgs, args[5:9]...)
				newArgs = append(newArgs, &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "govkeeper"},
						Sel: &ast.Ident{Name: "NewDefaultCalculateVoteResultsAndVotingPower"},
					},
					Args: []ast.Expr{stakingKeeper},
				})
				return newArgs
			},
		},
	}

	// --- Parse ---
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "app.go", v53AppGo, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// --- Apply all AST transformations ---
	if _, err := updateImports(node, importReplacements); err != nil {
		t.Fatalf("updateImports error: %v", err)
	}
	if _, err := updateStructFieldRemovals(node, fieldRemovals); err != nil {
		t.Fatalf("updateStructFieldRemovals error: %v", err)
	}
	if _, err := updateStructFieldModifications(node, fieldMods); err != nil {
		t.Fatalf("updateStructFieldModifications error: %v", err)
	}
	if _, err := updateStatementRemovals(node, stmtRemovals); err != nil {
		t.Fatalf("updateStatementRemovals error: %v", err)
	}
	if _, err := updateMapEntryRemovals(node, mapRemovals); err != nil {
		t.Fatalf("updateMapEntryRemovals error: %v", err)
	}
	if _, err := updateCallArgRemovals(node, callEdits); err != nil {
		t.Fatalf("updateCallArgRemovals error: %v", err)
	}
	if _, err := updateArgSurgeryAST(node, argSurgeries); err != nil {
		t.Fatalf("updateArgSurgeryAST error: %v", err)
	}

	// --- Render ---
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		t.Fatalf("print error: %v", err)
	}
	output := buf.String()

	// --- Apply text replacements on rendered output ---
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

	// === ASSERTIONS ===

	// Import rewrites
	assertContains(t, output, `"cosmossdk.io/log/v2"`, "log import should be rewritten to v2")
	assertContains(t, output, `"github.com/cosmos/cosmos-sdk/x/feegrant"`, "feegrant import should be rewritten")
	assertContains(t, output, `"github.com/cosmos/cosmos-sdk/x/evidence"`, "evidence import should be rewritten")
	assertContains(t, output, `"github.com/cosmos/cosmos-sdk/x/upgrade"`, "upgrade import should be rewritten")
	assertContains(t, output, `"github.com/cosmos/cosmos-sdk/contrib/x/circuit"`, "circuit import should be rewritten to contrib")
	assertMissing(t, output, `"cosmossdk.io/x/circuit"`, "old circuit import should be gone")
	assertMissing(t, output, `"cosmossdk.io/x/nft"`, "old nft import should be gone")

	// Struct field removals
	assertMissing(t, output, "CircuitKeeper", "CircuitKeeper field should be removed")
	assertMissing(t, output, "NFTKeeper", "NFTKeeper field should be removed")
	assertMissing(t, output, "GroupKeeper", "GroupKeeper field should be removed")

	// Struct field modification
	assertContains(t, output, "*epochskeeper.Keeper", "EpochsKeeper should be pointer")

	// Statement removals
	assertMissing(t, output, "circuitkeeper.NewKeeper", "circuit keeper init should be removed")
	assertMissing(t, output, "SetCircuitBreaker", "SetCircuitBreaker should be removed")
	assertMissing(t, output, "nftkeeper.NewKeeper", "nft keeper init should be removed")
	assertMissing(t, output, "groupkeeper.NewKeeper", "group keeper init should be removed")
	assertMissing(t, output, "groupConfig", "groupConfig should be removed")

	// Map entry removal
	assertMissing(t, output, "nft.ModuleName", "nft.ModuleName should be removed from maccPerms")

	// Call arg removals — store keys
	assertMissing(t, output, "circuittypes.StoreKey", "circuit store key should be removed")
	assertMissing(t, output, "nftkeeper.StoreKey", "nft store key should be removed")
	assertMissing(t, output, "group.StoreKey", "group store key should be removed")

	// Call arg removals — module.NewManager
	assertMissing(t, output, "circuit.NewAppModule", "circuit AppModule should be removed")
	assertMissing(t, output, "nftmodule.NewAppModule", "nft AppModule should be removed")
	assertMissing(t, output, "groupmodule.NewAppModule", "group AppModule should be removed")

	// Call arg removals — ordering
	assertContains(t, output, "banktypes.ModuleName", "banktypes.ModuleName should be added to EndBlockers")

	// Arg surgery — govkeeper.NewKeeper
	assertContains(t, output, "NewDefaultCalculateVoteResultsAndVotingPower(stakingKeeper)", "govkeeper should have wrapped stakingKeeper")
	assertMissing(t, output, "stakingKeeper, distrKeeper", "stakingKeeper should no longer be in main args before distrKeeper")

	// Text replacements — EpochsKeeper
	assertContains(t, output, "epochsKeeper := epochskeeper.NewKeeper(", "EpochsKeeper should use local var")
	assertContains(t, output, "app.EpochsKeeper = &epochsKeeper", "EpochsKeeper pointer assignment should be inserted")

	t.Logf("=== Transformed output ===\n%s", output)
}

// TestIntegrationV53ToV54AppConfig tests the text replacement patterns that clean up
// app_config.go after AST transforms (import rewrites) have run.
func TestIntegrationV53ToV54AppConfig(t *testing.T) {
	// This is a simplified v53 app_config.go AFTER import rewriting has been applied.
	// Circuit/nft imports have been rewritten to contrib paths; group imports are unchanged.
	postASTContent := `package simapp

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	circuitmodulev1 "cosmossdk.io/api/cosmos/circuit/module/v1"
	groupmodulev1 "cosmossdk.io/api/cosmos/group/module/v1"
	nftmodulev1 "cosmossdk.io/api/cosmos/nft/module/v1"
	"cosmossdk.io/core/appconfig"
	_ "github.com/cosmos/cosmos-sdk/contrib/x/circuit" // import for side-effects
	circuittypes "github.com/cosmos/cosmos-sdk/contrib/x/circuit/types"
	"github.com/cosmos/cosmos-sdk/contrib/x/nft"
	_ "github.com/cosmos/cosmos-sdk/contrib/x/nft/module" // import for side-effects
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	_ "github.com/cosmos/cosmos-sdk/x/group/module" // import for side-effects
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
				EndBlockers: []string{
					govtypes.ModuleName,
					stakingtypes.ModuleName,
					group.ModuleName,
				},
				InitGenesis: []string{
					authtypes.ModuleName,
					banktypes.ModuleName,
					nft.ModuleName,
					group.ModuleName,
					circuittypes.ModuleName,
				},
				ExportGenesis: []string{
					authtypes.ModuleName,
					nft.ModuleName,
					group.ModuleName,
					circuittypes.ModuleName,
				},
			}),
		},
		{
			Name: group.ModuleName,
			Config: appconfig.WrapAny(&groupmodulev1.Module{
				MaxExecutionPeriod: durationpb.New(time.Second * 1209600),
				MaxMetadataLen:     255,
			}),
		},
		{
			Name:   nft.ModuleName,
			Config: appconfig.WrapAny(&nftmodulev1.Module{}),
		},
		{
			Name:   govtypes.ModuleName,
			Config: appconfig.WrapAny(&authmodulev1.Module{}),
		},
		{
			Name:   circuittypes.ModuleName,
			Config: appconfig.WrapAny(&circuitmodulev1.Module{}),
		},
	}
)
`

	// Text replacements scoped to app_config.go
	appConfigReplacements := []TextReplacement{
		// API imports
		{Old: "\tcircuitmodulev1 \"cosmossdk.io/api/cosmos/circuit/module/v1\"\n", New: "", FileMatch: "app_config.go"},
		{Old: "\tgroupmodulev1 \"cosmossdk.io/api/cosmos/group/module/v1\"\n", New: "", FileMatch: "app_config.go"},
		{Old: "\tnftmodulev1 \"cosmossdk.io/api/cosmos/nft/module/v1\"\n", New: "", FileMatch: "app_config.go"},
		// Circuit imports (post-rewrite)
		{Old: "\t_ \"github.com/cosmos/cosmos-sdk/contrib/x/circuit\" // import for side-effects\n", New: "", FileMatch: "app_config.go"},
		{Old: "\tcircuittypes \"github.com/cosmos/cosmos-sdk/contrib/x/circuit/types\"\n", New: "", FileMatch: "app_config.go"},
		// NFT imports (post-rewrite)
		{Old: "\t\"github.com/cosmos/cosmos-sdk/contrib/x/nft\"\n", New: "", FileMatch: "app_config.go"},
		{Old: "\t_ \"github.com/cosmos/cosmos-sdk/contrib/x/nft/module\" // import for side-effects\n", New: "", FileMatch: "app_config.go"},
		// Group imports
		{Old: "\t\"github.com/cosmos/cosmos-sdk/x/group\"\n", New: "", FileMatch: "app_config.go"},
		{Old: "\t_ \"github.com/cosmos/cosmos-sdk/x/group/module\" // import for side-effects\n", New: "", FileMatch: "app_config.go"},
		// time and durationpb (only used by group)
		{Old: "\t\"time\"\n", New: "", FileMatch: "app_config.go"},
		{Old: "\t\"google.golang.org/protobuf/types/known/durationpb\"\n", New: "", FileMatch: "app_config.go"},
		// moduleAccPerms
		{Old: "\t\t{Account: nft.ModuleName},\n", New: "", FileMatch: "app_config.go"},
		// blockAccAddrs
		{Old: "\t\tnft.ModuleName,\n", New: "", FileMatch: "app_config.go"},
		// String arrays (5-tab indent)
		{Old: "\t\t\t\t\tgroup.ModuleName,\n", New: "", FileMatch: "app_config.go"},
		{Old: "\t\t\t\t\tnft.ModuleName,\n", New: "", FileMatch: "app_config.go"},
		{Old: "\t\t\t\t\tcircuittypes.ModuleName,\n", New: "", FileMatch: "app_config.go"},
		// ModuleConfig entries
		{Old: "\t\t{\n\t\t\tName: group.ModuleName,\n\t\t\tConfig: appconfig.WrapAny(&groupmodulev1.Module{\n\t\t\t\tMaxExecutionPeriod: durationpb.New(time.Second * 1209600),\n\t\t\t\tMaxMetadataLen:     255,\n\t\t\t}),\n\t\t},\n", New: "", FileMatch: "app_config.go"},
		{Old: "\t\t{\n\t\t\tName:   nft.ModuleName,\n\t\t\tConfig: appconfig.WrapAny(&nftmodulev1.Module{}),\n\t\t},\n", New: "", FileMatch: "app_config.go"},
		{Old: "\t\t{\n\t\t\tName:   circuittypes.ModuleName,\n\t\t\tConfig: appconfig.WrapAny(&circuitmodulev1.Module{}),\n\t\t},\n", New: "", FileMatch: "app_config.go"},
	}

	// Write to a temp file named app_config.go (so FileMatch works)
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "app_config.go")
	if err := os.WriteFile(tmpFile, []byte(postASTContent), 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	modified, err := applyTextReplacements(tmpFile, appConfigReplacements)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !modified {
		t.Error("expected modifications to app_config.go")
	}

	result, _ := os.ReadFile(tmpFile)
	output := string(result)

	// === ASSERTIONS ===

	// Imports removed
	assertMissing(t, output, "circuitmodulev1", "circuitmodulev1 API import should be removed")
	assertMissing(t, output, "groupmodulev1", "groupmodulev1 API import should be removed")
	assertMissing(t, output, "nftmodulev1", "nftmodulev1 API import should be removed")
	assertMissing(t, output, "contrib/x/circuit", "circuit contrib imports should be removed")
	assertMissing(t, output, "contrib/x/nft", "nft contrib imports should be removed")
	assertMissing(t, output, "x/group", "group imports should be removed")
	assertMissing(t, output, "\"time\"", "time import should be removed")
	assertMissing(t, output, "durationpb", "durationpb import should be removed")

	// Surviving imports should still be there
	assertContains(t, output, "cosmossdk.io/core/appconfig", "appconfig import should survive")
	assertContains(t, output, "authtypes", "authtypes import should survive")
	assertContains(t, output, "banktypes", "banktypes import should survive")
	assertContains(t, output, "govtypes", "govtypes import should survive")
	assertContains(t, output, "stakingtypes", "stakingtypes import should survive")

	// moduleAccPerms: nft.ModuleName removed
	assertMissing(t, output, "nft.ModuleName", "nft.ModuleName should be removed everywhere")
	assertContains(t, output, "authtypes.FeeCollectorName", "FeeCollectorName should survive in moduleAccPerms")
	assertContains(t, output, "govtypes.ModuleName", "govtypes.ModuleName should survive")

	// blockAccAddrs: nft.ModuleName removed
	assertContains(t, output, "stakingtypes.BondedPoolName", "BondedPoolName should survive in blockAccAddrs")

	// String arrays: group, nft, circuittypes removed
	assertMissing(t, output, "group.ModuleName", "group.ModuleName should be removed from arrays")
	assertMissing(t, output, "circuittypes.ModuleName", "circuittypes.ModuleName should be removed from arrays")

	// ModuleConfig entries removed
	assertMissing(t, output, "MaxExecutionPeriod", "group ModuleConfig should be removed")
	assertMissing(t, output, "nftmodulev1.Module", "nft ModuleConfig should be removed")
	assertMissing(t, output, "circuitmodulev1.Module", "circuit ModuleConfig should be removed")

	// Surviving ModuleConfig entries
	assertContains(t, output, "govtypes.ModuleName", "gov ModuleConfig should survive")

	t.Logf("=== Transformed app_config.go ===\n%s", output)
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
