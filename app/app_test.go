package app

import (
	"testing"

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

func TestConfigureExecutionModePanicsOnUnsupportedMode(t *testing.T) {
	app := &SDKApp{
		cfg: SDKAppConfig{
			ExecutionMode: "invalid-mode",
		},
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unsupported execution mode")
		}
	}()

	app.configureExecutionMode()
}
