package baseapp

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

var (
	capKey1 = sdk.NewKVStoreKey("key1")
	capKey2 = sdk.NewKVStoreKey("key2")
)

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func newBaseApp(name string, options ...func(*BaseApp)) *BaseApp {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	codec := codec.NewLegacyAmino()
	registerTestCodec(codec)
	return NewBaseApp(name, logger, db, testTxDecoder(codec), options...)
}

func registerTestCodec(cdc *codec.LegacyAmino) {
	// register Tx, Msg
	sdk.RegisterLegacyAminoCodec(cdc)

	// register test types
	cdc.RegisterConcrete(&txTest{}, "cosmos-sdk/baseapp/txTest", nil)
	cdc.RegisterConcrete(&msgCounter{}, "cosmos-sdk/baseapp/msgCounter", nil)
	cdc.RegisterConcrete(&msgCounter2{}, "cosmos-sdk/baseapp/msgCounter2", nil)
	cdc.RegisterConcrete(&msgKeyValue{}, "cosmos-sdk/baseapp/msgKeyValue", nil)
	cdc.RegisterConcrete(&msgNoRoute{}, "cosmos-sdk/baseapp/msgNoRoute", nil)
}

// aminoTxEncoder creates a amino TxEncoder for testing purposes.
func aminoTxEncoder() sdk.TxEncoder {
	cdc := codec.NewLegacyAmino()
	registerTestCodec(cdc)

	return legacytx.StdTxConfig{Cdc: cdc}.TxEncoder()
}

// simple one store baseapp
func setupBaseApp(t *testing.T, options ...func(*BaseApp)) *BaseApp {
	app := newBaseApp(t.Name(), options...)
	require.Equal(t, t.Name(), app.Name())

	app.MountStores(capKey1, capKey2)
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})

	// stores are mounted
	err := app.LoadLatestVersion()
	require.Nil(t, err)
	return app
}

func TestLoadVersionPruning(t *testing.T) {
	logger := log.NewNopLogger()
	pruningOptions := store.PruningOptions{
		KeepRecent: 2,
		KeepEvery:  3,
		Interval:   1,
	}
	pruningOpt := SetPruning(pruningOptions)
	db := dbm.NewMemDB()
	name := t.Name()
	app := NewBaseApp(name, logger, db, nil, pruningOpt)

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("key1")
	app.MountStores(capKey)

	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := sdk.CommitID{}

	// fresh store has zero/empty last commit
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	var lastCommitID sdk.CommitID

	// Commit seven blocks, of which 7 (latest) is kept in addition to 6, 5
	// (keep recent) and 3 (keep every).
	for i := int64(1); i <= 7; i++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: i}})
		res := app.Commit()
		lastCommitID = sdk.CommitID{Version: i, Hash: res.Data}
	}

	for _, v := range []int64{1, 2, 4} {
		_, err = app.cms.CacheMultiStoreWithVersion(v)
		require.Error(t, err)
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

func testLoadVersionHelper(t *testing.T, app *BaseApp, expectedHeight int64, expectedID sdk.CommitID) {
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

func TestSetMinGasPrices(t *testing.T) {
	minGasPrices := sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5000)}
	app := newBaseApp(t.Name(), SetMinGasPrices(minGasPrices.String()))
	require.Equal(t, minGasPrices, app.minGasPrices)
}

func TestGetMaximumBlockGas(t *testing.T) {
	app := setupBaseApp(t)
	app.InitChain(abci.RequestInitChain{})
	ctx := app.NewContext(true, tmproto.Header{})

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 0}})
	require.Equal(t, uint64(0), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: -1}})
	require.Equal(t, uint64(0), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 5000000}})
	require.Equal(t, uint64(5000000), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: -5000000}})
	require.Panics(t, func() { app.getMaximumBlockGas(ctx) })
}

func TestListSnapshots(t *testing.T) {
	type setupConfig struct {
		blocks            uint64
		blockTxs          int
		snapshotInterval  uint64
		snapshotKeepEvery uint32
	}

	app, _ := setupBaseAppWithSnapshots(t, 2, 5)

	expected := abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{
		{Height: 2, Format: 2, Chunks: 2},
	}}

	resp := app.ListSnapshots(abci.RequestListSnapshots{})
	queryResponse := app.Query(abci.RequestQuery{
		Path: "/app/snapshots",
	})

	queryListSnapshotsResp := abci.ResponseListSnapshots{}
	err := json.Unmarshal(queryResponse.Value, &queryListSnapshotsResp)
	require.NoError(t, err)

	for i, s := range resp.Snapshots {
		querySnapshot := queryListSnapshotsResp.Snapshots[i]
		// we check that the query snapshot and function snapshot are equal
		// Then we check that the hash and metadata are not empty. We atm
		// do not have a good way to generate the expected value for these.
		assert.Equal(t, *s, *querySnapshot)
		assert.NotEmpty(t, s.Hash)
		assert.NotEmpty(t, s.Metadata)
		// Set hash and metadata to nil, so we can check the other snapshot
		// fields against expected
		s.Hash = nil
		s.Metadata = nil
	}
	assert.Equal(t, expected, resp)
}
