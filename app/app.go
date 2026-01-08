package app

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/app/services"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
)

type AppI interface { //nolint:revive // keeping this name for clarity
	servertypes.Application

	ModuleManager() *module.Manager
	BasicModuleManager() module.BasicManager
	UpgradeKeeper() *upgradekeeper.Keeper
	Configurator() module.Configurator
	SetStoreLoader(loader baseapp.StoreLoader)

	ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (servertypes.ExportedApp, error)
	SimulationManager() *module.SimulationManager
	EncodingConfig() EncodingConfig

	AutoCliOpts() autocli.AppOptions

	DefaultGenesis() map[string]json.RawMessage
}

var _ AppI = &SDKApp{}

type SDKApp struct {
	loaded sync.Once

	cfg SDKAppConfig

	*baseapp.BaseApp
	encodingConfig EncodingConfig

	// storeKeys to access the substores
	storeKeys map[string]*storetypes.KVStoreKey

	// the module manager
	moduleManager      *module.Manager
	basicModuleManager module.BasicManager

	// simulation manager
	simulationManager *module.SimulationManager

	// module Configurator
	configurator module.Configurator

	// essential keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	upgradeKeeper         *upgradekeeper.Keeper
	EvidenceKeeper        *evidencekeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	// supplementary keepers
	MintKeeper         *mintkeeper.Keeper
	FeeGrantKeeper     *feegrantkeeper.Keeper
	AuthzKeeper        *authzkeeper.Keeper
	EpochsKeeper       *epochskeeper.Keeper
	ProtocolPoolKeeper *protocolpoolkeeper.Keeper

	moduleAccountPerms map[string][]string
	orderPreBlockers   []string
	orderBeginBlockers []string
	orderEndBlockers   []string
	orderInitGenesis   []string
	orderExportGenesis []string

	moduleLoader
}

type moduleLoader struct {
	requiredModules []module.AppModule
	optionalModules []module.AppModule
	customModules   []module.AppModule
}

func newModuleLoader() moduleLoader {
	return moduleLoader{
		requiredModules: make([]module.AppModule, 0),
		optionalModules: make([]module.AppModule, 0),
		customModules:   make([]module.AppModule, 0),
	}
}

func initBaseApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	encodingConfig EncodingConfig,
	appConfig SDKAppConfig,
) *baseapp.BaseApp {
	baseAppOptions := []func(*baseapp.BaseApp){
		func(app *baseapp.BaseApp) {
			app.SetMempool(appConfig.Mempool)
		},
		func(app *baseapp.BaseApp) {
			app.SetVerifyVoteExtensionHandler(appConfig.VerifyVoteExtensionHandler)
		},
		func(app *baseapp.BaseApp) {
			app.SetExtendVoteHandler(appConfig.ExtendVoteHandler)
		},
		func(app *baseapp.BaseApp) {
			app.SetPrepareProposal(appConfig.PrepareProposalHandler)
		},
		func(app *baseapp.BaseApp) {
			app.SetProcessProposal(appConfig.ProcessProposalHandler)
		},
	}

	baseAppOptions = append(baseAppOptions, appConfig.BaseAppOptions...)

	bApp := baseapp.NewBaseApp(appConfig.AppName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	bApp.SetTxEncoder(encodingConfig.TxConfig.TxEncoder())

	return bApp
}

func NewSDKApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appConfig SDKAppConfig,
) *SDKApp {
	appConfig.processOptionalModules()

	encodingConfig := NewEncodingConfigFromOptions(appConfig.InterfaceRegistryOptions)
	bApp := initBaseApp(logger, db, traceStore, encodingConfig, appConfig)

	storeKeys := storetypes.NewKVStoreKeys(
		defaultKeys...,
	)

	sdkApp := &SDKApp{
		cfg:                appConfig,
		BaseApp:            bApp,
		encodingConfig:     encodingConfig,
		storeKeys:          storeKeys,
		orderPreBlockers:   appConfig.OrderPreBlockers,
		orderBeginBlockers: appConfig.OrderBeginBlockers,
		orderEndBlockers:   appConfig.OrderEndBlockers,
		orderInitGenesis:   appConfig.OrderInitGenesis,
		orderExportGenesis: appConfig.OrderExportGenesis,
		moduleAccountPerms: appConfig.ModuleAccountPerms,
		moduleLoader:       newModuleLoader(),
	}

	// add keepers
	sdkApp.initConsensusModule(appConfig)
	sdkApp.initAccountModule(appConfig)
	sdkApp.initBankModule(appConfig)
	sdkApp.initVestingModule(appConfig)
	sdkApp.initStakingModule(appConfig)
	sdkApp.initGenutilModule(appConfig)
	sdkApp.initMintModule(appConfig)
	sdkApp.initDistrModules(appConfig)
	sdkApp.initSlashingModule(appConfig)
	sdkApp.initFeeGrantModule(appConfig)
	sdkApp.initAuthzModule(appConfig)
	sdkApp.initUpgradeModule(appConfig)
	sdkApp.initGovModule(appConfig)
	sdkApp.initEvidenceModule(appConfig)
	sdkApp.initEpochsModule(appConfig)

	sdkApp.processHooks()

	return sdkApp
}

func (app *SDKApp) AddModules(modules ...Module) error {
	for _, mod := range modules {
		err := app.addModule(mod)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *SDKApp) addModule(mod Module) error {
	// update MaccPerms
	for moduleAcc, perms := range mod.ModuleAccountPermissions() {
		if _, found := app.moduleAccountPerms[moduleAcc]; found {
			return fmt.Errorf("module account %s already exists in app: %v", moduleAcc, app.moduleAccountPerms)
		}

		app.moduleAccountPerms[moduleAcc] = perms
	}

	// add to store key list
	for name, storeKey := range mod.StoreKeys() {
		if _, found := app.storeKeys[name]; found {
			return fmt.Errorf("module store key %s already exists in app: %v", mod.Name(), app.storeKeys)
		}
		app.storeKeys[name] = storeKey
	}

	// append actual module to the custom module list
	app.customModules = append(app.customModules, mod)

	// append to order of genesis etc
	app.orderPreBlockers = append(app.orderPreBlockers, mod.Name())
	app.orderBeginBlockers = append(app.orderBeginBlockers, mod.Name())
	app.orderEndBlockers = append(app.orderEndBlockers, mod.Name())
	app.orderInitGenesis = append(app.orderInitGenesis, mod.Name())
	app.orderExportGenesis = append(app.orderExportGenesis, mod.Name())

	return nil
}

func (app *SDKApp) LoadModules() {
	app.loaded.Do(app.loadModules)
}

func (app *SDKApp) loadModules() {
	// TODO: set macc perms updated
	app.AccountKeeper.SetAccountPermissions(app.moduleAccountPerms)

	// TODO: set blocked addresses updated
	app.BankKeeper.SetBlockedAddresses(app.BlockedAddresses())

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.moduleManager = module.NewManager(
		append(append(app.requiredModules, app.optionalModules...), app.customModules...)...,
	)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependent module elements, such as codec registration and genesis verification.
	// By default, it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.basicModuleManager = module.NewBasicManagerFromManager(
		app.moduleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{},
			),
		})
	app.basicModuleManager.RegisterLegacyAminoCodec(app.encodingConfig.LegacyAmino)
	app.basicModuleManager.RegisterInterfaces(app.encodingConfig.InterfaceRegistry)

	app.moduleManager.SetOrderPreBlockers(app.orderPreBlockers...)
	app.moduleManager.SetOrderBeginBlockers(app.orderBeginBlockers...)
	app.moduleManager.SetOrderEndBlockers(app.orderEndBlockers...)
	app.moduleManager.SetOrderInitGenesis(app.orderInitGenesis...)
	app.moduleManager.SetOrderExportGenesis(app.orderExportGenesis...)

	app.configurator = module.NewConfigurator(app.encodingConfig.Codec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err := app.moduleManager.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), services.NewAutoCLIQueryService(app.moduleManager.Modules))

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.encodingConfig.Codec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
	}
	app.simulationManager = module.NewSimulationManagerFromAppModules(app.moduleManager.Modules, overrideModules)

	app.simulationManager.RegisterStoreDecoders()

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// default pre and post handlers
	app.setAnteHandler(app.TxConfig())
	app.setPostHandler()

	// initialize stores
	app.MountKVStores(app.storeKeys)
}

// Name returns the Name of the App
func (app *SDKApp) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *SDKApp) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.moduleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *SDKApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.moduleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *SDKApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.moduleManager.EndBlock(ctx)
}

func (app *SDKApp) Configurator() module.Configurator {
	return app.configurator
}

// InitChainer application update at chain initialization
func (app *SDKApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState sdk.GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	_ = app.upgradeKeeper.SetModuleVersionMap(ctx, app.moduleManager.GetVersionMap())
	return app.moduleManager.InitGenesis(ctx, app.encodingConfig.Codec, genesisState)
}

// LoadHeight loads a particular height
func (app *SDKApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SDKApp) LegacyAmino() *codec.LegacyAmino {
	return app.encodingConfig.LegacyAmino
}

// AppCodec returns SimApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SDKApp) AppCodec() *codec.ProtoCodec {
	return app.encodingConfig.Codec
}

// InterfaceRegistry returns SimApp's InterfaceRegistry
func (app *SDKApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.encodingConfig.InterfaceRegistry
}

// TxConfig returns SimApp's TxConfig
func (app *SDKApp) TxConfig() client.TxConfig {
	return app.encodingConfig.TxConfig
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *SDKApp) DefaultGenesis() map[string]json.RawMessage {
	return app.BasicModuleManager().DefaultGenesis(app.encodingConfig.Codec)
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *SDKApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.mustGetStoreKey(storeKey)
}

// GetStoreKeys returns all the stored store Keys.
func (app *SDKApp) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.storeKeys))
	for _, key := range app.storeKeys {
		keys = append(keys, key)
	}

	return keys
}

// SimulationManager implements the SimulationApp interface
func (app *SDKApp) SimulationManager() *module.SimulationManager {
	return app.simulationManager
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *SDKApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new CometBFT queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.basicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *SDKApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.encodingConfig.InterfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *SDKApp) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.encodingConfig.InterfaceRegistry,
		cmtApp.Query,
	)
}

func (app *SDKApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

func (app *SDKApp) setAnteHandler(txConfig client.TxConfig) {
	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
			FeegrantKeeper:  app.FeeGrantKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			SigVerifyOptions: []ante.SigVerificationDecoratorOption{
				// change below as needed.
				ante.WithUnorderedTxGasCost(ante.DefaultUnorderedTxGasCost),
				ante.WithMaxUnorderedTxTimeoutDuration(ante.DefaultMaxTimeoutDuration),
			},
		},
	)
	if err != nil {
		panic(err)
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
}

func (app *SDKApp) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// BlockedAddresses returns all the app's blocked account addresses.
func (app *SDKApp) BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range app.moduleAccountPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}

// AutoCliOpts returns the autocli options for the app.
func (app *SDKApp) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule)
	for _, m := range app.moduleManager.Modules {
		if moduleWithName, ok := m.(NameProvider); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules: modules,
		// TODO options?????
		ModuleOptions:         services.ExtractAutoCLIOptions(app.moduleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

func (app *SDKApp) ModuleManager() *module.Manager {
	return app.moduleManager
}

func (app *SDKApp) BasicModuleManager() module.BasicManager {
	return app.basicModuleManager
}

func (app *SDKApp) UpgradeKeeper() *upgradekeeper.Keeper {
	return app.upgradeKeeper
}

func (app *SDKApp) EncodingConfig() EncodingConfig {
	return app.encodingConfig
}

// UnsafeFindStoreKey fetches a registered StoreKey from the App in linear time.
//
// NOTE: This should only be used in testing.
func (a *SDKApp) UnsafeFindStoreKey(storeKey string) storetypes.StoreKey {
	return a.storeKeys[storeKey]
}
