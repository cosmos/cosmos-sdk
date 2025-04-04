package rootmulti_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/simapp"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"
)

func TestRollback(t *testing.T) {
	db := dbm.NewMemDB()
	options := simapp.SetupOptions{
		Logger:  log.NewNopLogger(),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	}
	app := simapp.NewSimappWithCustomOptions(t, false, options)
	app.Commit()
	ver0 := app.LastBlockHeight()
	// commit 10 blocks
	for i := int64(1); i <= 10; i++ {
		header := tmproto.Header{
			Height:  ver0 + i,
			AppHash: app.LastCommitID().Hash,
		}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})
		ctx := app.NewContext(false, header)
		store := ctx.KVStore(app.GetKey("bank"))
		store.Set([]byte("key"), []byte(fmt.Sprintf("value%d", i)))
		app.Commit()
	}

	require.Equal(t, ver0+10, app.LastBlockHeight())
	store := app.NewContext(true, tmproto.Header{}).KVStore(app.GetKey("bank"))
	require.Equal(t, []byte("value10"), store.Get([]byte("key")))

	// rollback 5 blocks
	target := ver0 + 5
	require.NoError(t, app.CommitMultiStore().RollbackToVersion(target))
	require.Equal(t, target, app.LastBlockHeight())

	// recreate app to have clean check state
	app = simapp.NewSimApp(options.Logger, options.DB, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	store = app.NewContext(true, tmproto.Header{}).KVStore(app.GetKey("bank"))
	require.Equal(t, []byte("value5"), store.Get([]byte("key")))

	// commit another 5 blocks with different values
	for i := int64(6); i <= 10; i++ {
		header := tmproto.Header{
			Height:  ver0 + i,
			AppHash: app.LastCommitID().Hash,
		}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})
		ctx := app.NewContext(false, header)
		store := ctx.KVStore(app.GetKey("bank"))
		store.Set([]byte("key"), []byte(fmt.Sprintf("VALUE%d", i)))
		app.Commit()
	}

	require.Equal(t, ver0+10, app.LastBlockHeight())
	store = app.NewContext(true, tmproto.Header{}).KVStore(app.GetKey("bank"))
	require.Equal(t, []byte("VALUE10"), store.Get([]byte("key")))
}
