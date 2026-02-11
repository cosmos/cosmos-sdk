// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package simapp

import (
	"encoding/json"
	"fmt"
	"io"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log/v2"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/simapp/params"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa"
	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	sigtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const appName = "SimApp"

var (
	DefaultNodeHome string

	_ runtime.AppI            = (*SimApp)(nil)
	_ servertypes.Application = (*SimApp)(nil)
)

type SimApp struct {
	*baseapp.BaseApp
	encodingConfig     params.EncodingConfig
	storeKeys          map[string]*storetypes.KVStoreKey
	transientStoreKeys map[string]*storetypes.TransientStoreKey

	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	POAKeeper             *poakeeper.Keeper

	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	sm *module.SimulationManager

	configurator module.Configurator
}

func init() {
	var err error
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory(".simapp")
	if err != nil {
		panic(err)
	}
}

func NewSimApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	interfaceRegistry, _ := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	if err := interfaceRegistry.SigningContext().Validate(); err != nil {
		panic(err)
	}

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	baseAppOptions = append(baseAppOptions, baseapp.SetOptimisticExecution())

	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	storeKeys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		consensusparamtypes.StoreKey,
		govtypes.StoreKey,
		poatypes.StoreKey,
	)

	transientStoreKeys := storetypes.NewTransientStoreKeys(poatypes.TransientStoreKey)

	if err := bApp.RegisterStreamingServices(appOpts, storeKeys); err != nil {
		panic(err)
	}

	app := &SimApp{
		BaseApp: bApp,
		encodingConfig: params.EncodingConfig{
			Amino:             legacyAmino,
			Codec:             appCodec,
			TxConfig:          txConfig,
			InterfaceRegistry: interfaceRegistry,
		},
		storeKeys:          storeKeys,
		transientStoreKeys: transientStoreKeys,
	}

	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(storeKeys[consensusparamtypes.StoreKey]),
		"",
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(storeKeys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		map[string][]string{
			authtypes.FeeCollectorName: nil,
			govtypes.ModuleName:        {authtypes.Burner, authtypes.Staking},
		},
		authcodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(storeKeys[banktypes.StoreKey]),
		app.AccountKeeper,
		map[string]bool{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	app.POAKeeper = poakeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(storeKeys[poatypes.StoreKey]),
		runtime.NewTransientStoreService(transientStoreKeys[poatypes.TransientStoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
	)

	govConfig := govtypes.DefaultConfig()
	app.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(storeKeys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		nil,
		app.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		poakeeper.NewPOACalculateVoteResultsAndVotingPowerFn(*app.POAKeeper),
	)

	app.GovKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
			app.POAKeeper.NewGovHooks(),
		),
	)

	enabledSignModes := append(authtx.DefaultSignModes, sigtypes.SignMode_SIGN_MODE_TEXTUAL)
	txConfigOpts := authtx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: txmodule.NewBankKeeperCoinMetadataQueryFn(app.BankKeeper),
	}
	txConfig, err := authtx.NewTxConfigWithOptions(
		appCodec,
		txConfigOpts,
	)
	if err != nil {
		panic(err)
	}
	app.encodingConfig.TxConfig = txConfig

	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, nil, app,
			txConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		poa.NewAppModule(appCodec, app.POAKeeper, poa.WithSecp256k1Support()), // poa.WithSecp256k1Support() is an optional parameter to allow Secp256k1 keys as validators
	)

	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		})
	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	app.ModuleManager.SetOrderPreBlockers(
		authtypes.ModuleName,
	)
	app.ModuleManager.SetOrderBeginBlockers(
		genutiltypes.ModuleName,
		poatypes.ModuleName,
	)
	app.ModuleManager.SetOrderEndBlockers(
		genutiltypes.ModuleName,
		govtypes.ModuleName,
		poatypes.ModuleName,
		banktypes.ModuleName,
	)

	genesisModuleOrder := []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		genutiltypes.ModuleName,
		consensusparamtypes.ModuleName,
		govtypes.ModuleName,
		poatypes.ModuleName,
	}

	exportModuleOrder := []string{
		consensusparamtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		genutiltypes.ModuleName,
		govtypes.ModuleName,
		poatypes.ModuleName,
	}

	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(exportModuleOrder...)

	app.configurator = module.NewConfigurator(app.encodingConfig.Codec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err = app.ModuleManager.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.encodingConfig.Codec, app.AccountKeeper, authsims.RandomGenesisAccounts),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	app.MountKVStores(storeKeys)
	app.MountTransientStores(transientStoreKeys)

	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler(txConfig)

	app.setPostHandler()

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return app
}

func (app *SimApp) setAnteHandler(txConfig client.TxConfig) {
	anteHandler, err := NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		SignModeHandler: txConfig.SignModeHandler(),
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
}

func (app *SimApp) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

func (app *SimApp) Name() string { return app.BaseApp.Name() }

func (app *SimApp) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

func (app *SimApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

func (app *SimApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

func (a *SimApp) Configurator() module.Configurator {
	return a.configurator
}

func (app *SimApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	return app.ModuleManager.InitGenesis(ctx, app.encodingConfig.Codec, genesisState)
}

func (app *SimApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func (app *SimApp) LegacyAmino() *codec.LegacyAmino {
	return app.encodingConfig.Amino
}

func (app *SimApp) AppCodec() codec.Codec {
	return app.encodingConfig.Codec
}

func (app *SimApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.encodingConfig.InterfaceRegistry
}

func (app *SimApp) TxConfig() client.TxConfig {
	return app.encodingConfig.TxConfig
}

func (app *SimApp) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

func (a *SimApp) DefaultGenesis() map[string]json.RawMessage {
	return a.BasicModuleManager.DefaultGenesis(a.encodingConfig.Codec)
}

func (app *SimApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.storeKeys[storeKey]
}

func (app *SimApp) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.storeKeys))
	for _, key := range app.storeKeys {
		keys = append(keys, key)
	}

	return keys
}

func (app *SimApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

func (app *SimApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

func (app *SimApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.encodingConfig.InterfaceRegistry)
}

func (app *SimApp) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.encodingConfig.InterfaceRegistry,
		cmtApp.Query,
	)
}

func (app *SimApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	earliestHeightFn := func() int64 {
		return app.CommitMultiStore().EarliestVersion()
	}
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg, earliestHeightFn)
}
