package keeper_test

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func newTestAccountKeeper(t *testing.T, perms map[string][]string) keeper.AccountKeeper {
	t.Helper()
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	return keeper.NewAccountKeeper(
		encCfg.Codec,
		sdk.NewKVStoreService(key),
		types.ProtoBaseAccount,
		perms,
		authcodec.NewBech32Codec("cosmos"),
		"cosmos",
		types.NewModuleAddress("gov").String(),
	)
}

func TestLoadMaccPermsReplacesFullSet(t *testing.T) {
	ak := newTestAccountKeeper(t, map[string][]string{"a": nil})
	if _, ok := ak.GetModulePermissions()["a"]; !ok {
		t.Fatal("expected 'a' to be present after construction")
	}

	ak.LoadMaccPerms(map[string][]string{"b": {types.Minter}})

	perms := ak.GetModulePermissions()
	if _, ok := perms["a"]; ok {
		t.Fatal("expected 'a' to be removed after LoadMaccPerms")
	}
	if _, ok := perms["b"]; !ok {
		t.Fatal("expected 'b' to be present after LoadMaccPerms")
	}
}

func TestLoadMaccPermsCalledTwiceReplaces(t *testing.T) {
	ak := newTestAccountKeeper(t, nil)

	ak.LoadMaccPerms(map[string][]string{"first": nil})
	ak.LoadMaccPerms(map[string][]string{"second": nil})

	perms := ak.GetModulePermissions()
	if _, ok := perms["first"]; ok {
		t.Fatal("expected 'first' to be replaced by second LoadMaccPerms call")
	}
	if _, ok := perms["second"]; !ok {
		t.Fatal("expected 'second' to be present after second LoadMaccPerms call")
	}
}

func TestLoadMaccPermsVisibleToExistingCopies(t *testing.T) {
	ak := newTestAccountKeeper(t, map[string][]string{"original": nil})

	// take a value copy before the reload
	akCopy := ak

	ak.LoadMaccPerms(map[string][]string{"updated": nil})

	// both the original and the copy share the same underlying map
	if _, ok := akCopy.GetModulePermissions()["updated"]; !ok {
		t.Fatal("expected LoadMaccPerms to be visible to existing keeper copies (shared map)")
	}
	if _, ok := akCopy.GetModulePermissions()["original"]; ok {
		t.Fatal("expected old entry to be gone in keeper copy after LoadMaccPerms")
	}
}
