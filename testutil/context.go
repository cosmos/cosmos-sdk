package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
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
	if err = opts.RegisterSubstore(key, stypes.StoreTypePersistent); err != nil {
		return
	}
	if err = opts.RegisterSubstore(tkey, stypes.StoreTypeTransient); err != nil {
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
	DB  *memdb.MemDB
	CMS store.CommitMultiStore
}

func DefaultContextWithDB(t *testing.T, key store.Key, tkey storetypes.StoreKey) TestContext {
	db := memdb.NewDB()
	opts := multi.DefaultStoreParams()
	assert.NoError(t, opts.RegisterSubstore(key, stypes.StoreTypePersistent))
	assert.NoError(t, opts.RegisterSubstore(tkey, stypes.StoreTypeTransient))
	cms, err := multi.NewV1MultiStoreAsV2(db, opts)
	assert.NoError(t, err)

	ctx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())

	return TestContext{ctx, db, cms}
}
