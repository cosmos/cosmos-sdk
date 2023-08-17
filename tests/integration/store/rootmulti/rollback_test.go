package rootmulti_test

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"gotest.tools/v3/assert"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func TestRollback(t *testing.T) {
	db := dbm.NewMemDB()
	options := simapp.SetupOptions{
		Logger:  log.NewNopLogger(),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	}
	app := simapp.NewSimappWithCustomOptions(t, false, options)
	ver0 := app.LastBlockHeight()
	// commit 10 blocks
	for i := int64(1); i <= 10; i++ {
		header := cmtproto.Header{
			Height:  ver0 + i,
			AppHash: app.LastCommitID().Hash,
		}

		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: header.Height,
		})
		assert.NilError(t, err)
		ctx := app.NewContextLegacy(false, header)
		store := ctx.KVStore(app.GetKey("bank"))
		store.Set([]byte("key"), []byte(fmt.Sprintf("value%d", i)))
		_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: header.Height,
		})
		assert.NilError(t, err)
		_, err = app.Commit()
		assert.NilError(t, err)
	}

	assert.Equal(t, ver0+10, app.LastBlockHeight())
	store := app.NewContext(true).KVStore(app.GetKey("bank"))
	assert.DeepEqual(t, []byte("value10"), store.Get([]byte("key")))

	// rollback 5 blocks
	target := ver0 + 5
	assert.NilError(t, app.CommitMultiStore().RollbackToVersion(target))
	assert.Equal(t, target, app.LastBlockHeight())

	// recreate app to have clean check state
	app = simapp.NewSimApp(options.Logger, options.DB, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	store = app.NewContext(true).KVStore(app.GetKey("bank"))
	assert.DeepEqual(t, []byte("value5"), store.Get([]byte("key")))

	// commit another 5 blocks with different values
	for i := int64(6); i <= 10; i++ {
		header := cmtproto.Header{
			Height:  ver0 + i,
			AppHash: app.LastCommitID().Hash,
		}
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: header.Height})
		assert.NilError(t, err)
		ctx := app.NewContextLegacy(false, header)
		store := ctx.KVStore(app.GetKey("bank"))
		store.Set([]byte("key"), []byte(fmt.Sprintf("VALUE%d", i)))
		_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: header.Height,
		})
		assert.NilError(t, err)
		_, err = app.Commit()
		assert.NilError(t, err)
	}

	assert.Equal(t, ver0+10, app.LastBlockHeight())
	store = app.NewContext(true).KVStore(app.GetKey("bank"))
	assert.DeepEqual(t, []byte("VALUE10"), store.Get([]byte("key")))
}
