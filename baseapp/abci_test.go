package baseapp_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/baseapp/testutil/mock"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func TestABCI_Info(t *testing.T) {
	suite := NewBaseAppSuite(t)
	ctx := suite.baseApp.NewContext(true)
	err := suite.baseApp.StoreConsensusParams(ctx, cmttypes.DefaultConsensusParams().ToProto())
	require.NoError(t, err)

	reqInfo := abci.RequestInfo{}
	res, err := suite.baseApp.Info(&reqInfo)
	require.NoError(t, err)

	require.Equal(t, "", res.Version)
	require.Equal(t, t.Name(), res.GetData())
	require.Equal(t, int64(0), res.LastBlockHeight)
	require.Equal(t, []uint8(nil), res.LastBlockAppHash)
	appVersion, err := suite.baseApp.AppVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, appVersion, res.AppVersion)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = suite.baseApp.Commit()
	require.NoError(t, err)
	require.NoError(t, suite.baseApp.SetAppVersion(ctx, 1))
	res, err = suite.baseApp.Info(&reqInfo)
	require.NoError(t, err)
	require.Equal(t, uint64(1), res.AppVersion)
}

func TestABCI_First_block_Height(t *testing.T) {
	suite := NewBaseAppSuite(t, baseapp.SetChainID("test-chain-id"))
	app := suite.baseApp

	_, err := app.InitChain(&abci.RequestInitChain{
		ChainId:         "test-chain-id",
		ConsensusParams: &cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: 5000000}},
		InitialHeight:   1,
	})
	require.NoError(t, err)

	_, err = app.Commit()
	require.NoError(t, err)

	ctx := app.GetContextForCheckTx(nil)
	require.Equal(t, int64(1), ctx.BlockHeight())
}

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

	require.Equal(
		t,
		apphash,
		initChainRes.AppHash,
	)

	// assert that chainID is set correctly in InitChain
	chainID := getFinalizeBlockStateCtx(app).ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in deliverState not set correctly in InitChain")

	chainID = getCheckStateCtx(app).ChainID()
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
							Key:   "foo",
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
							Key:   "foo",
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
	require.Equal(t, "foo", res.Events[0].Attributes[0].Key)
	require.Equal(t, "bar", res.Events[0].Attributes[0].Value)
	require.Equal(t, "mode", res.Events[0].Attributes[1].Key)
	require.Equal(t, "BeginBlock", res.Events[0].Attributes[1].Value)

	require.Equal(t, "anothertype", res.Events[1].Type)
	require.Equal(t, "foo", res.Events[1].Attributes[0].Key)
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
		voteExt := "foo" + hex.EncodeToString(req.Hash) + strconv.FormatInt(req.Height, 10)
		return &abci.ResponseExtendVote{VoteExtension: []byte(voteExt)}, nil
	})

	app.SetVerifyVoteExtensionHandler(func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		// do some kind of verification here
		expectedVoteExt := "foo" + hex.EncodeToString(req.Hash) + strconv.FormatInt(req.Height, 10)
		if !bytes.Equal(req.VoteExtension, []byte(expectedVoteExt)) {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	})

	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})
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
		expectedVoteExt := "foo" + hex.EncodeToString(req.Hash) + strconv.FormatInt(req.Height, 10)
		if !bytes.Equal(req.VoteExtension, []byte(expectedVoteExt)) {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	})

	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})
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

func TestABCI_GRPCQuery(t *testing.T) {
	grpcQueryOpt := func(bapp *baseapp.BaseApp) {
		testdata.RegisterQueryServer(
			bapp.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	}

	suite := NewBaseAppSuite(t, grpcQueryOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	req := testdata.SayHelloRequest{Name: "foo"}
	reqBz, err := req.Marshal()
	require.NoError(t, err)

	resQuery, err := suite.baseApp.Query(context.TODO(), &abci.RequestQuery{
		Data: reqBz,
		Path: "/testpb.Query/SayHello",
	})
	require.NoError(t, err)
	require.Equal(t, sdkerrors.ErrInvalidHeight.ABCICode(), resQuery.Code, resQuery)
	require.Contains(t, resQuery.Log, "TestABCI_GRPCQuery is not ready; please wait for first block")

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: suite.baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)

	_, err = suite.baseApp.Commit()
	require.NoError(t, err)

	reqQuery := abci.RequestQuery{
		Data: reqBz,
		Path: "/testpb.Query/SayHello",
	}

	resQuery, err = suite.baseApp.Query(context.TODO(), &reqQuery)
	require.NoError(t, err)
	require.Equal(t, abci.CodeTypeOK, resQuery.Code, resQuery)

	var res testdata.SayHelloResponse
	require.NoError(t, res.Unmarshal(resQuery.Value))
	require.Equal(t, "Hello foo!", res.Greeting)
}

func TestABCI_P2PQuery(t *testing.T) {
	addrPeerFilterOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAddrPeerFilter(func(addrport string) *abci.ResponseQuery {
			require.Equal(t, "1.1.1.1:8000", addrport)
			return &abci.ResponseQuery{Code: uint32(3)}
		})
	}

	idPeerFilterOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetIDPeerFilter(func(id string) *abci.ResponseQuery {
			require.Equal(t, "testid", id)
			return &abci.ResponseQuery{Code: uint32(4)}
		})
	}

	suite := NewBaseAppSuite(t, addrPeerFilterOpt, idPeerFilterOpt)

	addrQuery := abci.RequestQuery{
		Path: "/p2p/filter/addr/1.1.1.1:8000",
	}
	res, err := suite.baseApp.Query(context.TODO(), &addrQuery)
	require.NoError(t, err)
	require.Equal(t, uint32(3), res.Code)

	idQuery := abci.RequestQuery{
		Path: "/p2p/filter/id/testid",
	}
	res, err = suite.baseApp.Query(context.TODO(), &idQuery)
	require.NoError(t, err)
	require.Equal(t, uint32(4), res.Code)
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
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})
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
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})
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

func TestABCI_CheckTx(t *testing.T) {
	// This ante handler reads the key and checks that the value matches the
	// current counter. This ensures changes to the KVStore persist across
	// successive CheckTx runs.
	counterKey := []byte("counter-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, counterKey)) }
	suite := NewBaseAppSuite(t, anteOpt)

	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, counterKey})

	nTxs := int64(5)
	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	for i := int64(0); i < nTxs; i++ {
		tx := newTxCounter(t, suite.txConfig, i, 0) // no messages
		txBytes, err := suite.txConfig.TxEncoder()(tx)
		require.NoError(t, err)

		r, err := suite.baseApp.CheckTx(&abci.RequestCheckTx{Tx: txBytes})
		require.NoError(t, err)
		require.True(t, r.IsOK(), fmt.Sprintf("%v", r))
		require.Empty(t, r.GetEvents())
	}

	checkStateStore := getCheckStateCtx(suite.baseApp).KVStore(capKey1)
	storedCounter := getIntFromStore(t, checkStateStore, counterKey)

	// ensure AnteHandler ran
	require.Equal(t, nTxs, storedCounter)

	// if a block is committed, CheckTx state should be reset
	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   []byte("hash"),
	})
	require.NoError(t, err)

	require.NotNil(t, getCheckStateCtx(suite.baseApp).BlockGasMeter(), "block gas meter should have been set to checkState")
	require.NotEmpty(t, getCheckStateCtx(suite.baseApp).HeaderHash())

	_, err = suite.baseApp.Commit()
	require.NoError(t, err)

	checkStateStore = getCheckStateCtx(suite.baseApp).KVStore(capKey1)
	storedBytes := checkStateStore.Get(counterKey)
	require.Nil(t, storedBytes)
}

func TestABCI_FinalizeBlock_DeliverTx(t *testing.T) {
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }
	suite := NewBaseAppSuite(t, anteOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	nBlocks := 3
	txPerHeight := 5

	for blockN := 0; blockN < nBlocks; blockN++ {

		txs := [][]byte{}
		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(t, suite.txConfig, counter, counter)

			txBytes, err := suite.txConfig.TxEncoder()(tx)
			require.NoError(t, err)

			txs = append(txs, txBytes)
		}

		res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: int64(blockN) + 1,
			Txs:    txs,
		})
		require.NoError(t, err)

		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			require.True(t, res.TxResults[i].IsOK(), fmt.Sprintf("%v", res))

			events := res.TxResults[i].GetEvents()
			require.Len(t, events, 3, "should contain ante handler, message type and counter events respectively")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent("ante_handler", counter).ToABCIEvents(), map[string]struct{}{})[0], events[0], "ante handler event")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent(sdk.EventTypeMessage, counter).ToABCIEvents(), map[string]struct{}{})[0].Attributes[0], events[2].Attributes[0], "msg handler update counter event")
		}

		_, err = suite.baseApp.Commit()
		require.NoError(t, err)

	}
}

func TestABCI_FinalizeBlock_MultiMsg(t *testing.T) {
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }
	suite := NewBaseAppSuite(t, anteOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	deliverKey2 := []byte("deliver-key2")
	baseapptestutil.RegisterCounter2Server(suite.baseApp.MsgServiceRouter(), Counter2ServerImpl{t, capKey1, deliverKey2})

	// run a multi-msg tx
	// with all msgs the same route
	tx := newTxCounter(t, suite.txConfig, 0, 0, 1, 2)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{txBytes},
	})
	require.NoError(t, err)

	store := getFinalizeBlockStateCtx(suite.baseApp).KVStore(capKey1)

	// tx counter only incremented once
	txCounter := getIntFromStore(t, store, anteKey)
	require.Equal(t, int64(1), txCounter)

	// msg counter incremented three times
	msgCounter := getIntFromStore(t, store, deliverKey)
	require.Equal(t, int64(3), msgCounter)

	// replace the second message with a Counter2
	tx = newTxCounter(t, suite.txConfig, 1, 3)

	builder := suite.txConfig.NewTxBuilder()
	msgs := tx.GetMsgs()
	_, _, addr := testdata.KeyTestPubAddr()
	msgs = append(msgs, &baseapptestutil.MsgCounter2{Counter: 0, Signer: addr.String()})
	msgs = append(msgs, &baseapptestutil.MsgCounter2{Counter: 1, Signer: addr.String()})

	err = builder.SetMsgs(msgs...)
	require.NoError(t, err)

	builder.SetMemo(tx.GetMemo())
	setTxSignature(t, builder, 0)

	txBytes, err = suite.txConfig.TxEncoder()(builder.GetTx())
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{txBytes},
	})
	require.NoError(t, err)

	store = getFinalizeBlockStateCtx(suite.baseApp).KVStore(capKey1)

	// tx counter only incremented once
	txCounter = getIntFromStore(t, store, anteKey)
	require.Equal(t, int64(2), txCounter)

	// original counter increments by one
	// new counter increments by two
	msgCounter = getIntFromStore(t, store, deliverKey)
	require.Equal(t, int64(4), msgCounter)

	msgCounter2 := getIntFromStore(t, store, deliverKey2)
	require.Equal(t, int64(2), msgCounter2)
}

func TestABCI_Query_SimulateTx(t *testing.T) {
	gasConsumed := uint64(5)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(storetypes.NewGasMeter(gasConsumed))
			return
		})
	}
	suite := NewBaseAppSuite(t, anteOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{gasConsumed})

	nBlocks := 3
	for blockN := 0; blockN < nBlocks; blockN++ {
		count := int64(blockN + 1)

		tx := newTxCounter(t, suite.txConfig, count, count)

		txBytes, err := suite.txConfig.TxEncoder()(tx)
		require.Nil(t, err)

		// simulate a message, check gas reported
		gInfo, result, err := suite.baseApp.Simulate(txBytes)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, gasConsumed, gInfo.GasUsed)

		// simulate again, same result
		gInfo, result, err = suite.baseApp.Simulate(txBytes)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, gasConsumed, gInfo.GasUsed)

		// simulate by calling Query with encoded tx
		query := abci.RequestQuery{
			Path: "/app/simulate",
			Data: txBytes,
		}
		queryResult, err := suite.baseApp.Query(context.TODO(), &query)
		require.NoError(t, err)
		require.True(t, queryResult.IsOK(), queryResult.Log)

		var simRes sdk.SimulationResponse
		require.NoError(t, jsonpb.Unmarshal(strings.NewReader(string(queryResult.Value)), &simRes))

		require.Equal(t, gInfo, simRes.GasInfo)
		require.Equal(t, result.Log, simRes.Result.Log)
		require.Equal(t, result.Events, simRes.Result.Events)
		require.True(t, bytes.Equal(result.Data, simRes.Result.Data))

		_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: count})
		require.NoError(t, err)
		_, err = suite.baseApp.Commit()
		require.NoError(t, err)
	}
}

func TestABCI_InvalidTransaction(t *testing.T) {
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			return
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)
	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err)
	// malformed transaction bytes
	{
		bz := []byte("example vote extension")
		result, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: 1,
			Txs:    [][]byte{bz},
		})

		require.EqualValues(t, sdkerrors.ErrTxDecode.Codespace(), result.TxResults[0].Codespace, err)
		require.EqualValues(t, sdkerrors.ErrTxDecode.ABCICode(), result.TxResults[0].Code, err)
		require.EqualValues(t, 0, result.TxResults[0].GasUsed, err)
		require.EqualValues(t, 0, result.TxResults[0].GasWanted, err)
	}
	// transaction with no messages
	{
		emptyTx := suite.txConfig.NewTxBuilder().GetTx()
		bz, err := suite.txConfig.TxEncoder()(emptyTx)
		require.NoError(t, err)
		result, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: 1,
			Txs:    [][]byte{bz},
		})
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.Codespace(), result.TxResults[0].Codespace, err)
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.ABCICode(), result.TxResults[0].Code, err)
	}

	// transaction where ValidateBasic fails
	{
		testCases := []struct {
			tx   signing.Tx
			fail bool
		}{
			{newTxCounter(t, suite.txConfig, 0, 0), false},
			{newTxCounter(t, suite.txConfig, -1, 0), false},
			{newTxCounter(t, suite.txConfig, 100, 100), false},
			{newTxCounter(t, suite.txConfig, 100, 5, 4, 3, 2, 1), false},

			{newTxCounter(t, suite.txConfig, 0, -1), true},
			{newTxCounter(t, suite.txConfig, 0, 1, -2), true},
			{newTxCounter(t, suite.txConfig, 0, 1, 2, -10, 5), true},
		}

		for _, testCase := range testCases {
			tx := testCase.tx
			_, result, err := suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), tx)

			if testCase.fail {
				require.Error(t, err)

				space, code, _ := errorsmod.ABCIInfo(err, false)
				require.EqualValues(t, sdkerrors.ErrInvalidSequence.Codespace(), space, err)
				require.EqualValues(t, sdkerrors.ErrInvalidSequence.ABCICode(), code, err)
			} else {
				require.NotNil(t, result)
			}
		}
	}

	// transaction with no known route
	{
		txBuilder := suite.txConfig.NewTxBuilder()
		_, _, addr := testdata.KeyTestPubAddr()
		err = txBuilder.SetMsgs(&baseapptestutil.MsgCounter2{Signer: addr.String()})
		require.NoError(t, err)
		setTxSignature(t, txBuilder, 0)
		unknownRouteTx := txBuilder.GetTx()

		_, result, err := suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ := errorsmod.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)

		txBuilder = suite.txConfig.NewTxBuilder()
		err = txBuilder.SetMsgs(
			&baseapptestutil.MsgCounter{Signer: addr.String()},
			&baseapptestutil.MsgCounter2{Signer: addr.String()},
		)
		require.NoError(t, err)
		setTxSignature(t, txBuilder, 0)
		unknownRouteTx = txBuilder.GetTx()

		_, result, err = suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ = errorsmod.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)
	}

	// Transaction with an unregistered message
	{
		txBuilder := suite.txConfig.NewTxBuilder()
		err = txBuilder.SetMsgs(&testdata.MsgCreateDog{})
		require.NoError(t, err)
		tx := txBuilder.GetTx()

		_, _, err := suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), tx)
		space, code, _ := errorsmod.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrTxDecode.ABCICode(), code)
		require.EqualValues(t, sdkerrors.ErrTxDecode.Codespace(), space)
	}
}

func TestABCI_TxGasLimits(t *testing.T) {
	gasGranted := uint64(10)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(storetypes.NewGasMeter(gasGranted))

			// AnteHandlers must have their own defer/recover in order for the BaseApp
			// to know how much gas was used! This is because the GasMeter is created in
			// the AnteHandler, but if it panics the context won't be set properly in
			// runTx's recover call.
			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case storetypes.ErrorOutOfGas:
						err = errorsmod.Wrapf(sdkerrors.ErrOutOfGas, "out of gas in location: %v", rType.Descriptor)
					default:
						panic(r)
					}
				}
			}()

			count, _ := parseTxMemo(t, tx)
			newCtx.GasMeter().ConsumeGas(uint64(count), "counter-ante")

			return newCtx, nil
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err)

	_, err = suite.baseApp.Commit()
	require.NoError(t, err)

	testCases := []struct {
		tx      signing.Tx
		gasUsed int64
		fail    bool
	}{
		{newTxCounter(t, suite.txConfig, 0, 0), 0, false},
		{newTxCounter(t, suite.txConfig, 1, 1), 2, false},
		{newTxCounter(t, suite.txConfig, 9, 1), 10, false},
		{newTxCounter(t, suite.txConfig, 1, 9), 10, false},
		{newTxCounter(t, suite.txConfig, 10, 0), 10, false},

		{newTxCounter(t, suite.txConfig, 9, 2), 11, true},
		{newTxCounter(t, suite.txConfig, 2, 9), 11, true},
		// {newTxCounter(t, suite.txConfig, 9, 1, 1), 11, true},
		// {newTxCounter(t, suite.txConfig, 1, 8, 1, 1), 11, true},
		//  {newTxCounter(t, suite.txConfig, 11, 0), 11, true},
		//  {newTxCounter(t, suite.txConfig, 0, 11), 11, true},
		//  {newTxCounter(t, suite.txConfig, 0, 5, 11), 16, true},
	}

	txs := [][]byte{}
	for _, tc := range testCases {
		tx := tc.tx
		bz, err := suite.txConfig.TxEncoder()(tx)
		require.NoError(t, err)
		txs = append(txs, bz)
	}

	// Deliver the txs
	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 2,
		Txs:    txs,
	})

	require.NoError(t, err)

	for i, tc := range testCases {

		result := res.TxResults[i]

		require.Equal(t, tc.gasUsed, result.GasUsed, fmt.Sprintf("tc #%d; gas: %v, result: %v, err: %s", i, result.GasUsed, result, err))

		// check for out of gas
		if !tc.fail {
			require.NotNil(t, result, fmt.Sprintf("%d: %v, %v", i, tc, err))
		} else {
			require.EqualValues(t, sdkerrors.ErrOutOfGas.Codespace(), result.Codespace, err)
			require.EqualValues(t, sdkerrors.ErrOutOfGas.ABCICode(), result.Code, err)
		}
	}
}

func TestABCI_MaxBlockGasLimits(t *testing.T) {
	gasGranted := uint64(10)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(storetypes.NewGasMeter(gasGranted))

			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case storetypes.ErrorOutOfGas:
						err = errorsmod.Wrapf(sdkerrors.ErrOutOfGas, "out of gas in location: %v", rType.Descriptor)
					default:
						panic(r)
					}
				}
			}()

			count, _ := parseTxMemo(t, tx)
			newCtx.GasMeter().ConsumeGas(uint64(count), "counter-ante")

			return
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{
				MaxGas: 100,
			},
		},
	})
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)

	testCases := []struct {
		tx                signing.Tx
		numDelivers       int
		gasUsedPerDeliver uint64
		fail              bool
		failAfterDeliver  int
	}{
		{newTxCounter(t, suite.txConfig, 0, 0), 0, 0, false, 0},
		{newTxCounter(t, suite.txConfig, 9, 1), 2, 10, false, 0},
		{newTxCounter(t, suite.txConfig, 10, 0), 3, 10, false, 0},
		{newTxCounter(t, suite.txConfig, 10, 0), 10, 10, false, 0},
		{newTxCounter(t, suite.txConfig, 2, 7), 11, 9, false, 0},
		// {newTxCounter(t, suite.txConfig, 10, 0), 10, 10, false, 0}, // hit the limit but pass

		// {newTxCounter(t, suite.txConfig, 10, 0), 11, 10, true, 10},
		// {newTxCounter(t, suite.txConfig, 10, 0), 15, 10, true, 10},
		// {newTxCounter(t, suite.txConfig, 9, 0), 12, 9, true, 11}, // fly past the limit
	}

	for i, tc := range testCases {
		tx := tc.tx

		// reset block gas
		_, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: suite.baseApp.LastBlockHeight() + 1})
		require.NoError(t, err)

		// execute the transaction multiple times
		for j := 0; j < tc.numDelivers; j++ {

			_, result, err := suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), tx)

			ctx := getFinalizeBlockStateCtx(suite.baseApp)

			// check for failed transactions
			if tc.fail && (j+1) > tc.failAfterDeliver {
				require.Error(t, err, fmt.Sprintf("tc #%d; result: %v, err: %s", i, result, err))
				require.Nil(t, tx, fmt.Sprintf("tc #%d; result: %v, err: %s", i, result, err))

				space, code, _ := errorsmod.ABCIInfo(err, false)
				require.EqualValues(t, sdkerrors.ErrOutOfGas.Codespace(), space, err)
				require.EqualValues(t, sdkerrors.ErrOutOfGas.ABCICode(), code, err)
				require.True(t, ctx.BlockGasMeter().IsOutOfGas())
			} else {
				// check gas used and wanted
				blockGasUsed := ctx.BlockGasMeter().GasConsumed()
				expBlockGasUsed := tc.gasUsedPerDeliver * uint64(j+1)
				require.Equal(
					t, expBlockGasUsed, blockGasUsed,
					fmt.Sprintf("%d,%d: %v, %v, %v, %v", i, j, tc, expBlockGasUsed, blockGasUsed, result),
				)

				require.NotNil(t, tx, fmt.Sprintf("tc #%d; currDeliver: %d, result: %v, err: %s", i, j, result, err))
				require.False(t, ctx.BlockGasMeter().IsPastLimit())
			}
		}
	}
}

func TestABCI_GasConsumptionBadTx(t *testing.T) {
	gasWanted := uint64(5)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(storetypes.NewGasMeter(gasWanted))

			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case storetypes.ErrorOutOfGas:
						log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
						err = errorsmod.Wrap(sdkerrors.ErrOutOfGas, log)
					default:
						panic(r)
					}
				}
			}()

			counter, failOnAnte := parseTxMemo(t, tx)
			newCtx.GasMeter().ConsumeGas(uint64(counter), "counter-ante")
			if failOnAnte {
				return newCtx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "ante handler failure")
			}

			return
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{
				MaxGas: 9,
			},
		},
	})
	require.NoError(t, err)

	tx := newTxCounter(t, suite.txConfig, 5, 0)
	tx = setFailOnAnte(t, suite.txConfig, tx, true)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	// require next tx to fail due to black gas limit
	tx = newTxCounter(t, suite.txConfig, 5, 0)
	txBytes2, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.baseApp.LastBlockHeight() + 1,
		Txs:    [][]byte{txBytes, txBytes2},
	})
	require.NoError(t, err)
}

func TestABCI_Query(t *testing.T) {
	key, value := []byte("hello"), []byte("goodbye")
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			store := ctx.KVStore(capKey1)
			store.Set(key, value)
			return
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	// NOTE: "/store/key1" tells us KVStore
	// and the final "/key" says to use the data as the
	// key in the given KVStore ...
	query := abci.RequestQuery{
		Path: "/store/key1/key",
		Data: key,
	}
	tx := newTxCounter(t, suite.txConfig, 0, 0)

	// query is empty before we do anything
	res, err := suite.baseApp.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a CheckTx
	_, resTx, err := suite.baseApp.SimCheck(suite.txConfig.TxEncoder(), tx)
	require.NoError(t, err)
	require.NotNil(t, resTx)

	res, err = suite.baseApp.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, 0, len(res.Value))

	bz, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{bz},
	})
	require.NoError(t, err)

	res, err = suite.baseApp.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, 0, len(res.Value))

	// query returns correct value after Commit
	_, err = suite.baseApp.Commit()
	require.NoError(t, err)

	res, err = suite.baseApp.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, value, res.Value)
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

		tc.bapp.SetParamStore(&paramStore{db: dbm.NewMemDB()})
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

func TestABCI_Proposal_HappyPath(t *testing.T) {
	anteKey := []byte("ante-key")
	pool := mempool.NewSenderNonceMempool()
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	suite := NewBaseAppSuite(t, anteOpt, baseapp.SetMempool(pool))
	baseapptestutil.RegisterKeyValueServer(suite.baseApp.MsgServiceRouter(), MsgKeyValueImpl{})
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	tx := newTxCounter(t, suite.txConfig, 0, 1)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	reqCheckTx := abci.RequestCheckTx{
		Tx:   txBytes,
		Type: abci.CheckTxType_New,
	}
	_, err = suite.baseApp.CheckTx(&reqCheckTx)
	require.NoError(t, err)

	tx2 := newTxCounter(t, suite.txConfig, 1, 1)

	tx2Bytes, err := suite.txConfig.TxEncoder()(tx2)
	require.NoError(t, err)

	err = pool.Insert(sdk.Context{}, tx2)
	require.NoError(t, err)

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1,
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 2, len(resPrepareProposal.Txs))

	reqProposalTxBytes := [2][]byte{
		txBytes,
		tx2Bytes,
	}
	reqProcessProposal := abci.RequestProcessProposal{
		Txs:    reqProposalTxBytes[:],
		Height: reqPrepareProposal.Height,
	}

	resProcessProposal, err := suite.baseApp.ProcessProposal(&reqProcessProposal)
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_ACCEPT, resProcessProposal.Status)

	// the same txs as in PrepareProposal
	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.baseApp.LastBlockHeight() + 1,
		Txs:    reqProposalTxBytes[:],
	})
	require.NoError(t, err)

	require.Equal(t, 0, pool.CountTx())

	require.NotEmpty(t, res.TxResults[0].Events)
	require.True(t, res.TxResults[0].IsOK(), fmt.Sprintf("%v", res))
}

func TestABCI_Proposal_Read_State_PrepareProposal(t *testing.T) {
	someKey := []byte("some-key")

	setInitChainerOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
			ctx.KVStore(capKey1).Set(someKey, []byte("foo"))
			return &abci.ResponseInitChain{}, nil
		})
	}

	prepareOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetPrepareProposal(func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			value := ctx.KVStore(capKey1).Get(someKey)
			// We should be able to access any state written in InitChain
			require.Equal(t, "foo", string(value))
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		})
	}

	suite := NewBaseAppSuite(t, setInitChainerOpt, prepareOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		InitialHeight:   1,
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1, // this value can't be 0
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 0, len(resPrepareProposal.Txs))

	reqProposalTxBytes := [][]byte{}
	reqProcessProposal := abci.RequestProcessProposal{
		Txs:    reqProposalTxBytes,
		Height: reqPrepareProposal.Height,
	}

	resProcessProposal, err := suite.baseApp.ProcessProposal(&reqProcessProposal)
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_ACCEPT, resProcessProposal.Status)
}

func TestABCI_Proposals_WithVE(t *testing.T) {
	someVoteExtension := []byte("some-vote-extension")

	setInitChainerOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
			return &abci.ResponseInitChain{}, nil
		})
	}

	prepareOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetPrepareProposal(func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			// Inject the vote extension to the beginning of the proposal
			txs := make([][]byte, len(req.Txs)+1)
			txs[0] = someVoteExtension
			copy(txs[1:], req.Txs)

			return &abci.ResponsePrepareProposal{Txs: txs}, nil
		})

		bapp.SetProcessProposal(func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
			// Check that the vote extension is still there
			require.Equal(t, someVoteExtension, req.Txs[0])
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		})
	}

	suite := NewBaseAppSuite(t, setInitChainerOpt, prepareOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		InitialHeight:   1,
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)
	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 100000,
		Height:     1, // this value can't be 0
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 1, len(resPrepareProposal.Txs))

	reqProcessProposal := abci.RequestProcessProposal{
		Txs:    resPrepareProposal.Txs,
		Height: reqPrepareProposal.Height,
	}
	resProcessProposal, err := suite.baseApp.ProcessProposal(&reqProcessProposal)
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_ACCEPT, resProcessProposal.Status)

	// Run finalize block and ensure that the vote extension is still there and that
	// the proposal is accepted
	result, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Txs:    resPrepareProposal.Txs,
		Height: reqPrepareProposal.Height,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(result.TxResults))
	require.EqualValues(t, sdkerrors.ErrTxDecode.Codespace(), result.TxResults[0].Codespace, err)
	require.EqualValues(t, sdkerrors.ErrTxDecode.ABCICode(), result.TxResults[0].Code, err)
	require.EqualValues(t, 0, result.TxResults[0].GasUsed, err)
	require.EqualValues(t, 0, result.TxResults[0].GasWanted, err)
}

func TestABCI_PrepareProposal_ReachedMaxBytes(t *testing.T) {
	anteKey := []byte("ante-key")
	pool := mempool.NewSenderNonceMempool()
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	suite := NewBaseAppSuite(t, anteOpt, baseapp.SetMempool(pool))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		tx2 := newTxCounter(t, suite.txConfig, int64(i), int64(i))
		err := pool.Insert(sdk.Context{}, tx2)
		require.NoError(t, err)
	}

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1500,
		Height:     1,
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 8, len(resPrepareProposal.Txs))
}

func TestABCI_PrepareProposal_BadEncoding(t *testing.T) {
	anteKey := []byte("ante-key")
	pool := mempool.NewSenderNonceMempool()
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	suite := NewBaseAppSuite(t, anteOpt, baseapp.SetMempool(pool))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	tx := newTxCounter(t, suite.txConfig, 0, 0)
	err = pool.Insert(sdk.Context{}, tx)
	require.NoError(t, err)

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1,
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 1, len(resPrepareProposal.Txs))
}

func TestABCI_PrepareProposal_OverGasUnderBytes(t *testing.T) {
	pool := mempool.NewSenderNonceMempool()
	suite := NewBaseAppSuite(t, baseapp.SetMempool(pool))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	// set max block gas limit to 99, this will allow 9 txs of 10 gas each.
	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{MaxGas: 99},
		},
	})

	require.NoError(t, err)
	// insert 100 txs, each with a gas limit of 10
	_, _, addr := testdata.KeyTestPubAddr()
	for i := int64(0); i < 100; i++ {
		msg := &baseapptestutil.MsgCounter{Counter: i, FailOnHandler: false, Signer: addr.String()}
		msgs := []sdk.Msg{msg}

		builder := suite.txConfig.NewTxBuilder()
		err = builder.SetMsgs(msgs...)
		require.NoError(t, err)
		builder.SetMemo("counter=" + strconv.FormatInt(i, 10) + "&failOnAnte=false")
		builder.SetGasLimit(10)
		setTxSignature(t, builder, uint64(i))

		err := pool.Insert(sdk.Context{}, builder.GetTx())
		require.NoError(t, err)
	}

	// ensure we only select transactions that fit within the block gas limit
	res, err := suite.baseApp.PrepareProposal(&abci.RequestPrepareProposal{
		MaxTxBytes: 1_000_000, // large enough to ignore restriction
		Height:     1,
	})
	require.NoError(t, err)

	// Should include 9 transactions
	require.Len(t, res.Txs, 9, "invalid number of transactions returned")
}

func TestABCI_PrepareProposal_MaxGas(t *testing.T) {
	pool := mempool.NewSenderNonceMempool()
	suite := NewBaseAppSuite(t, baseapp.SetMempool(pool))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	// set max block gas limit to 100
	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{MaxGas: 100},
		},
	})
	require.NoError(t, err)
	// insert 100 txs, each with a gas limit of 10
	_, _, addr := testdata.KeyTestPubAddr()
	for i := int64(0); i < 100; i++ {
		msg := &baseapptestutil.MsgCounter{Counter: i, FailOnHandler: false, Signer: addr.String()}
		msgs := []sdk.Msg{msg}

		builder := suite.txConfig.NewTxBuilder()
		err = builder.SetMsgs(msgs...)
		require.NoError(t, err)
		builder.SetMemo("counter=" + strconv.FormatInt(i, 10) + "&failOnAnte=false")
		builder.SetGasLimit(10)
		setTxSignature(t, builder, uint64(i))

		err := pool.Insert(sdk.Context{}, builder.GetTx())
		require.NoError(t, err)
	}

	// ensure we only select transactions that fit within the block gas limit
	res, err := suite.baseApp.PrepareProposal(&abci.RequestPrepareProposal{
		MaxTxBytes: 1_000_000, // large enough to ignore restriction
		Height:     1,
	})
	require.NoError(t, err)
	require.Len(t, res.Txs, 10, "invalid number of transactions returned")
}

func TestABCI_PrepareProposal_Failures(t *testing.T) {
	anteKey := []byte("ante-key")
	pool := mempool.NewSenderNonceMempool()
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	suite := NewBaseAppSuite(t, anteOpt, baseapp.SetMempool(pool))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	tx := newTxCounter(t, suite.txConfig, 0, 0)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	reqCheckTx := abci.RequestCheckTx{
		Tx:   txBytes,
		Type: abci.CheckTxType_New,
	}
	checkTxRes, err := suite.baseApp.CheckTx(&reqCheckTx)
	require.NoError(t, err)
	require.True(t, checkTxRes.IsOK())

	failTx := newTxCounter(t, suite.txConfig, 1, 1)
	failTx = setFailOnAnte(t, suite.txConfig, failTx, true)

	err = pool.Insert(sdk.Context{}, failTx)
	require.NoError(t, err)
	require.Equal(t, 2, pool.CountTx())

	req := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1,
	}
	res, err := suite.baseApp.PrepareProposal(&req)
	require.NoError(t, err)
	require.Equal(t, 1, len(res.Txs))
}

func TestABCI_PrepareProposal_PanicRecovery(t *testing.T) {
	prepareOpt := func(app *baseapp.BaseApp) {
		app.SetPrepareProposal(func(ctx sdk.Context, rpp *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			panic(errors.New("test"))
		})
	}
	suite := NewBaseAppSuite(t, prepareOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	req := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1,
	}

	require.NotPanics(t, func() {
		res, err := suite.baseApp.PrepareProposal(&req)
		require.NoError(t, err)
		require.Equal(t, req.Txs, res.Txs)
	})
}

func TestABCI_PrepareProposal_VoteExtensions(t *testing.T) {
	// set up mocks
	ctrl := gomock.NewController(t)
	valStore := mock.NewMockValidatorStore(ctrl)
	privkey := secp256k1.GenPrivKey()
	pubkey := privkey.PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	tmPk := cmtprotocrypto.PublicKey{
		Sum: &cmtprotocrypto.PublicKey_Secp256K1{
			Secp256K1: pubkey.Bytes(),
		},
	}

	consAddr := sdk.ConsAddress(addr.String())
	valStore.EXPECT().GetPubKeyByConsAddr(gomock.Any(), consAddr.Bytes()).Return(tmPk, nil)

	// set up baseapp
	prepareOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetPrepareProposal(func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			cp := ctx.ConsensusParams()
			extsEnabled := cp.Abci != nil && req.Height >= cp.Abci.VoteExtensionsEnableHeight && cp.Abci.VoteExtensionsEnableHeight != 0
			if extsEnabled {
				err := baseapp.ValidateVoteExtensions(ctx, valStore, req.Height, bapp.ChainID(), req.LocalLastCommit)
				if err != nil {
					return nil, err
				}

				req.Txs = append(req.Txs, []byte("some-tx-that-does-something-from-votes"))

			}
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		})
	}

	suite := NewBaseAppSuite(t, prepareOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		InitialHeight: 1,
		ConsensusParams: &cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 2,
			},
		},
	})
	require.NoError(t, err)

	// first test without vote extensions, no new txs should be added
	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1, // this value can't be 0
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 0, len(resPrepareProposal.Txs))

	// now we try with vote extensions, a new tx should show up
	marshalDelimitedFn := func(msg proto.Message) ([]byte, error) {
		var buf bytes.Buffer
		if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	ext := []byte("something")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2, // the vote extension was signed in the previous height
		Round:     int64(0),
		ChainId:   suite.baseApp.ChainID(),
	}

	bz, err := marshalDelimitedFn(&cve)
	require.NoError(t, err)

	extSig, err := privkey.Sign(bz)
	require.NoError(t, err)

	reqPrepareProposal = abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     3, // this value can't be 0
		LocalLastCommit: abci.ExtendedCommitInfo{
			Round: 0,
			Votes: []abci.ExtendedVoteInfo{
				{
					Validator: abci.Validator{
						Address: consAddr.Bytes(),
						Power:   666,
					},
					VoteExtension:      ext,
					ExtensionSignature: extSig,
					BlockIdFlag:        cmtproto.BlockIDFlagCommit,
				},
			},
		},
	}
	resPrepareProposal, err = suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 1, len(resPrepareProposal.Txs))

	// now vote extensions but our sole voter doesn't reach majority
	reqPrepareProposal = abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     3, // this value can't be 0
		LocalLastCommit: abci.ExtendedCommitInfo{
			Round: 0,
			Votes: []abci.ExtendedVoteInfo{
				{
					Validator: abci.Validator{
						Address: consAddr.Bytes(),
						Power:   666,
					},
					VoteExtension:      ext,
					ExtensionSignature: extSig,
					BlockIdFlag:        cmtproto.BlockIDFlagNil, // This will ignore the vote extension
				},
			},
		},
	}
	resPrepareProposal, err = suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 0, len(resPrepareProposal.Txs))
}

func TestABCI_ProcessProposal_PanicRecovery(t *testing.T) {
	processOpt := func(app *baseapp.BaseApp) {
		app.SetProcessProposal(func(ctx sdk.Context, rpp *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
			panic(errors.New("test"))
		})
	}
	suite := NewBaseAppSuite(t, processOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	require.NotPanics(t, func() {
		res, err := suite.baseApp.ProcessProposal(&abci.RequestProcessProposal{Height: 1})
		require.NoError(t, err)
		require.Equal(t, res.Status, abci.ResponseProcessProposal_REJECT)
	})
}

// TestABCI_Proposal_Reset_State ensures that state is reset between runs of
// PrepareProposal and ProcessProposal in case they are called multiple times.
// This is only valid for heights > 1, given that on height 1 we always set the
// state to be deliverState.
func TestABCI_Proposal_Reset_State_Between_Calls(t *testing.T) {
	someKey := []byte("some-key")

	prepareOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetPrepareProposal(func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			// This key should not exist given that we reset the state on every call.
			require.False(t, ctx.KVStore(capKey1).Has(someKey))
			ctx.KVStore(capKey1).Set(someKey, someKey)
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		})
	}

	processOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetProcessProposal(func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
			// This key should not exist given that we reset the state on every call.
			require.False(t, ctx.KVStore(capKey1).Has(someKey))
			ctx.KVStore(capKey1).Set(someKey, someKey)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		})
	}

	suite := NewBaseAppSuite(t, prepareOpt, processOpt)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     2, // this value can't be 0
	}

	// Let's pretend something happened and PrepareProposal gets called many
	// times, this must be safe to do.
	for i := 0; i < 5; i++ {
		resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
		require.NoError(t, err)
		require.Equal(t, 0, len(resPrepareProposal.Txs))
	}

	reqProposalTxBytes := [][]byte{}
	reqProcessProposal := abci.RequestProcessProposal{
		Txs:    reqProposalTxBytes,
		Height: 2,
	}

	// Let's pretend something happened and ProcessProposal gets called many
	// times, this must be safe to do.
	for i := 0; i < 5; i++ {
		resProcessProposal, err := suite.baseApp.ProcessProposal(&reqProcessProposal)
		require.NoError(t, err)
		require.Equal(t, abci.ResponseProcessProposal_ACCEPT, resProcessProposal.Status)
	}
}

func TestABCI_HaltChain(t *testing.T) {
	testCases := []struct {
		name        string
		haltHeight  uint64
		haltTime    uint64
		blockHeight int64
		blockTime   int64
		expHalt     bool
	}{
		{"default", 0, 0, 10, 0, false},
		{"halt-height-edge", 10, 0, 10, 0, false},
		{"halt-height", 10, 0, 11, 0, true},
		{"halt-time-edge", 0, 10, 1, 10, false},
		{"halt-time", 0, 10, 1, 11, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := NewBaseAppSuite(t, baseapp.SetHaltHeight(tc.haltHeight), baseapp.SetHaltTime(tc.haltTime))
			_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
				ConsensusParams: &cmtproto.ConsensusParams{},
				InitialHeight:   tc.blockHeight,
			})
			require.NoError(t, err)
			app := suite.baseApp
			_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
				Height: tc.blockHeight,
				Time:   time.Unix(tc.blockTime, 0),
			})
			if !tc.expHalt {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.True(t, strings.HasPrefix(err.Error(), "halt per configuration"))
			}
		})
	}
}

func TestBaseApp_PreBlocker(t *testing.T) {
	db := dbm.NewMemDB()
	name := t.Name()
	logger := log.NewTestLogger(t)

	app := baseapp.NewBaseApp(name, logger, db, nil)
	_, err := app.InitChain(&abci.RequestInitChain{})
	require.NoError(t, err)

	wasHookCalled := false
	app.SetPreBlocker(func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		wasHookCalled = true
		return &sdk.ResponsePreBlock{
			ConsensusParamsChanged: true,
		}, nil
	})
	app.Seal()

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	require.Equal(t, true, wasHookCalled)

	// Now try erroring
	app = baseapp.NewBaseApp(name, logger, db, nil)
	_, err = app.InitChain(&abci.RequestInitChain{})
	require.NoError(t, err)

	app.SetPreBlocker(func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		return nil, errors.New("some error")
	})
	app.Seal()

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.Error(t, err)
}

// TestBaseApp_VoteExtensions tests vote extensions using a price as an example.
func TestBaseApp_VoteExtensions(t *testing.T) {
	baseappOpts := func(app *baseapp.BaseApp) {
		app.SetExtendVoteHandler(func(sdk.Context, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
			// here we would have a process to get the price from an external source
			price := 10000000 + rand.Int63n(1000000)
			ve := make([]byte, 8)
			binary.BigEndian.PutUint64(ve, uint64(price))
			return &abci.ResponseExtendVote{VoteExtension: ve}, nil
		})

		app.SetVerifyVoteExtensionHandler(func(_ sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
			vePrice := binary.BigEndian.Uint64(req.VoteExtension)
			// here we would do some price validation, must not be 0 and not too high
			if vePrice > 11000000 || vePrice == 0 {
				// usually application should always return ACCEPT unless they really want to discard the entire vote
				return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
			}

			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
		})

		app.SetPrepareProposal(func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			txs := [][]byte{}

			// add all VE as txs (in a real scenario we would need to check signatures too)
			for _, v := range req.LocalLastCommit.Votes {
				if len(v.VoteExtension) == 8 {
					// pretend this is a way to check if the VE is valid
					if binary.BigEndian.Uint64(v.VoteExtension) < 11000000 && binary.BigEndian.Uint64(v.VoteExtension) > 0 {
						txs = append(txs, v.VoteExtension)
					}
				}
			}

			return &abci.ResponsePrepareProposal{Txs: txs}, nil
		})

		app.SetProcessProposal(func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
			// here we check if the proposal is valid, mainly if the vote extensions appended to the txs are valid
			for _, v := range req.Txs {
				// pretend this is a way to check if the tx is actually a VE
				if len(v) == 8 {
					// pretend this is a way to check if the VE is valid
					if binary.BigEndian.Uint64(v) > 11000000 || binary.BigEndian.Uint64(v) == 0 {
						return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
					}
				}
			}

			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		})

		app.SetPreBlocker(func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
			count := uint64(0)
			pricesSum := uint64(0)
			for _, v := range req.Txs {
				// pretend this is a way to check if the tx is actually a VE
				if len(v) == 8 {
					count++
					pricesSum += binary.BigEndian.Uint64(v)
				}
			}

			if count > 0 {
				// we process the average price and store it in the context to make it available for FinalizeBlock
				avgPrice := pricesSum / count
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, avgPrice)
				ctx.KVStore(capKey1).Set([]byte("avgPrice"), buf)
			}

			return &sdk.ResponsePreBlock{
				ConsensusParamsChanged: true,
			}, nil
		})
	}

	suite := NewBaseAppSuite(t, baseappOpts)

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 1,
			},
		},
	})
	require.NoError(t, err)

	allVEs := [][]byte{}
	// simulate getting 10 vote extensions from 10 validators
	for i := 0; i < 10; i++ {
		ve, err := suite.baseApp.ExtendVote(context.TODO(), &abci.RequestExtendVote{Height: 1})
		require.NoError(t, err)
		allVEs = append(allVEs, ve.VoteExtension)
	}

	// add a couple of invalid vote extensions (in what regards to the check we are doing in VerifyVoteExtension/ProcessProposal)
	// add a 0 price
	ve := make([]byte, 8)
	binary.BigEndian.PutUint64(ve, uint64(0))
	allVEs = append(allVEs, ve)

	// add a price too high
	ve = make([]byte, 8)
	binary.BigEndian.PutUint64(ve, uint64(13000000))
	allVEs = append(allVEs, ve)

	// verify all votes, only 10 should be accepted
	successful := 0
	for _, v := range allVEs {
		res, err := suite.baseApp.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{
			Height:        1,
			VoteExtension: v,
		})
		require.NoError(t, err)
		if res.Status == abci.ResponseVerifyVoteExtension_ACCEPT {
			successful++
		}
	}
	require.Equal(t, 10, successful)

	prepPropReq := &abci.RequestPrepareProposal{
		Height: 1,
		LocalLastCommit: abci.ExtendedCommitInfo{
			Round: 0,
			Votes: []abci.ExtendedVoteInfo{},
		},
	}

	// add all VEs to the local last commit
	for _, ve := range allVEs {
		prepPropReq.LocalLastCommit.Votes = append(prepPropReq.LocalLastCommit.Votes, abci.ExtendedVoteInfo{VoteExtension: ve})
	}

	resp, err := suite.baseApp.PrepareProposal(prepPropReq)
	require.NoError(t, err)
	require.Len(t, resp.Txs, 10)

	procPropRes, err := suite.baseApp.ProcessProposal(&abci.RequestProcessProposal{Height: 1, Txs: resp.Txs})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_ACCEPT, procPropRes.Status)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: resp.Txs})
	require.NoError(t, err)

	// Check if the average price was available in FinalizeBlock's context
	avgPrice := getFinalizeBlockStateCtx(suite.baseApp).KVStore(capKey1).Get([]byte("avgPrice"))
	require.NotNil(t, avgPrice)
	require.GreaterOrEqual(t, binary.BigEndian.Uint64(avgPrice), uint64(10000000))
	require.Less(t, binary.BigEndian.Uint64(avgPrice), uint64(11000000))

	_, err = suite.baseApp.Commit()
	require.NoError(t, err)

	// check if avgPrice was committed
	committedAvgPrice := suite.baseApp.NewContext(true).KVStore(capKey1).Get([]byte("avgPrice"))
	require.Equal(t, avgPrice, committedAvgPrice)
}

func TestOptimisticExecution(t *testing.T) {
	suite := NewBaseAppSuite(t, baseapp.SetOptimisticExecution())

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	// run 50 blocks
	for i := 0; i < 50; i++ {
		tx := newTxCounter(t, suite.txConfig, 0, 1)
		txBytes, err := suite.txConfig.TxEncoder()(tx)
		require.NoError(t, err)

		reqProcProp := abci.RequestProcessProposal{
			Txs:    [][]byte{txBytes},
			Height: suite.baseApp.LastBlockHeight() + 1,
			Hash:   []byte("some-hash" + strconv.FormatInt(suite.baseApp.LastBlockHeight()+1, 10)),
		}

		respProcProp, err := suite.baseApp.ProcessProposal(&reqProcProp)
		require.Equal(t, abci.ResponseProcessProposal_ACCEPT, respProcProp.Status)
		require.NoError(t, err)

		reqFinalizeBlock := abci.RequestFinalizeBlock{
			Height: reqProcProp.Height,
			Txs:    reqProcProp.Txs,
			Hash:   reqProcProp.Hash,
		}

		respFinalizeBlock, err := suite.baseApp.FinalizeBlock(&reqFinalizeBlock)
		require.NoError(t, err)
		require.Len(t, respFinalizeBlock.TxResults, 1)

		_, err = suite.baseApp.Commit()
		require.NoError(t, err)
	}

	require.Equal(t, int64(50), suite.baseApp.LastBlockHeight())
}
