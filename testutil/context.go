package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	stypes "github.com/cosmos/cosmos-sdk/store/v2alpha1"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/multi"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultContext creates a sdk.Context with a fresh MemDB that can be used in tests.
func DefaultContext(key, tkey stypes.StoreKey) (ret sdk.Context) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	db := memdb.NewDB()
	opts := multi.DefaultStoreParams()
	err = opts.RegisterSubstore(key, stypes.StoreTypePersistent)
	if err != nil {
		return
	}
	err = opts.RegisterSubstore(tkey, stypes.StoreTypeTransient)
	if err != nil {
		return
	}
	rs, err := multi.NewV1MultiStoreAsV2(db, opts)
	if err != nil {
		return
	}
	ret = sdk.NewContext(rs.CacheWrap(), tmproto.Header{}, false, log.NewNopLogger())
	return
}

type TestContext struct {
	Ctx sdk.Context
	DB  *dbm.MemDB
	CMS store.CommitMultiStore
}

func DefaultContextWithDB(t *testing.T, key storetypes.StoreKey, tkey storetypes.StoreKey) TestContext {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	err := cms.LoadLatestVersion()
	assert.NoError(t, err)

	ctx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())

	return TestContext{ctx, db, cms}
}
