package integration

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	cmtabcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
)

const appName = "integration-app"

// App is a test application that can be used to test the integration of modules.
type App struct {
	*baseapp.BaseApp

	logger        log.Logger
	moduleManager module.Manager
	queryHelper   *baseapp.QueryServiceTestHelper

	cdc codec.Codec
}

// NewIntegrationApp creates an application for testing purposes. This application
// is able to route messages to their respective handlers.
func NewIntegrationApp(
	logger log.Logger,
	keys map[string]*storetypes.KVStoreKey,
	appCodec codec.Codec,
	modules map[string]appmodule.AppModule,
) *App {
	db := dbm.NewMemDB()

	interfaceRegistry := appCodec.InterfaceRegistry()
	moduleManager := module.NewManagerFromMap(modules)
	basicModuleManager := module.NewBasicManagerFromManager(moduleManager, nil)
	basicModuleManager.RegisterInterfaces(interfaceRegistry)

	txConfig := authtx.NewTxConfig(codec.NewProtoCodec(interfaceRegistry), authtx.DefaultSignModes)
	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseapp.SetChainID(appName))
	bApp.MountKVStores(keys)

	bApp.SetInitChainer(func(ctx sdk.Context, _ *cmtabcitypes.RequestInitChain) (*cmtabcitypes.ResponseInitChain, error) {
		for _, mod := range modules {
			if m, ok := mod.(module.HasGenesis); ok {
				m.InitGenesis(ctx, appCodec, m.DefaultGenesis(appCodec))
			}
		}

		return &cmtabcitypes.ResponseInitChain{}, nil
	})

	bApp.SetBeginBlocker(moduleManager.BeginBlock)
	bApp.SetEndBlocker(moduleManager.EndBlock)

	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetMsgServiceRouter(router)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	if keys[consensusparamtypes.StoreKey] != nil {
		// set baseApp param store
		consensusParamsKeeper := consensusparamkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]), authtypes.NewModuleAddress("gov").String(), runtime.EventService{})
		bApp.SetParamStore(consensusParamsKeeper.ParamsStore)

		if err := bApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("failed to load application version from store: %w", err))
		}

		if _, err := bApp.InitChain(&cmtabcitypes.RequestInitChain{ChainId: appName, ConsensusParams: simtestutil.DefaultConsensusParams}); err != nil {
			panic(fmt.Errorf("failed to initialize application: %w", err))
		}
	} else {
		if err := bApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("failed to load application version from store: %w", err))
		}

		if _, err := bApp.InitChain(&cmtabcitypes.RequestInitChain{ChainId: appName}); err != nil {
			panic(fmt.Errorf("failed to initialize application: %w", err))
		}
	}

	ctx := bApp.NewContext(true).WithBlockHeader(cmtproto.Header{ChainID: appName}).WithIsCheckTx(true)

	return &App{
		BaseApp:       bApp,
		logger:        logger,
		moduleManager: *moduleManager,
		queryHelper:   baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry),
		cdc:           appCodec,
	}
}

// RunMsg provides the ability to run a message and return the response.
// In order to run a message, the application must have a handler for it.
// These handlers are registered on the application message service router.
// The result of the message execution is returned as an Any type.
// That any type can be unmarshaled to the expected response type.
// If the message execution fails, an error is returned.
func (app *App) RunMsg(msg sdk.Msg, option ...Option) (*cmtabcitypes.ResponseFinalizeBlock, error) {
	// set options
	cfg := &Config{}
	for _, opt := range option {
		opt(cfg)
	}

	txConfig := tx.NewTxConfig(app.cdc, tx.DefaultSignModes)
	app.SetTxDecoder(txConfig.TxDecoder())
	app.SetTxEncoder(txConfig.TxEncoder())

	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		[]sdk.Msg{msg},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		simtestutil.DefaultGenTxGas,
		app.ChainID(),
		[]uint64{0},
		[]uint64{0},
		secp256k1.GenPrivKey(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed tx: %w", err)
	}

	bz, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tx: %w", err)
	}

	if cfg.AutomaticCommit {
		defer app.Commit()
	}

	height := app.LastBlockHeight() + 1

	if cfg.AutomaticProcessProposal {
		req := &cmtabcitypes.RequestProcessProposal{
			Height: height,
			Txs:    [][]byte{bz},
		}

		if cfg.FinalizeBlockRequest != nil {
			req.ProposedLastCommit = cfg.FinalizeBlockRequest.DecidedLastCommit
			req.ProposerAddress = cfg.FinalizeBlockRequest.ProposerAddress
			req.Misbehavior = cfg.FinalizeBlockRequest.Misbehavior
			req.Time = cfg.FinalizeBlockRequest.Time
		}

		_, err = app.ProcessProposal(req)

		if err != nil {
			return nil, fmt.Errorf("failed to run process proposal: %w", err)
		}
	}

	req := &cmtabcitypes.RequestFinalizeBlock{
		Height: height,
		Txs:    [][]byte{bz},
	}

	if cfg.FinalizeBlockRequest != nil {
		req.DecidedLastCommit = cfg.FinalizeBlockRequest.DecidedLastCommit
		req.ProposerAddress = cfg.FinalizeBlockRequest.ProposerAddress
		req.Misbehavior = cfg.FinalizeBlockRequest.Misbehavior
		req.Time = cfg.FinalizeBlockRequest.Time
	}
	res, err := app.FinalizeBlock(req)
	if err != nil {
		return nil, fmt.Errorf("failed to run finalize block: %w", err)
	}

	if res.TxResults[0].Code != 0 {
		return res, fmt.Errorf("tx returned a non-zero code: %s", res.TxResults[0].Log)
	}

	return res, nil
}

// Context returns the application context. It can be unwrapped to a sdk.Context,
// with the sdk.UnwrapSDKContext function.
func (app *App) Context() context.Context {
	return app.NewContext(true)
}

// QueryHelper returns the application query helper.
// It can be used when registering query services.
func (app *App) QueryHelper() *baseapp.QueryServiceTestHelper {
	return app.queryHelper
}

// CreateMultiStore is a helper for setting up multiple stores for provided modules.
func CreateMultiStore(keys map[string]*storetypes.KVStoreKey, logger log.Logger) storetypes.CommitMultiStore {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())

	for key := range keys {
		cms.MountStoreWithDB(keys[key], storetypes.StoreTypeIAVL, db)
	}

	_ = cms.LoadLatestVersion()
	return cms
}
