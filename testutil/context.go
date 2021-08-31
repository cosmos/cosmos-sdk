package testutil

import (
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	// "github.com/cosmos/cosmos-sdk/store"
	stypes "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/flat"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultContext creates a sdk.Context with a fresh MemDB that can be used in tests.
func DefaultContext(key, tkey stypes.StoreKey) (ret sdk.Context, err error) {
	db := memdb.NewDB()
	opts := flat.DefaultRootStoreConfig()
	err = opts.ReservePrefix(key, stypes.StoreTypePersistent)
	if err != nil {
		return
	}
	err = opts.ReservePrefix(tkey, stypes.StoreTypeTransient)
	if err != nil {
		return
	}
	rs, err := flat.NewRootStore(db, opts)
	if err != nil {
		return
	}
	ret = sdk.NewContext(rs, tmproto.Header{}, false, log.NewNopLogger())
	return
}
