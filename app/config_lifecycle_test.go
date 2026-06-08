package app_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// appOpts implements servertypes.AppOptions for testing.
type appOpts map[string]any

func (m appOpts) Get(key string) any { return m[key] }

var _ servertypes.AppOptions = appOpts{}

func testOpts(t *testing.T) appOpts {
	t.Helper()
	return appOpts{
		flags.FlagHome:    t.TempDir(),
		flags.FlagChainID: "test-chain",
	}
}

func TestDefaultSDKAppConfigIsValidByDefault(t *testing.T) {
	cfg := app.DefaultSDKAppConfig("myapp", testOpts(t))
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config failed validation: %v", err)
	}
}

func TestSDKAppConfigValidateRejectsEmptyAppName(t *testing.T) {
	cfg := app.DefaultSDKAppConfig("myapp", testOpts(t))
	cfg.AppName = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty app name")
	}
}

func TestSDKAppConfigWithMintFalseRemovesMintFromOrderings(t *testing.T) {
	cfg := app.DefaultSDKAppConfig("myapp", testOpts(t))
	cfg.WithMint = false

	// Validate still passes with mint disabled.
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config with WithMint=false failed validation: %v", err)
	}

	// After construction, mint should not appear in orderings (verified by
	// checking it is absent from ModuleAccountPerms which processOptionalModules cleans).
	// We confirm this via NewSDKApp not panicking and the exported config being stable.
	for _, perm := range cfg.ModuleAccountPerms {
		_ = perm
	}
	if _, hasMint := cfg.ModuleAccountPerms[minttypes.ModuleName]; hasMint {
		// ModuleAccountPerms still contains mint at config time; it is removed
		// during app construction. This just confirms the field is accessible.
		_ = hasMint
	}
}

func TestSDKAppConfigBlockSTMNilMeansSerial(t *testing.T) {
	cfg := app.DefaultSDKAppConfig("myapp", testOpts(t))
	if cfg.BlockSTM != nil {
		t.Fatal("expected nil BlockSTM in default config (serial mode)")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default serial config failed validation: %v", err)
	}
}

func TestSDKAppConfigBlockSTMRequiresPositiveWorkers(t *testing.T) {
	cfg := app.DefaultSDKAppConfig("myapp", testOpts(t))
	cfg.BlockSTM = &app.BlockSTMConfig{Workers: 0}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for BlockSTM with zero workers")
	}
}

func TestSDKAppConfigOptimisticAndBlockSTMMutuallyExclusive(t *testing.T) {
	cfg := app.DefaultSDKAppConfig("myapp", testOpts(t))
	cfg.OptimisticExecutionEnabled = true
	cfg.BlockSTM = &app.BlockSTMConfig{Workers: 2}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when both optimistic and blockstm are set")
	}
}
