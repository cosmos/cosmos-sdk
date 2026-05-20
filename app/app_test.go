package app

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestAppendIfMissing(t *testing.T) {
	order := []string{"a", "b"}

	unchanged := appendIfMissing(order, "a")
	if len(unchanged) != 2 {
		t.Fatalf("expected unchanged length, got %d", len(unchanged))
	}

	extended := appendIfMissing(order, "c")
	if len(extended) != 3 {
		t.Fatalf("expected extended length, got %d", len(extended))
	}
	if extended[2] != "c" {
		t.Fatalf("expected appended module c, got %q", extended[2])
	}
}

func TestBlockedAddressesExcludesGovModuleAddress(t *testing.T) {
	app := &SDKApp{
		moduleAccountPerms: map[string][]string{
			govtypes.ModuleName:  nil,
			authtypes.ModuleName: nil,
		},
	}

	blocked := app.BlockedAddresses()

	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	if blocked[govAddr] {
		t.Fatalf("expected gov module address %s to be allowed", govAddr)
	}

	authAddr := authtypes.NewModuleAddress(authtypes.ModuleName).String()
	if !blocked[authAddr] {
		t.Fatalf("expected auth module address %s to remain blocked", authAddr)
	}
}

func TestConfigureExecutionModeUsesSerialWhenBlockSTMIsNil(t *testing.T) {
	app := &SDKApp{
		cfg: SDKAppConfig{},
	}

	app.configureExecutionMode()
}

func TestNewSDKAppWithOptimisticExecutionEnabledLoadsModules(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	cfg.OptimisticExecutionEnabled = true

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected startup with optimistic execution enabled to succeed, got panic: %v", r)
		}
	}()

	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	if app == nil {
		t.Fatal("expected NewSDKApp to return a non-nil app")
	}

	app.LoadModules()
}

func TestNewSDKAppRegistersConfigStoreKeys(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	cfg.Keys = []string{"custom_test_key"}

	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	if app.GetKey("custom_test_key") == nil {
		t.Fatal("expected custom key from SDKAppConfig.Keys to be registered")
	}
}

func TestNewSDKAppRegistersTransientStoreKeys(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	cfg.TransientStoreKeys = []string{"custom_transient_key"}

	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	if app.GetTransientStoreKey("custom_transient_key") == nil {
		t.Fatal("expected custom transient key from SDKAppConfig.TransientStoreKeys to be registered")
	}
}
