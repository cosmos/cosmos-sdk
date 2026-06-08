package app

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
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
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("configureExecutionMode panicked with nil BlockSTM: %v", r)
		}
	}()

	a := &SDKApp{cfg: SDKAppConfig{}}
	a.configureExecutionMode()
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

func TestAddModulesRegistersCustomStoreKey(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)

	if err := app.AddModules(newTestCustomModule("custom", "custom_test_key")); err != nil {
		t.Fatalf("expected AddModules to succeed, got: %v", err)
	}

	if app.GetKey("custom_test_key") == nil {
		t.Fatal("expected custom store key to be registered via AddModules")
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

type testCustomModule struct {
	module.AppModule

	name          string
	perms         map[string][]string
	keys          map[string]*storetypes.KVStoreKey
	transientKeys map[string]*storetypes.TransientStoreKey
}

func (m testCustomModule) Name() string {
	return m.name
}

func (m testCustomModule) ModuleAccountPermissions() map[string][]string {
	return m.perms
}

func (m testCustomModule) StoreKeys() map[string]*storetypes.KVStoreKey {
	return m.keys
}

func (m testCustomModule) TransientStoreKeys() map[string]*storetypes.TransientStoreKey {
	return m.transientKeys
}

type testGenesisModule struct {
	name string
}

func (m testGenesisModule) Name() string { return m.name }

func (testGenesisModule) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

func (testGenesisModule) RegisterInterfaces(codectypes.InterfaceRegistry) {}

func (testGenesisModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

func (testGenesisModule) DefaultGenesis(codec.JSONCodec) json.RawMessage { return nil }

func (testGenesisModule) ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error {
	return nil
}

func (testGenesisModule) InitGenesis(sdk.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

func (testGenesisModule) ExportGenesis(sdk.Context, codec.JSONCodec) json.RawMessage { return nil }

func newTestCustomModule(name, storeKey string) testCustomModule {
	return testCustomModule{
		AppModule: module.NewGenesisOnlyAppModule(testGenesisModule{name: name}),
		name:      name,
		perms:     map[string][]string{},
		keys: map[string]*storetypes.KVStoreKey{
			storeKey: storetypes.NewKVStoreKey(storeKey),
		},
	}
}

func TestAddModulesFailsAfterLoadModules(t *testing.T) {
	app := &SDKApp{
		// Use loaded=true to simulate the post-LoadModules state; addModule now
		// guards on app.loaded rather than app.moduleManager.
		loaded: true,
	}

	err := app.AddModules(testCustomModule{name: "custom"})
	if err == nil {
		t.Fatal("expected AddModules to fail after LoadModules")
	}
}

func TestAddModulesAfterLoadModulesDoesNotMutateAppState(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	app.LoadModules()

	originalStoreKeyCount := len(app.storeKeys)
	originalCustomModuleCount := len(app.customModules)
	originalBeginBlockerCount := len(app.orderBeginBlockers)

	err := app.AddModules(testCustomModule{
		name: "custom",
		keys: map[string]*storetypes.KVStoreKey{
			"custom": storetypes.NewKVStoreKey("custom"),
		},
	})
	if err == nil {
		t.Fatal("expected AddModules to fail after LoadModules")
	}
	if len(app.storeKeys) != originalStoreKeyCount {
		t.Fatal("expected store keys to remain unchanged after failed AddModules")
	}
	if len(app.customModules) != originalCustomModuleCount {
		t.Fatal("expected custom modules to remain unchanged after failed AddModules")
	}
	if len(app.orderBeginBlockers) != originalBeginBlockerCount {
		t.Fatal("expected ordering slices to remain unchanged after failed AddModules")
	}
}

func TestAddModulesAcceptsAndMergesModuleAccountPermissions(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)

	mod := newTestCustomModule("custmod", "custmod_store")
	mod.perms = map[string][]string{"custmod": {authtypes.Minter}}

	if err := app.AddModules(mod); err != nil {
		t.Fatalf("expected AddModules to accept module account permissions, got: %v", err)
	}
	if _, ok := app.moduleAccountPerms["custmod"]; !ok {
		t.Fatal("expected custmod perm to be merged into moduleAccountPerms")
	}
	custAddr := authtypes.NewModuleAddress("custmod").String()
	if !app.BankKeeper.GetBlockedAddresses()[custAddr] {
		t.Fatal("expected custmod address to be blocked in BankKeeper immediately after AddModules")
	}
}

func TestAddModulesRejectsDuplicateModuleAccountPermissions(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)

	modA := newTestCustomModule("modA", "modA_store")
	modA.perms = map[string][]string{"shared": nil}
	if err := app.AddModules(modA); err != nil {
		t.Fatalf("first AddModules should succeed, got: %v", err)
	}

	originalPermCount := len(app.moduleAccountPerms)
	originalBlockedCount := len(app.BankKeeper.GetBlockedAddresses())

	modB := newTestCustomModule("modB", "modB_store")
	modB.perms = map[string][]string{"shared": nil}
	if err := app.AddModules(modB); err == nil {
		t.Fatal("expected AddModules to reject duplicate module account permission")
	}
	if len(app.moduleAccountPerms) != originalPermCount {
		t.Fatal("expected moduleAccountPerms to be unchanged after duplicate perm rejection")
	}
	if len(app.BankKeeper.GetBlockedAddresses()) != originalBlockedCount {
		t.Fatal("expected blocked addresses to be unchanged after duplicate perm rejection")
	}
}

func TestLoadModulesAppliesCustomPermsToAccountKeeper(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)

	mod := newTestCustomModule("custmod", "custmod_store")
	mod.perms = map[string][]string{"custmod": nil}
	if err := app.AddModules(mod); err != nil {
		t.Fatalf("AddModules failed: %v", err)
	}

	app.LoadModules()

	if _, ok := app.AccountKeeper.GetModulePermissions()["custmod"]; !ok {
		t.Fatal("expected custmod perm to appear in AccountKeeper after LoadModules")
	}
}

func TestAddModulesRegistersTransientStoreKeysBeforeLoadModules(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	customModule := newTestCustomModule("custom", "custom_store")
	customModule.transientKeys = map[string]*storetypes.TransientStoreKey{
		"custom_tstore": storetypes.NewTransientStoreKey("custom_tstore"),
	}

	err := app.AddModules(customModule)
	if err != nil {
		t.Fatalf("expected AddModules to succeed, got: %v", err)
	}

	if app.GetTransientStoreKey("custom_tstore") == nil {
		t.Fatal("expected transient store key to be registered before LoadModules")
	}

	app.LoadModules()

	if app.GetTransientStoreKey("custom_tstore") == nil {
		t.Fatal("expected transient store key to remain registered after LoadModules")
	}
}

func TestAddModulesBeforeLoadModulesRegistersInManagerAndOrdering(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	customModule := newTestCustomModule("custom", "custom_store")

	if err := app.AddModules(customModule); err != nil {
		t.Fatalf("expected AddModules before LoadModules to succeed, got: %v", err)
	}
	if app.GetKey("custom_store") == nil {
		t.Fatal("expected custom store key to be registered before LoadModules")
	}

	app.LoadModules()

	if _, found := app.ModuleManager().Modules["custom"]; !found {
		t.Fatal("expected custom module to be present in module manager after LoadModules")
	}
	if !containsModule(app.ModuleManager().OrderBeginBlockers, "custom") {
		t.Fatal("expected custom module in begin blocker ordering")
	}
	if !containsModule(app.ModuleManager().OrderInitGenesis, "custom") {
		t.Fatal("expected custom module in init genesis ordering")
	}
}

func TestAddModulesRejectsDuplicateKVStoreKeysWithoutMutation(t *testing.T) {
	app := &SDKApp{
		storeKeys: map[string]*storetypes.KVStoreKey{
			"dup_key": storetypes.NewKVStoreKey("dup_key"),
		},
		transientStoreKeys: map[string]*storetypes.TransientStoreKey{},
	}

	originalStoreKeyCount := len(app.storeKeys)
	originalCustomModuleCount := len(app.customModules)

	err := app.AddModules(testCustomModule{
		name:  "custom",
		perms: map[string][]string{},
		keys: map[string]*storetypes.KVStoreKey{
			"dup_key": storetypes.NewKVStoreKey("dup_key"),
		},
	})
	if err == nil {
		t.Fatal("expected AddModules to reject duplicate KV store key")
	}
	if len(app.storeKeys) != originalStoreKeyCount {
		t.Fatal("expected store keys to remain unchanged after duplicate key rejection")
	}
	if len(app.customModules) != originalCustomModuleCount {
		t.Fatal("expected custom modules to remain unchanged after duplicate key rejection")
	}
}

func TestAddModulesRejectsDuplicateTransientStoreKeysWithoutMutation(t *testing.T) {
	app := &SDKApp{
		storeKeys: map[string]*storetypes.KVStoreKey{
			"custom_store": storetypes.NewKVStoreKey("custom_store"),
		},
		transientStoreKeys: map[string]*storetypes.TransientStoreKey{
			"dup_tkey": storetypes.NewTransientStoreKey("dup_tkey"),
		},
	}

	originalStoreKeyCount := len(app.storeKeys)
	originalTransientCount := len(app.transientStoreKeys)
	originalCustomModuleCount := len(app.customModules)

	// the module introduces a new KV key (custom_store_2) and a duplicate transient
	// key. the duplicate check must prevent custom_store_2 from leaking into
	// app.storeKeys even though the KV key itself is not the collision.
	err := app.AddModules(testCustomModule{
		name:  "custom",
		perms: map[string][]string{},
		keys: map[string]*storetypes.KVStoreKey{
			"custom_store_2": storetypes.NewKVStoreKey("custom_store_2"),
		},
		transientKeys: map[string]*storetypes.TransientStoreKey{
			"dup_tkey": storetypes.NewTransientStoreKey("dup_tkey"),
		},
	})
	if err == nil {
		t.Fatal("expected AddModules to reject duplicate transient store key")
	}
	if len(app.storeKeys) != originalStoreKeyCount {
		t.Fatal("expected KV store keys to remain unchanged after transient key collision")
	}
	if len(app.transientStoreKeys) != originalTransientCount {
		t.Fatal("expected transient store keys to remain unchanged after duplicate key rejection")
	}
	if len(app.customModules) != originalCustomModuleCount {
		t.Fatal("expected custom modules to remain unchanged after duplicate key rejection")
	}
}

func TestSetAnteHandlerKeepsFeegrantInterfaceNilWhenDisabled(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	cfg.WithFeeGrant = false
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	handlerOpts := app.buildAnteHandlerOptions(app.TxConfig())
	if handlerOpts.FeegrantKeeper != nil {
		t.Fatal("expected feegrant keeper interface to remain nil when disabled")
	}
}

func TestSetAnteHandlerSetsFeegrantInterfaceWhenEnabled(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	cfg.WithFeeGrant = true
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	handlerOpts := app.buildAnteHandlerOptions(app.TxConfig())
	if handlerOpts.FeegrantKeeper == nil {
		t.Fatal("expected feegrant keeper interface to be set when enabled")
	}
}

func TestLoadModulesRegistersConfiguredUpgrades(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	cfg.Upgrades = []Upgrade[AppI]{
		{Name: "test-upgrade"},
	}

	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	app.LoadModules()

	if !app.UpgradeKeeper().HasHandler("test-upgrade") {
		t.Fatal("expected configured upgrade handler to be registered")
	}
}

func TestLoadModulesIsIdempotent(t *testing.T) {
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	app := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)

	app.LoadModules()
	firstManager := app.ModuleManager()
	firstBeginOrderCount := len(firstManager.OrderBeginBlockers)

	app.LoadModules()
	secondManager := app.ModuleManager()

	if firstManager != secondManager {
		t.Fatal("expected LoadModules to keep the same module manager instance on subsequent calls")
	}
	if len(secondManager.OrderBeginBlockers) != firstBeginOrderCount {
		t.Fatal("expected LoadModules to avoid mutating module ordering on subsequent calls")
	}
}

func containsModule(modules []string, name string) bool {
	for _, moduleName := range modules {
		if moduleName == name {
			return true
		}
	}
	return false
}
