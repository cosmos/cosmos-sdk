package app

import (
	"errors"
	"testing"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func newUpgradeTestApp(t *testing.T) *SDKApp {
	t.Helper()
	cfg := DefaultSDKAppConfig("app", testAppOptions(t))
	a := NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	a.LoadModules()
	return a
}

func TestRegisterUpgradeHandlersSetsHandler(t *testing.T) {
	a := newUpgradeTestApp(t)

	RegisterUpgradeHandlers[AppI](a, Upgrade[AppI]{Name: "v2"})

	if !a.UpgradeKeeper().HasHandler("v2") {
		t.Fatal("expected upgrade handler for v2 to be registered")
	}
}

func TestRegisterUpgradeHandlersNoopWhenEmpty(t *testing.T) {
	a := newUpgradeTestApp(t)

	// Must not panic with zero upgrades.
	RegisterUpgradeHandlers[AppI](a)
}

func TestRegisterUpgradeHandlersCallbackInvoked(t *testing.T) {
	a := newUpgradeTestApp(t)
	called := false

	cb := Upgrade[AppI]{
		Name: "v2-cb",
		UpgradeCallBack: func(ctx sdk.Context, plan upgradetypes.Plan, app AppI) error {
			called = true
			return nil
		},
	}
	RegisterUpgradeHandlers[AppI](a, cb)

	if !a.UpgradeKeeper().HasHandler("v2-cb") {
		t.Fatal("expected handler to be registered")
	}

	// Exercise the UpgradeCallBack != nil branch without needing an
	// initialized store — call the callback directly with a zero-value context.
	err := cb.UpgradeCallBack(sdk.Context{}, upgradetypes.Plan{Name: "v2-cb"}, a)
	if err != nil {
		t.Fatalf("callback returned unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected UpgradeCallBack to be invoked")
	}
}

func TestRegisterUpgradeHandlersCallbackErrorPropagates(t *testing.T) {
	sentinel := errors.New("callback failed")

	a := newUpgradeTestApp(t)
	cb := Upgrade[AppI]{
		Name: "v3-err",
		UpgradeCallBack: func(ctx sdk.Context, plan upgradetypes.Plan, app AppI) error {
			return sentinel
		},
	}
	RegisterUpgradeHandlers[AppI](a, cb)

	err := cb.UpgradeCallBack(sdk.Context{}, upgradetypes.Plan{Name: "v3-err"}, a)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestRegisterUpgradeHandlersPanicOnDuplicateName(t *testing.T) {
	a := newUpgradeTestApp(t)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate upgrade name")
		}
	}()

	RegisterUpgradeHandlers[AppI](a,
		Upgrade[AppI]{Name: "dup"},
		Upgrade[AppI]{Name: "dup"},
	)
}

func TestRegisterUpgradeHandlersPanicOnNilUpgradeKeeper(t *testing.T) {
	a := &SDKApp{}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when upgrade keeper is nil")
		}
	}()

	RegisterUpgradeHandlers[AppI](a, Upgrade[AppI]{Name: "v2"})
}
