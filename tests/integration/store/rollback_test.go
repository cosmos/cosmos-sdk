package rootmulti_test

import (
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

func createApp(t *testing.T, db dbm.DB) runtime.App {
	config := simtestutil.DefaultStartUpConfig()
	config.DB = db
	config.AtGenesis = true

	app, err := simtestutil.SetupWithConfiguration(
		configurator.NewAppConfig(
			configurator.AuthModule(),
			configurator.BankModule(),
			configurator.StakingModule(),
			configurator.TxModule(),
			configurator.ConsensusModule(),
			configurator.ParamsModule(),
		),
		config,
	)
	assert.NilError(t, err)

	return *app
}

func getBankKey(app runtime.App) types.StoreKey {
	var storeKey types.StoreKey
	for _, key := range app.GetStoreKeys() {
		if key.Name() == "bank" {
			storeKey = key
			break
		}
	}
	return storeKey
}

func TestRollback(t *testing.T) {
	db := dbm.NewMemDB()
	app := createApp(t, db)
	ver0 := app.LastBlockHeight()

	storeKey := getBankKey(app)

	// commit 10 blocks
	for i := int64(1); i <= 10; i++ {
		header := tmproto.Header{
			Height:  ver0 + i,
			AppHash: app.LastCommitID().Hash,
		}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})
		ctx := app.NewContext(false, header)
		store := ctx.KVStore(storeKey)
		store.Set([]byte("key"), []byte(fmt.Sprintf("value%d", i)))
		app.Commit()
	}

	assert.Equal(t, ver0+10, app.LastBlockHeight())
	store := app.NewContext(true, tmproto.Header{}).KVStore(storeKey)
	assert.DeepEqual(t, []byte("value10"), store.Get([]byte("key")))

	// rollback 5 blocks
	target := ver0 + 5
	assert.NilError(t, app.CommitMultiStore().RollbackToVersion(target))
	assert.Equal(t, target, app.LastBlockHeight())

	// recreate app to have clean check state
	app = createApp(t, db)
	storeKey = getBankKey(app)

	store = app.NewContext(true, tmproto.Header{}).KVStore(storeKey)
	assert.DeepEqual(t, []byte("value5"), store.Get([]byte("key")))

	// commit another 5 blocks with different values
	for i := int64(6); i <= 10; i++ {
		header := tmproto.Header{
			Height:  ver0 + i,
			AppHash: app.LastCommitID().Hash,
		}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})
		ctx := app.NewContext(false, header)
		store := ctx.KVStore(storeKey)
		store.Set([]byte("key"), []byte(fmt.Sprintf("VALUE%d", i)))
		app.Commit()
	}

	assert.Equal(t, ver0+10, app.LastBlockHeight())
	store = app.NewContext(true, tmproto.Header{}).KVStore(storeKey)
	assert.DeepEqual(t, []byte("VALUE10"), store.Get([]byte("key")))
}
