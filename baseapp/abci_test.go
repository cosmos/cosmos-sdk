package baseapp_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/pruning/types"
	snapshottypes "github.com/cosmos/cosmos-sdk/store/snapshots/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestABCI_Info(t *testing.T) {
	suite := NewBaseAppSuite(t)

	reqInfo := abci.RequestInfo{}
	res := suite.baseApp.Info(reqInfo)

	require.Equal(t, "", res.Version)
	require.Equal(t, t.Name(), res.GetData())
	require.Equal(t, int64(0), res.LastBlockHeight)
	require.Equal(t, []uint8(nil), res.LastBlockAppHash)
	require.Equal(t, suite.baseApp.AppVersion(), res.AppVersion)
}

func TestABCI_InitChain(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	logger := defaultLogger()
	app := baseapp.NewBaseApp(name, logger, db, nil)

	capKey := sdk.NewKVStoreKey("main")
	capKey2 := sdk.NewKVStoreKey("key2")
	app.MountStores(capKey, capKey2)

	// set a value in the store on init chain
	key, value := []byte("hello"), []byte("goodbye")
	var initChainer sdk.InitChainer = func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		store := ctx.KVStore(capKey)
		store.Set(key, value)
		return abci.ResponseInitChain{}
	}

	query := abci.RequestQuery{
		Path: "/store/main/key",
		Data: key,
	}

	// initChain is nil - nothing happens
	app.InitChain(abci.RequestInitChain{})
	res := app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// set initChainer and try again - should see the value
	app.SetInitChainer(initChainer)

	// stores are mounted and private members are set - sealing baseapp
	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)
	require.Equal(t, int64(0), app.LastBlockHeight())

	initChainRes := app.InitChain(abci.RequestInitChain{AppStateBytes: []byte("{}"), ChainId: "test-chain-id"}) // must have valid JSON genesis file, even if empty

	// The AppHash returned by a new chain is the sha256 hash of "".
	// $ echo -n '' | sha256sum
	// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	require.Equal(
		t,
		[]byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55},
		initChainRes.AppHash,
	)

	// assert that chainID is set correctly in InitChain
	chainID := getDeliverStateCtx(app).ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in deliverState not set correctly in InitChain")

	chainID = getCheckStateCtx(app).ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in checkState not set correctly in InitChain")

	app.Commit()
	res = app.Query(query)
	require.Equal(t, int64(1), app.LastBlockHeight())
	require.Equal(t, value, res.Value)

	// reload app
	app = baseapp.NewBaseApp(name, logger, db, nil)
	app.SetInitChainer(initChainer)
	app.MountStores(capKey, capKey2)
	err = app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)
	require.Equal(t, int64(1), app.LastBlockHeight())

	// ensure we can still query after reloading
	res = app.Query(query)
	require.Equal(t, value, res.Value)

	// commit and ensure we can still query
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()

	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

func TestABCI_InitChain_WithInitialHeight(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	logger := defaultLogger()
	app := baseapp.NewBaseApp(name, logger, db, nil)

	app.InitChain(
		abci.RequestInitChain{
			InitialHeight: 3,
		},
	)
	app.Commit()

	require.Equal(t, int64(3), app.LastBlockHeight())
}

func TestABCI_BeginBlock_WithInitialHeight(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	logger := defaultLogger()
	app := baseapp.NewBaseApp(name, logger, db, nil)

	app.InitChain(
		abci.RequestInitChain{
			InitialHeight: 3,
		},
	)

	require.PanicsWithError(t, "invalid height: 4; expected: 3", func() {
		app.BeginBlock(abci.RequestBeginBlock{
			Header: tmproto.Header{
				Height: 4,
			},
		})
	})

	app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height: 3,
		},
	})
	app.Commit()

	require.Equal(t, int64(3), app.LastBlockHeight())
}

func TestGRPCQuery(t *testing.T) {
	grpcQueryOpt := func(bapp *baseapp.BaseApp) {
		testdata.RegisterQueryServer(
			bapp.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	}

	suite := NewBaseAppSuite(t, grpcQueryOpt)

	suite.baseApp.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{},
	})

	header := tmproto.Header{Height: suite.baseApp.LastBlockHeight() + 1}
	suite.baseApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	suite.baseApp.Commit()

	req := testdata.SayHelloRequest{Name: "foo"}
	reqBz, err := req.Marshal()
	require.NoError(t, err)

	reqQuery := abci.RequestQuery{
		Data: reqBz,
		Path: "/testdata.Query/SayHello",
	}

	resQuery := suite.baseApp.Query(reqQuery)
	require.Equal(t, abci.CodeTypeOK, resQuery.Code, resQuery)

	var res testdata.SayHelloResponse
	require.NoError(t, res.Unmarshal(resQuery.Value))
	require.Equal(t, "Hello foo!", res.Greeting)
}

func TestABCI_P2PQuery(t *testing.T) {
	addrPeerFilterOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAddrPeerFilter(func(addrport string) abci.ResponseQuery {
			require.Equal(t, "1.1.1.1:8000", addrport)
			return abci.ResponseQuery{Code: uint32(3)}
		})
	}

	idPeerFilterOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetIDPeerFilter(func(id string) abci.ResponseQuery {
			require.Equal(t, "testid", id)
			return abci.ResponseQuery{Code: uint32(4)}
		})
	}

	suite := NewBaseAppSuite(t, addrPeerFilterOpt, idPeerFilterOpt)

	addrQuery := abci.RequestQuery{
		Path: "/p2p/filter/addr/1.1.1.1:8000",
	}
	res := suite.baseApp.Query(addrQuery)
	require.Equal(t, uint32(3), res.Code)

	idQuery := abci.RequestQuery{
		Path: "/p2p/filter/id/testid",
	}
	res = suite.baseApp.Query(idQuery)
	require.Equal(t, uint32(4), res.Code)
}

func TestABCI_ListSnapshots(t *testing.T) {
	ssCfg := SnapshotsConfig{
		blocks:             5,
		blockTxs:           4,
		snapshotInterval:   2,
		snapshotKeepRecent: 2,
		pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
	}

	suite := NewBaseAppSuiteWithSnapshots(t, ssCfg)

	resp := suite.baseApp.ListSnapshots(abci.RequestListSnapshots{})
	for _, s := range resp.Snapshots {
		require.NotEmpty(t, s.Hash)
		require.NotEmpty(t, s.Metadata)

		s.Hash = nil
		s.Metadata = nil
	}

	require.Equal(t, abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{
		{Height: 4, Format: snapshottypes.CurrentFormat, Chunks: 2},
		{Height: 2, Format: snapshottypes.CurrentFormat, Chunks: 1},
	}}, resp)
}

func TestABCI_SnapshotWithPruning(t *testing.T) {
	testCases := map[string]struct {
		ssCfg             SnapshotsConfig
		expectedSnapshots []*abci.Snapshot
	}{
		"prune nothing with snapshot": {
			ssCfg: SnapshotsConfig{
				blocks:             20,
				blockTxs:           2,
				snapshotInterval:   5,
				snapshotKeepRecent: 1,
				pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 20, Format: snapshottypes.CurrentFormat, Chunks: 5},
			},
		},
		"prune everything with snapshot": {
			ssCfg: SnapshotsConfig{
				blocks:             20,
				blockTxs:           2,
				snapshotInterval:   5,
				snapshotKeepRecent: 1,
				pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningEverything),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 20, Format: snapshottypes.CurrentFormat, Chunks: 5},
			},
		},
		"default pruning with snapshot": {
			ssCfg: SnapshotsConfig{
				blocks:             20,
				blockTxs:           2,
				snapshotInterval:   5,
				snapshotKeepRecent: 1,
				pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningDefault),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 20, Format: snapshottypes.CurrentFormat, Chunks: 5},
			},
		},
		"custom": {
			ssCfg: SnapshotsConfig{
				blocks:             25,
				blockTxs:           2,
				snapshotInterval:   5,
				snapshotKeepRecent: 2,
				pruningOpts:        pruningtypes.NewCustomPruningOptions(12, 12),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 25, Format: snapshottypes.CurrentFormat, Chunks: 6},
				{Height: 20, Format: snapshottypes.CurrentFormat, Chunks: 5},
			},
		},
		"no snapshots": {
			ssCfg: SnapshotsConfig{
				blocks:           10,
				blockTxs:         2,
				snapshotInterval: 0, // 0 implies disable snapshots
				pruningOpts:      pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
			},
			expectedSnapshots: []*abci.Snapshot{},
		},
		"keep all snapshots": {
			ssCfg: SnapshotsConfig{
				blocks:             10,
				blockTxs:           2,
				snapshotInterval:   3,
				snapshotKeepRecent: 0, // 0 implies keep all snapshots
				pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 9, Format: snapshottypes.CurrentFormat, Chunks: 2},
				{Height: 6, Format: snapshottypes.CurrentFormat, Chunks: 2},
				{Height: 3, Format: snapshottypes.CurrentFormat, Chunks: 1},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			suite := NewBaseAppSuiteWithSnapshots(t, tc.ssCfg)

			resp := suite.baseApp.ListSnapshots(abci.RequestListSnapshots{})
			for _, s := range resp.Snapshots {
				require.NotEmpty(t, s.Hash)
				require.NotEmpty(t, s.Metadata)

				s.Hash = nil
				s.Metadata = nil
			}

			require.Equal(t, abci.ResponseListSnapshots{Snapshots: tc.expectedSnapshots}, resp)

			// Validate that heights were pruned correctly by querying the state at the last height that should be present relative to latest
			// and the first height that should be pruned.
			//
			// Exceptions:
			//   * Prune nothing: should be able to query all heights (we only test first and latest)
			//   * Prune default: should be able to query all heights (we only test first and latest)
			//      * The reason for default behaving this way is that we only commit 20 heights but default has 100_000 keep-recent
			var lastExistingHeight int64
			if tc.ssCfg.pruningOpts.GetPruningStrategy() == pruningtypes.PruningNothing || tc.ssCfg.pruningOpts.GetPruningStrategy() == pruningtypes.PruningDefault {
				lastExistingHeight = 1
			} else {
				// Integer division rounds down so by multiplying back we get the last height at which we pruned
				lastExistingHeight = int64((tc.ssCfg.blocks/tc.ssCfg.pruningOpts.Interval)*tc.ssCfg.pruningOpts.Interval - tc.ssCfg.pruningOpts.KeepRecent)
			}

			// Query 1
			res := suite.baseApp.Query(abci.RequestQuery{Path: fmt.Sprintf("/store/%s/key", capKey2.Name()), Data: []byte("0"), Height: lastExistingHeight})
			require.NotNil(t, res, "height: %d", lastExistingHeight)
			require.NotNil(t, res.Value, "height: %d", lastExistingHeight)

			// Query 2
			res = suite.baseApp.Query(abci.RequestQuery{Path: fmt.Sprintf("/store/%s/key", capKey2.Name()), Data: []byte("0"), Height: lastExistingHeight - 1})
			require.NotNil(t, res, "height: %d", lastExistingHeight-1)

			if tc.ssCfg.pruningOpts.GetPruningStrategy() == pruningtypes.PruningNothing || tc.ssCfg.pruningOpts.GetPruningStrategy() == pruningtypes.PruningDefault {
				// With prune nothing or default, we query height 0 which translates to the latest height.
				require.NotNil(t, res.Value, "height: %d", lastExistingHeight-1)
			}
		})
	}
}

func TestABCI_LoadSnapshotChunk(t *testing.T) {
	ssCfg := SnapshotsConfig{
		blocks:             2,
		blockTxs:           5,
		snapshotInterval:   2,
		snapshotKeepRecent: snapshottypes.CurrentFormat,
		pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
	}
	suite := NewBaseAppSuiteWithSnapshots(t, ssCfg)

	testCases := map[string]struct {
		height      uint64
		format      uint32
		chunk       uint32
		expectEmpty bool
	}{
		"Existing snapshot": {2, snapshottypes.CurrentFormat, 1, false},
		"Missing height":    {100, snapshottypes.CurrentFormat, 1, true},
		"Missing format":    {2, snapshottypes.CurrentFormat + 1, 1, true},
		"Missing chunk":     {2, snapshottypes.CurrentFormat, 9, true},
		"Zero height":       {0, snapshottypes.CurrentFormat, 1, true},
		"Zero format":       {2, 0, 1, true},
		"Zero chunk":        {2, snapshottypes.CurrentFormat, 0, false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			resp := suite.baseApp.LoadSnapshotChunk(abci.RequestLoadSnapshotChunk{
				Height: tc.height,
				Format: tc.format,
				Chunk:  tc.chunk,
			})
			if tc.expectEmpty {
				require.Equal(t, abci.ResponseLoadSnapshotChunk{}, resp)
				return
			}

			require.NotEmpty(t, resp.Chunk)
		})
	}
}

func TestABCI_OfferSnapshot_Errors(t *testing.T) {
	ssCfg := SnapshotsConfig{
		blocks:             0,
		blockTxs:           0,
		snapshotInterval:   2,
		snapshotKeepRecent: 2,
		pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
	}
	suite := NewBaseAppSuiteWithSnapshots(t, ssCfg)

	m := snapshottypes.Metadata{ChunkHashes: [][]byte{{1}, {2}, {3}}}
	metadata, err := m.Marshal()
	require.NoError(t, err)

	hash := []byte{1, 2, 3}

	testCases := map[string]struct {
		snapshot *abci.Snapshot
		result   abci.ResponseOfferSnapshot_Result
	}{
		"nil snapshot": {nil, abci.ResponseOfferSnapshot_REJECT},
		"invalid format": {&abci.Snapshot{
			Height: 1, Format: 9, Chunks: 3, Hash: hash, Metadata: metadata,
		}, abci.ResponseOfferSnapshot_REJECT_FORMAT},
		"incorrect chunk count": {&abci.Snapshot{
			Height: 1, Format: snapshottypes.CurrentFormat, Chunks: 2, Hash: hash, Metadata: metadata,
		}, abci.ResponseOfferSnapshot_REJECT},
		"no chunks": {&abci.Snapshot{
			Height: 1, Format: snapshottypes.CurrentFormat, Chunks: 0, Hash: hash, Metadata: metadata,
		}, abci.ResponseOfferSnapshot_REJECT},
		"invalid metadata serialization": {&abci.Snapshot{
			Height: 1, Format: snapshottypes.CurrentFormat, Chunks: 0, Hash: hash, Metadata: []byte{3, 1, 4},
		}, abci.ResponseOfferSnapshot_REJECT},
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			resp := suite.baseApp.OfferSnapshot(abci.RequestOfferSnapshot{Snapshot: tc.snapshot})
			require.Equal(t, tc.result, resp.Result)
		})
	}

	// Offering a snapshot after one has been accepted should error
	resp := suite.baseApp.OfferSnapshot(abci.RequestOfferSnapshot{Snapshot: &abci.Snapshot{
		Height:   1,
		Format:   snapshottypes.CurrentFormat,
		Chunks:   3,
		Hash:     []byte{1, 2, 3},
		Metadata: metadata,
	}})
	require.Equal(t, abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ACCEPT}, resp)

	resp = suite.baseApp.OfferSnapshot(abci.RequestOfferSnapshot{Snapshot: &abci.Snapshot{
		Height:   2,
		Format:   snapshottypes.CurrentFormat,
		Chunks:   3,
		Hash:     []byte{1, 2, 3},
		Metadata: metadata,
	}})
	require.Equal(t, abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}, resp)
}
