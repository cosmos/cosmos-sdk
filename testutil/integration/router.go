package integration

import (
	"context"
	"fmt"

	cmtabcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

const appName = "integration-app"

// App is a test application that can be used to test the integration of modules.
type App struct {
	*baseapp.BaseApp

	ctx         sdk.Context
	logger      log.Logger
	queryHelper *baseapp.QueryServiceTestHelper
}

// NewIntegrationApp creates an application for testing purposes. This application is able to route messages to their respective handlers.
func NewIntegrationApp(logger log.Logger, keys map[string]*storetypes.KVStoreKey, appCodec codec.Codec, modules ...module.AppModule) *App {
	db := dbm.NewMemDB()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	for _, module := range modules {
		module.RegisterInterfaces(interfaceRegistry)
	}

	txConfig := authtx.NewTxConfig(codec.NewProtoCodec(interfaceRegistry), authtx.DefaultSignModes)
	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseapp.SetChainID(appName))
	bApp.MountKVStores(keys)

	bApp.SetInitChainer(func(ctx sdk.Context, req cmtabcitypes.RequestInitChain) (cmtabcitypes.ResponseInitChain, error) {
		for _, mod := range modules {
			if m, ok := mod.(module.HasGenesis); ok {
				m.InitGenesis(ctx, appCodec, m.DefaultGenesis(appCodec))
			}
		}

		return cmtabcitypes.ResponseInitChain{}, nil
	})

	moduleManager := module.NewManager(modules...)
	bApp.SetBeginBlocker(func(ctx sdk.Context, req cmtabcitypes.RequestBeginBlock) (cmtabcitypes.ResponseBeginBlock, error) {
		return moduleManager.BeginBlock(ctx, req)
	})
	bApp.SetEndBlocker(func(ctx sdk.Context, req cmtabcitypes.RequestEndBlock) (cmtabcitypes.ResponseEndBlock, error) {
		return moduleManager.EndBlock(ctx, req)
	})

	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetMsgServiceRouter(router)

	if err := bApp.LoadLatestVersion(); err != nil {
		panic(fmt.Errorf("failed to load application version from store: %w", err))
	}

	bApp.InitChain(cmtabcitypes.RequestInitChain{ChainId: appName})
	bApp.Commit()

	ctx := bApp.NewContext(true, cmtproto.Header{ChainID: appName})

	return &App{
		BaseApp: bApp,

		logger:      logger,
		ctx:         ctx,
		queryHelper: baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry),
	}
}

// RunMsg allows to run a message and return the response.
// In order to run a message, the application must have a handler for it.
// These handlers are registered on the application message service router.
// The result of the message execution is returned as a Any type.
// That any type can be unmarshaled to the expected response type.
// If the message execution fails, an error is returned.
func (app *App) RunMsg(msg sdk.Msg, option ...Option) (*codectypes.Any, error) {
	// set options
	cfg := Config{}
	for _, opt := range option {
		opt(&cfg)
	}

	if cfg.AutomaticCommit {
		defer app.Commit()
	}

	if cfg.AutomaticBeginEndBlock {
		height := app.LastBlockHeight() + 1
		app.logger.Info("Running beging block", "height", height)
		app.BeginBlock(cmtabcitypes.RequestBeginBlock{Header: cmtproto.Header{Height: height, ChainID: appName}})
		defer func() {
			app.logger.Info("Running end block", "height", height)
			app.EndBlock(cmtabcitypes.RequestEndBlock{})
		}()
	}

	app.logger.Info("Running msg", "msg", msg.String())

	handler := app.MsgServiceRouter().Handler(msg)
	if handler == nil {
		return nil, fmt.Errorf("handler is nil, can't route message %s: %+v", sdk.MsgTypeURL(msg), msg)
	}

	msgResult, err := handler(app.ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to execute message %s: %w", sdk.MsgTypeURL(msg), err)
	}

	var response *codectypes.Any
	if len(msgResult.MsgResponses) > 0 {
		msgResponse := msgResult.MsgResponses[0]
		if msgResponse == nil {
			return nil, fmt.Errorf("got nil msg response %s in message result: %s", sdk.MsgTypeURL(msg), msgResult.String())
		}

		response = msgResponse
	}

	return response, nil
}

// Context returns the application context.
// It can be unwraped to a sdk.Context, with the sdk.UnwrapSDKContext function.
func (app *App) Context() context.Context {
	return app.ctx
}

// QueryHelper returns the application query helper.
// It can be used when registering query services.
func (app *App) QueryHelper() *baseapp.QueryServiceTestHelper {
	return app.queryHelper
}
