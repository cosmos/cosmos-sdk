package integration

import (
	"context"
	"fmt"

	cmtabcitypes "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
)

const appName = "integration-app"

// App is a test application that can be used to test the integration of modules.
type App struct {
	*baseapp.BaseApp

	ctx           sdk.Context
	logger        log.Logger
	moduleManager module.Manager
	// queryHelper   *baseapp.QueryServiceTestHelper
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry
}

// NewIntegrationApp creates an application for testing purposes. This application
// is able to route messages to their respective handlers.
func NewIntegrationApp(
	logger log.Logger,
	keys map[string]*storetypes.KVStoreKey,
	appCodec codec.Codec,
	modules map[string]appmodule.AppModule,
	setupFn func(ctx sdk.Context) error,
) *App {
	db := dbm.NewMemDB()

	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(
		codectypes.InterfaceRegistryOptions{
			ProtoFiles: proto.HybridResolver,
			SigningOptions: signing.Options{
				AddressCodec:          address.NewBech32Codec("cosmos"),
				ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
			},
		},
	)
	if err != nil {
		panic(err)
	}

	moduleManager := module.NewManagerFromMap(modules)
	basicModuleManager := module.NewBasicManagerFromManager(moduleManager, nil)
	basicModuleManager.RegisterInterfaces(interfaceRegistry)

	txConfig := authtx.NewTxConfig(codec.NewProtoCodec(interfaceRegistry), authtx.DefaultSignModes)
	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseapp.SetChainID(appName))
	bApp.MountKVStores(keys)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	bApp.SetInitChainer(func(ctx sdk.Context, _ *cmtabcitypes.RequestInitChain) (*cmtabcitypes.ResponseInitChain, error) {
		for _, mod := range modules {
			if m, ok := mod.(module.HasGenesis); ok {
				m.InitGenesis(ctx, appCodec, m.DefaultGenesis(appCodec))
			}
		}

		return &cmtabcitypes.ResponseInitChain{}, nil
	})

	bApp.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
		return moduleManager.BeginBlock(ctx)
	})
	bApp.SetEndBlocker(func(ctx sdk.Context) (sdk.EndBlock, error) {
		return moduleManager.EndBlock(ctx)
	})

	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetMsgServiceRouter(router)

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

	if err = setupFn(bApp.GetContextForFinalizeBlock([]byte{})); err != nil {
		panic(err)
	}

	_, err = bApp.FinalizeBlock(&cmtabcitypes.RequestFinalizeBlock{Height: 1})
	if err != nil {
		panic(err)
	}

	_, err = bApp.Commit()
	if err != nil {
		panic(err)
	}

	return &App{
		BaseApp:       bApp,
		logger:        logger,
		moduleManager: *moduleManager,
		// queryHelper:   baseapp.NewQueryServerTestHelper(sdk.Context{}, interfaceRegistry), // tODO fix
		interfaceRegistry: interfaceRegistry,
		txConfig:          txConfig,
	}
}

// RunArbitraryCode allows executing arbitrary code on the application state.
// Useful when setting up the application state for testing purposes as it doesn't
// increase the block height.
func (app *App) RunArbitraryCode(fn func(ctx sdk.Context) error) error {
	// Run ProcessProposal to set finalizeBlockState.
	_, err := app.ProcessProposal(&cmtabcitypes.RequestProcessProposal{Height: app.LastBlockHeight() + 1})
	if err != nil {
		return err
	}

	ctx := app.GetContextForFinalizeBlock([]byte{})
	if err = fn(ctx); err != nil {
		return err
	}
	ms := ctx.MultiStore().(storetypes.CacheMultiStore)
	ms.Write()

	_, err = app.Commit()
	return err
}

// RunMsg provides the ability to run a message and return the response.
// In order to run a message, the application must have a handler for it.
// These handlers are registered on the application message service router.
// The result of the message execution is returned as an Any type.
// That any type can be unmarshaled to the expected response type.
// If the message execution fails, an error is returned.
func (app *App) RunMsg(msg sdk.Msg, option ...Option) (*cmtabcitypes.ExecTxResult, error) {
	// set options
	cfg := &Config{}
	for _, opt := range option {
		opt(cfg)
	}

	if cfg.AutomaticCommit {
		defer func() {
			_, err := app.Commit()
			if err != nil {
				panic(err)
			}
		}()
	}

	txBuilder := app.txConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, fmt.Errorf("failed to set message: %w", err)
	}

	bz, err := app.txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	// if cfg.AutomaticFinalizeBlock {
	height := app.LastBlockHeight() + 1

	_, err = app.ProcessProposal(&cmtabcitypes.RequestProcessProposal{
		Height: height,
		Txs:    [][]byte{bz},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run process proposal: %w", err)
	}

	app.logger.Info("Running msg", "msg", msg.String())

	resp, err := app.FinalizeBlock(&cmtabcitypes.RequestFinalizeBlock{
		Height: height,
		Txs:    [][]byte{bz},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run finalize block: %w", err)
	}

	if resp.GetTxResults()[0].GetCode() != 0 {
		return resp.GetTxResults()[0], fmt.Errorf("failed to run tx with error: %s", resp.GetTxResults()[0].GetLog())
	}

	return resp.GetTxResults()[0], nil
	// }

	// handler := app.MsgServiceRouter().Handler(msg)
	// if handler == nil {
	// 	return nil, fmt.Errorf("handler is nil, can't route message %s: %+v", sdk.MsgTypeURL(msg), msg)
	// }

	// msgResult, err := handler(app.ctx, msg)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to execute message %s: %w", sdk.MsgTypeURL(msg), err)
	// }

	// var response *codectypes.Any
	// if len(msgResult.MsgResponses) > 0 {
	// 	msgResponse := msgResult.MsgResponses[0]
	// 	if msgResponse == nil {
	// 		return nil, fmt.Errorf("got nil msg response %s in message result: %s", sdk.MsgTypeURL(msg), msgResult.String())
	// 	}

	// 	response = msgResponse
	// }

	// return nil, nil
}

// Context returns the application context. It can be unwrapped to a sdk.Context,
// with the sdk.UnwrapSDKContext function.
func (app *App) Context() context.Context {
	return app.ctx
}

// QueryHelper returns the application query helper.
// It can be used when registering query services.
func (app *App) QueryHelper() *baseapp.QueryServiceTestHelper {
	// return app.queryHelper
	return baseapp.NewQueryServerTestHelper(app.GetContextForCheckTx([]byte{}), app.interfaceRegistry)
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
