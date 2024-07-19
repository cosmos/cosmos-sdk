package testutil

import (
	"testing"
	"time"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/assert"

	"cosmossdk.io/core/header"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultContext creates a sdk.Context with a fresh MemDB that can be used in tests.
func DefaultContext(key, tkey storetypes.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, coretesting.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	ctx := sdk.NewContext(cms, false, coretesting.NewNopLogger())

	return ctx
}

// DefaultContextWithKeys creates a sdk.Context with a fresh MemDB, mounting the providing keys for usage in the multistore.
// This function is intended to be used for testing purposes only.
func DefaultContextWithKeys(
	keys map[string]*storetypes.KVStoreKey,
	transKeys map[string]*storetypes.TransientStoreKey,
	memKeys map[string]*storetypes.MemoryStoreKey,
) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, coretesting.NewNopLogger(), metrics.NewNoOpMetrics())

	for _, key := range keys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	}

	for _, tKey := range transKeys {
		cms.MountStoreWithDB(tKey, storetypes.StoreTypeTransient, db)
	}

	for _, memkey := range memKeys {
		cms.MountStoreWithDB(memkey, storetypes.StoreTypeMemory, db)
	}

	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	return sdk.NewContext(cms, false, coretesting.NewNopLogger())
}

type TestContext struct {
	Ctx sdk.Context
	DB  *dbm.MemDB
	CMS store.CommitMultiStore
}

func DefaultContextWithDB(tb testing.TB, key, tkey storetypes.StoreKey) TestContext {
	tb.Helper()
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, coretesting.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	err := cms.LoadLatestVersion()
	assert.NoError(tb, err)

	ctx := sdk.NewContext(cms, false, coretesting.NewNopLogger()).WithHeaderInfo(header.Info{Time: time.Now()})

	return TestContext{ctx, db, cms}
}
