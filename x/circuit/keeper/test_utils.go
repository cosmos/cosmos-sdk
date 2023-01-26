package keeper

import (
	"context"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"
)

type baseSuite struct {
	t   *testing.T
	err error
	ctx context.Context

	// k        Keeper //TODO uncomment this after implementing
	addrs    []sdk.AccAddress
	storeKey *storetypes.KVStoreKey
	sdkCtx   sdk.Context
}

func setupBase(t *testing.T) *baseSuite {
	s := &baseSuite{t: t}
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), nil)
	s.storeKey = storetypes.NewKVStoreKey("test")
	cms.MountStoreWithDB(s.storeKey, storetypes.StoreTypeIAVL, db)
	assert.NilError(t, cms.LoadLatestVersion())

	s.sdkCtx = sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())
	s.ctx = sdk.WrapSDKContext(s.sdkCtx)

	// s.k = NewKeeper()

	return s
}
