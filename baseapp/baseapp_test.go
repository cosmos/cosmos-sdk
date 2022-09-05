package baseapp

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	capKey1 = sdk.NewKVStoreKey("key1")
	capKey2 = sdk.NewKVStoreKey("key2")
)

func TestSetMinGasPrices(t *testing.T) {
	minGasPrices := sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5000)}
	app := setupBaseApp(t, SetMinGasPrices(minGasPrices.String()))
	require.Equal(t, minGasPrices, app.minGasPrices)
}

func TestGetMaximumBlockGas(t *testing.T) {
	app := setupBaseApp(t)
	app.InitChain(abci.RequestInitChain{})
	ctx := app.NewContext(true, tmproto.Header{})

	app.StoreConsensusParams(ctx, &tmproto.ConsensusParams{Block: &tmproto.BlockParams{MaxGas: 0}})
	require.Equal(t, uint64(0), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &tmproto.ConsensusParams{Block: &tmproto.BlockParams{MaxGas: -1}})
	require.Equal(t, uint64(0), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &tmproto.ConsensusParams{Block: &tmproto.BlockParams{MaxGas: 5000000}})
	require.Equal(t, uint64(5000000), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &tmproto.ConsensusParams{Block: &tmproto.BlockParams{MaxGas: -5000000}})
	require.Panics(t, func() { app.getMaximumBlockGas(ctx) })
}

func TestLoadVersionPruning(t *testing.T) {
	logger := log.NewNopLogger()
	pruningOptions := pruningtypes.NewCustomPruningOptions(10, 15)
	pruningOpt := SetPruning(pruningOptions)
	db := dbm.NewMemDB()
	name := t.Name()
	app := NewBaseApp(name, logger, db, nil, pruningOpt)

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("key1")
	app.MountStores(capKey)

	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := storetypes.CommitID{}

	// fresh store has zero/empty last commit
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	var lastCommitID storetypes.CommitID

	// Commit seven blocks, of which 7 (latest) is kept in addition to 6, 5
	// (keep recent) and 3 (keep every).
	for i := int64(1); i <= 7; i++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: i}})
		res := app.Commit()
		lastCommitID = storetypes.CommitID{Version: i, Hash: res.Data}
	}

	for _, v := range []int64{1, 2, 4} {
		_, err = app.cms.CacheMultiStoreWithVersion(v)
		require.NoError(t, err)
	}

	for _, v := range []int64{3, 5, 6, 7} {
		_, err = app.cms.CacheMultiStoreWithVersion(v)
		require.NoError(t, err)
	}

	// reload with LoadLatestVersion, check it loads last version
	app = NewBaseApp(name, logger, db, nil, pruningOpt)
	app.MountStores(capKey)

	err = app.LoadLatestVersion()
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(7), lastCommitID)
}

// simple one store baseapp
func setupBaseApp(t *testing.T, options ...func(*BaseApp)) *BaseApp {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	app := NewBaseApp(t.Name(), logger, db, nil, options...)
	require.Equal(t, t.Name(), app.Name())

	app.MountStores(capKey1, capKey2)
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})

	// stores are mounted
	err := app.LoadLatestVersion()
	require.Nil(t, err)
	return app
}

func testLoadVersionHelper(t *testing.T, app *BaseApp, expectedHeight int64, expectedID storetypes.CommitID) {
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}
