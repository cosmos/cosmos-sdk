package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keys for parameter access
const (
	TestParamStore = "ParamsTest"
)

// Returns components for testing
func DefaultTestComponents(t *testing.T, table Table) (sdk.Context, Store, func() sdk.CommitID) {
	cdc := codec.New()
	key := sdk.NewKVStoreKey("params")
	tkey := sdk.NewTransientStoreKey("tparams")
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.WithTracer(os.Stdout)
	ms.WithTracingContext(sdk.TraceContext{})
	ms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkey, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewTMLogger(os.Stdout))
	store := NewStore(cdc, key, tkey, TestParamStore, table)

	return ctx, store, ms.Commit
}
