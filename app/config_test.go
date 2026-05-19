package app

import (
	"slices"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

func TestDefaultSDKAppConfigReturnsIndependentSlices(t *testing.T) {
	cfgA := DefaultSDKAppConfig("a", testAppOptions(t))
	cfgB := DefaultSDKAppConfig("b", testAppOptions(t))

	if len(cfgA.OrderBeginBlockers) == 0 || len(cfgB.OrderBeginBlockers) == 0 {
		t.Fatal("expected non-empty begin blocker defaults")
	}

	cfgA.OrderBeginBlockers[0] = "changed"
	if cfgB.OrderBeginBlockers[0] == "changed" {
		t.Fatal("expected begin blocker slices to be independent")
	}
}

func TestDefaultSDKAppConfigReturnsIndependentMapsAndNestedSlices(t *testing.T) {
	cfgA := DefaultSDKAppConfig("a", testAppOptions(t))
	cfgB := DefaultSDKAppConfig("b", testAppOptions(t))

	cfgA.ModuleAccountPerms["new"] = []string{"burner"}
	if _, ok := cfgB.ModuleAccountPerms["new"]; ok {
		t.Fatal("expected module account maps to be independent")
	}

	for moduleName, perms := range cfgA.ModuleAccountPerms {
		if len(perms) == 0 {
			continue
		}

		old := perms[0]
		cfgA.ModuleAccountPerms[moduleName][0] = "changed"
		if len(cfgB.ModuleAccountPerms[moduleName]) > 0 && cfgB.ModuleAccountPerms[moduleName][0] == "changed" {
			t.Fatalf("expected nested permission slices to be independent for module %s", moduleName)
		}
		cfgA.ModuleAccountPerms[moduleName][0] = old
		break
	}
}

func TestProcessOptionalModulesDoesNotMutateOtherConfigs(t *testing.T) {
	cfgA := DefaultSDKAppConfig("a", testAppOptions(t))
	cfgB := DefaultSDKAppConfig("b", testAppOptions(t))

	cfgA.WithMint = false
	cfgA.processOptionalModules()

	if _, ok := cfgA.ModuleAccountPerms["mint"]; ok {
		t.Fatal("expected mint permissions to be removed from cfgA")
	}

	if _, ok := cfgB.ModuleAccountPerms["mint"]; !ok {
		t.Fatal("expected cfgB mint permissions to remain unchanged")
	}

	if !slices.Contains(cfgB.OrderBeginBlockers, "mint") {
		t.Fatal("expected cfgB order begin blockers to remain unchanged")
	}
}

func TestProcessOptionalModulesRemovesMintFromOrdering(t *testing.T) {
	cfg := DefaultSDKAppConfig("a", testAppOptions(t))

	cfg.WithMint = false
	cfg.processOptionalModules()

	if slices.Contains(cfg.OrderBeginBlockers, "mint") {
		t.Fatal("expected mint to be removed from begin blocker ordering when disabled")
	}
	if slices.Contains(cfg.OrderInitGenesis, "mint") {
		t.Fatal("expected mint to be removed from init genesis ordering when disabled")
	}
	if slices.Contains(cfg.OrderExportGenesis, "mint") {
		t.Fatal("expected mint to be removed from export genesis ordering when disabled")
	}
}

func TestDefaultSDKAppConfigRequiresAppOptions(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic with nil app opts")
		}
	}()
	_ = DefaultSDKAppConfig("app", nil)
}

func TestDefaultSDKAppConfigInjectsChainIDFallback(t *testing.T) {
	opts := appOptionsMap{
		flags.FlagHome: t.TempDir(),
	}

	cfg := DefaultSDKAppConfig("my-app", opts)
	if got := cfg.AppOpts.Get(flags.FlagChainID); got != "my-app" {
		t.Fatalf("expected chain-id fallback to app name, got %v", got)
	}
}

func TestSDKAppConfigValidateBlockSTMWorkers(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	cfg.ExecutionMode = ExecutionModeBlockSTM
	cfg.BlockSTM.Workers = 0

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for invalid blockstm workers")
	}
}

func testAppOptions(t *testing.T) appOptionsMap {
	t.Helper()
	return appOptionsMap{
		flags.FlagHome: t.TempDir(),
	}
}

type appOptionsMap map[string]any

func (m appOptionsMap) Get(key string) any {
	return m[key]
}
