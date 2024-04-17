package baseapp_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/testutil/mock"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	failStr    = "&failOnAnte=false"
	fooStr     = "foo"
	counterStr = "counter="
)

func TestABCI_InitChain(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	app := baseapp.NewBaseApp(name, logger, db, nil, baseapp.SetChainID("test-chain-id"))

	capKey := storetypes.NewKVStoreKey("main")
	capKey2 := storetypes.NewKVStoreKey("key2")
	app.MountStores(capKey, capKey2)

	// set a value in the store on init chain
	key, value := []byte("hello"), []byte("goodbye")
	var initChainer sdk.InitChainer = func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		store := ctx.KVStore(capKey)
		store.Set(key, value)
		return &abci.ResponseInitChain{}, nil
	}

	query := abci.RequestQuery{
		Path: "/store/main/key",
		Data: key,
	}

	// initChain is nil and chain ID is wrong - errors
	_, err := app.InitChain(&abci.RequestInitChain{ChainId: "wrong-chain-id"})
	require.Error(t, err)

	// initChain is nil - nothing happens
	_, err = app.InitChain(&abci.RequestInitChain{ChainId: "test-chain-id"})
	require.NoError(t, err)
	resQ, err := app.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, 0, len(resQ.Value))

	// set initChainer and try again - should see the value
	app.SetInitChainer(initChainer)

	// stores are mounted and private members are set - sealing baseapp
	err = app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)
	require.Equal(t, int64(0), app.LastBlockHeight())

	initChainRes, err := app.InitChain(&abci.RequestInitChain{AppStateBytes: []byte("{}"), ChainId: "test-chain-id"}) // must have valid JSON genesis file, even if empty
	require.NoError(t, err)

	// The AppHash returned by a new chain is the sha256 hash of "".
	// $ echo -n '' | sha256sum
	// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	apphash, err := hex.DecodeString("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	require.NoError(t, err)
	emptyHash := sha256.Sum256([]byte{})
	require.Equal(t, emptyHash[:], apphash)

	require.Equal(t, apphash, initChainRes.AppHash)

	// assert that chainID is set correctly in InitChain
	chainID := baseapptestutil.GetFinalizeBlockStateCtx(app).ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in deliverState not set correctly in InitChain")

	chainID = baseapptestutil.GetCheckStateCtx(app).ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in checkState not set correctly in InitChain")

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Hash:   initChainRes.AppHash,
		Height: 1,
	})
	require.NoError(t, err)

	_, err = app.Commit()
	require.NoError(t, err)

	resQ, err = app.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, int64(1), app.LastBlockHeight())
	require.Equal(t, value, resQ.Value)

	// reload app
	app = baseapp.NewBaseApp(name, logger, db, nil)
	app.SetInitChainer(initChainer)
	app.MountStores(capKey, capKey2)
	err = app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)
	require.Equal(t, int64(1), app.LastBlockHeight())

	// ensure we can still query after reloading
	resQ, err = app.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, value, resQ.Value)

	// commit and ensure we can still query
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	resQ, err = app.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, value, resQ.Value)
}

func TestABCI_InitChain_WithInitialHeight(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	_, err := app.InitChain(
		&abci.RequestInitChain{
			InitialHeight: 3,
		},
	)
	require.NoError(t, err)

	_, err = app.Commit()
	require.NoError(t, err)
	require.Equal(t, int64(3), app.LastBlockHeight())
}

func TestABCI_FinalizeBlock_WithInitialHeight(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	_, err := app.InitChain(
		&abci.RequestInitChain{
			InitialHeight: 3,
		},
	)
	require.NoError(t, err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 4})
	require.Error(t, err, "invalid height: 4; expected: 3")

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 3})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
	require.Equal(t, int64(3), app.LastBlockHeight())
}

func TestABCI_FinalizeBlock_WithBeginAndEndBlocker(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	app.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
		return sdk.BeginBlock{
			Events: []abci.Event{
				{
					Type: "sometype",
					Attributes: []abci.EventAttribute{
						{
							Key:   fooStr,
							Value: "bar",
						},
					},
				},
			},
		}, nil
	})

	app.SetEndBlocker(func(ctx sdk.Context) (sdk.EndBlock, error) {
		return sdk.EndBlock{
			Events: []abci.Event{
				{
					Type: "anothertype",
					Attributes: []abci.EventAttribute{
						{
							Key:   fooStr,
							Value: "bar",
						},
					},
				},
			},
		}, nil
	})

	_, err := app.InitChain(
		&abci.RequestInitChain{
			InitialHeight: 1,
		},
	)
	require.NoError(t, err)

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)

	require.Len(t, res.Events, 2)

	require.Equal(t, "sometype", res.Events[0].Type)
	require.Equal(t, fooStr, res.Events[0].Attributes[0].Key)
	require.Equal(t, "bar", res.Events[0].Attributes[0].Value)
	require.Equal(t, "mode", res.Events[0].Attributes[1].Key)
	require.Equal(t, "BeginBlock", res.Events[0].Attributes[1].Value)

	require.Equal(t, "anothertype", res.Events[1].Type)
	require.Equal(t, fooStr, res.Events[1].Attributes[0].Key)
	require.Equal(t, "bar", res.Events[1].Attributes[0].Value)
	require.Equal(t, "mode", res.Events[1].Attributes[1].Key)
	require.Equal(t, "EndBlock", res.Events[1].Attributes[1].Value)

	_, err = app.Commit()
	require.NoError(t, err)

	require.Equal(t, int64(1), app.LastBlockHeight())
}

func TestABCI_ExtendVote(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	app.SetExtendVoteHandler(func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		voteExt := fooStr + hex.EncodeToString(req.Hash) + strconv.FormatInt(req.Height, 10)
		return &abci.ResponseExtendVote{VoteExtension: []byte(voteExt)}, nil
	})

	app.SetVerifyVoteExtensionHandler(func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		// do some kind of verification here
		expectedVoteExt := fooStr + hex.EncodeToString(req.Hash) + strconv.FormatInt(req.Height, 10)
		if !bytes.Equal(req.VoteExtension, []byte(expectedVoteExt)) {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	})

	app.SetParamStore(mock.NewMockParamStore(dbm.NewMemDB()))
	_, err := app.InitChain(
		&abci.RequestInitChain{
			InitialHeight: 1,
			ConsensusParams: &cmtproto.ConsensusParams{
				Abci: &cmtproto.ABCIParams{
					VoteExtensionsEnableHeight: 200,
				},
			},
		},
	)
	require.NoError(t, err)

	// Votes not enabled yet
	_, err = app.ExtendVote(context.Background(), &abci.RequestExtendVote{Height: 123, Hash: []byte("thehash")})
	require.ErrorContains(t, err, "vote extensions are not enabled")

	// First vote on the first enabled height
	res, err := app.ExtendVote(context.Background(), &abci.RequestExtendVote{Height: 200, Hash: []byte("thehash")})
	require.NoError(t, err)
	require.Len(t, res.VoteExtension, 20)

	res, err = app.ExtendVote(context.Background(), &abci.RequestExtendVote{Height: 1000, Hash: []byte("thehash")})
	require.NoError(t, err)
	require.Len(t, res.VoteExtension, 21)

	// Error during vote extension should return an empty vote extension and no error
	app.SetExtendVoteHandler(func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		return nil, errors.New("some error")
	})
	res, err = app.ExtendVote(context.Background(), &abci.RequestExtendVote{Height: 1000, Hash: []byte("thehash")})
	require.NoError(t, err)
	require.Len(t, res.VoteExtension, 0)

	// Verify Vote Extensions
	_, err = app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 123, VoteExtension: []byte("1234567")})
	require.ErrorContains(t, err, "vote extensions are not enabled")

	// First vote on the first enabled height
	vres, err := app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 200, Hash: []byte("thehash"), VoteExtension: []byte("foo74686568617368200")})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_ACCEPT, vres.Status)

	vres, err = app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 1000, Hash: []byte("thehash"), VoteExtension: []byte("foo746865686173681000")})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_ACCEPT, vres.Status)

	// Reject because it's just some random bytes
	vres, err = app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 201, Hash: []byte("thehash"), VoteExtension: []byte("12345678")})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, vres.Status)

	// Reject because the verification failed (no error)
	app.SetVerifyVoteExtensionHandler(func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		return nil, errors.New("some error")
	})
	vres, err = app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 201, Hash: []byte("thehash"), VoteExtension: []byte("12345678")})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, vres.Status)
}

// TestABCI_OnlyVerifyVoteExtension makes sure we can call VerifyVoteExtension
// without having called ExtendVote before.
func TestABCI_OnlyVerifyVoteExtension(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	app.SetVerifyVoteExtensionHandler(func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		// do some kind of verification here
		expectedVoteExt := fooStr + hex.EncodeToString(req.Hash) + strconv.FormatInt(req.Height, 10)
		if !bytes.Equal(req.VoteExtension, []byte(expectedVoteExt)) {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	})

	app.SetParamStore(mock.NewMockParamStore(dbm.NewMemDB()))
	_, err := app.InitChain(
		&abci.RequestInitChain{
			InitialHeight: 1,
			ConsensusParams: &cmtproto.ConsensusParams{
				Abci: &cmtproto.ABCIParams{
					VoteExtensionsEnableHeight: 200,
				},
			},
		},
	)
	require.NoError(t, err)

	// Verify Vote Extensions
	_, err = app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 123, VoteExtension: []byte("1234567")})
	require.ErrorContains(t, err, "vote extensions are not enabled")

	// First vote on the first enabled height
	vres, err := app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 200, Hash: []byte("thehash"), VoteExtension: []byte("foo74686568617368200")})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_ACCEPT, vres.Status)

	vres, err = app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 1000, Hash: []byte("thehash"), VoteExtension: []byte("foo746865686173681000")})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_ACCEPT, vres.Status)

	// Reject because it's just some random bytes
	vres, err = app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 201, Hash: []byte("thehash"), VoteExtension: []byte("12345678")})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, vres.Status)

	// Reject because the verification failed (no error)
	app.SetVerifyVoteExtensionHandler(func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		return nil, errors.New("some error")
	})
	vres, err = app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: 201, Hash: []byte("thehash"), VoteExtension: []byte("12345678")})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, vres.Status)
}

func TestBaseApp_PrepareCheckState(t *testing.T) {
	db := dbm.NewMemDB()
	name := t.Name()
	logger := log.NewTestLogger(t)

	cp := &cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxGas: 5000000,
		},
	}

	app := baseapp.NewBaseApp(name, logger, db, nil)
	app.SetParamStore(mock.NewMockParamStore(dbm.NewMemDB()))
	_, err := app.InitChain(&abci.RequestInitChain{
		ConsensusParams: cp,
	})
	require.NoError(t, err)

	wasPrepareCheckStateCalled := false
	app.SetPrepareCheckStater(func(ctx sdk.Context) {
		wasPrepareCheckStateCalled = true
	})
	app.Seal()

	_, err = app.Commit()
	require.NoError(t, err)
	require.Equal(t, true, wasPrepareCheckStateCalled)
}

func TestBaseApp_Precommit(t *testing.T) {
	db := dbm.NewMemDB()
	name := t.Name()
	logger := log.NewTestLogger(t)

	cp := &cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxGas: 5000000,
		},
	}

	app := baseapp.NewBaseApp(name, logger, db, nil)
	app.SetParamStore(mock.NewMockParamStore(dbm.NewMemDB()))
	_, err := app.InitChain(&abci.RequestInitChain{
		ConsensusParams: cp,
	})
	require.NoError(t, err)

	wasPrecommiterCalled := false
	app.SetPrecommiter(func(ctx sdk.Context) {
		wasPrecommiterCalled = true
	})
	app.Seal()

	_, err = app.Commit()
	require.NoError(t, err)
	require.Equal(t, true, wasPrecommiterCalled)
}

func TestABCI_GetBlockRetentionHeight(t *testing.T) {
	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()
	name := t.Name()

	snapshotStore, err := snapshots.NewStore(dbm.NewMemDB(), testutil.GetTempDir(t))
	require.NoError(t, err)

	testCases := map[string]struct {
		bapp         *baseapp.BaseApp
		maxAgeBlocks int64
		commitHeight int64
		expected     int64
	}{
		"defaults": {
			bapp:         baseapp.NewBaseApp(name, logger, db, nil),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     0,
		},
		"pruning unbonding time only": {
			bapp:         baseapp.NewBaseApp(name, logger, db, nil, baseapp.SetMinRetainBlocks(1)),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     136120,
		},
		"pruning iavl snapshot only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing)),
				baseapp.SetMinRetainBlocks(1),
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(10000, 1)),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     489000,
		},
		"pruning state sync snapshot only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
				baseapp.SetMinRetainBlocks(1),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     349000,
		},
		"pruning min retention only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetMinRetainBlocks(400000),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     99000,
		},
		"pruning all conditions": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetPruning(pruningtypes.NewCustomPruningOptions(0, 0)),
				baseapp.SetMinRetainBlocks(400000),
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     99000,
		},
		"no pruning due to no persisted state": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetPruning(pruningtypes.NewCustomPruningOptions(0, 0)),
				baseapp.SetMinRetainBlocks(400000),
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 10000,
			expected:     0,
		},
		"disable pruning": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetPruning(pruningtypes.NewCustomPruningOptions(0, 0)),
				baseapp.SetMinRetainBlocks(0),
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     0,
		},
	}

	for name, tc := range testCases {
		tc := tc

		tc.bapp.SetParamStore(mock.NewMockParamStore(dbm.NewMemDB()))
		_, err := tc.bapp.InitChain(&abci.RequestInitChain{
			ConsensusParams: &cmtproto.ConsensusParams{
				Evidence: &cmtproto.EvidenceParams{
					MaxAgeNumBlocks: tc.maxAgeBlocks,
				},
			},
		})
		require.NoError(t, err)

		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.bapp.GetBlockRetentionHeight(tc.commitHeight))
		})
	}
}

// Verifies that PrepareCheckState is called with the checkState.
func TestPrepareCheckStateCalledWithCheckState(t *testing.T) {
	t.Parallel()

	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, nil)

	wasPrepareCheckStateCalled := false
	app.SetPrepareCheckStater(func(ctx sdk.Context) {
		require.Equal(t, true, ctx.IsCheckTx())
		wasPrepareCheckStateCalled = true
	})

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	require.Equal(t, true, wasPrepareCheckStateCalled)
}

// Verifies that the Precommiter is called with the deliverState.
func TestPrecommiterCalledWithDeliverState(t *testing.T) {
	t.Parallel()

	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, nil)

	wasPrecommiterCalled := false
	app.SetPrecommiter(func(ctx sdk.Context) {
		require.Equal(t, false, ctx.IsCheckTx())
		require.Equal(t, false, ctx.IsReCheckTx())
		wasPrecommiterCalled = true
	})

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	require.Equal(t, true, wasPrecommiterCalled)
}

func TestBaseApp_PreBlocker(t *testing.T) {
	db := dbm.NewMemDB()
	name := t.Name()
	logger := log.NewTestLogger(t)

	app := baseapp.NewBaseApp(name, logger, db, nil)
	_, err := app.InitChain(&abci.RequestInitChain{})
	require.NoError(t, err)

	wasHookCalled := false
	app.SetPreBlocker(func(ctx sdk.Context, req *abci.RequestFinalizeBlock) error {
		wasHookCalled = true
		return nil
	})
	app.Seal()

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	require.Equal(t, true, wasHookCalled)

	// Now try erroring
	app = baseapp.NewBaseApp(name, logger, db, nil)
	_, err = app.InitChain(&abci.RequestInitChain{})
	require.NoError(t, err)

	app.SetPreBlocker(func(ctx sdk.Context, req *abci.RequestFinalizeBlock) error {
		return errors.New("some error")
	})
	app.Seal()

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.Error(t, err)
}
