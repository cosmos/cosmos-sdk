//go:build !app_v1

package simapp

import (
	_ "embed"
	"fmt"
	"io"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/x/accounts"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	consensuskeeper "cosmossdk.io/x/consensus/keeper"
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	_ "cosmossdk.io/x/protocolpool"
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	testdata_pulsar "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/ante/unorderedtx"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome string

var (
	_ runtime.AppI            = (*SimApp)(nil)
	_ servertypes.Application = (*SimApp)(nil)
)

// SimApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type SimApp struct {
	*runtime.App
	legacyAmino       registry.AminoRegistrar
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// required keepers during wiring
	// others keepers are all in the app
	AccountsKeeper        accounts.Keeper
	AuthKeeper            authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper
	CircuitBreakerKeeper  circuitkeeper.Keeper

	// simulation manager
	sm *module.SimulationManager
}

func init() {
	var err error
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory(".simapp")
	if err != nil {
		panic(err)
	}
}

// AppConfig returns the default app config.
func AppConfig() depinject.Config {
	return depinject.Configs(
		appConfig,                               // Alternatively use appconfig.LoadYAML(AppConfigYAML)
		depinject.Provide(ProvideExampleMintFn), // optional: override the mint module's mint function with epoched minting
	)
}

// NewSimApp returns a reference to an initialized SimApp.
func NewSimApp(
	logger log.Logger,
	db corestore.KVStoreWithBatch,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	var (
		app        = &SimApp{}
		appBuilder *runtime.AppBuilder

		// merge the AppConfig and other configuration in one config
		appConfig = depinject.Configs(
			AppConfig(),
			depinject.Supply(
				// supply the application options
				appOpts,
				// supply the logger
				logger,
				// ADVANCED CONFIGURATION

				//
				// AUTH
				//
				// For providing a custom function required in auth to generate custom account types
				// add it below. By default the auth module uses simulation.RandomGenesisAccounts.
				//
				// authtypes.RandomGenesisAccountsFn(simulation.RandomGenesisAccounts),
				//
				// For providing a custom a base account type add it below.
				// By default the auth module uses authtypes.ProtoBaseAccount().
				//
				// func() sdk.AccountI { return authtypes.ProtoBaseAccount() },
				//
				// For providing a different address codec, add it below.
				// By default the auth module uses a Bech32 address codec,
				// with the prefix defined in the auth module configuration.
				//
				// func() address.Codec { return <- custom address codec type -> }

				//
				// STAKING
				//
				// For provinding a different validator and consensus address codec, add it below.
				// By default the staking module uses the bech32 prefix provided in the auth config,
				// and appends "valoper" and "valcons" for validator and consensus addresses respectively.
				// When providing a custom address codec in auth, custom address codecs must be provided here as well.
				//
				// func() runtime.ValidatorAddressCodec { return <- custom validator address codec type -> }
				// func() runtime.ConsensusAddressCodec { return <- custom consensus address codec type -> }

				//
				// MINT
				//

				// For providing a custom inflation function for x/mint add here your
				// custom function that implements the minttypes.MintFn interface.
			),
		)
	)

	var appModules map[string]appmodule.AppModule
	if err := depinject.Inject(appConfig,
		&appBuilder,
		&appModules,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.AuthKeeper,
		&app.AccountsKeeper,
		&app.BankKeeper,
		&app.StakingKeeper,
		&app.SlashingKeeper,
		&app.DistrKeeper,
		&app.UpgradeKeeper,
		&app.FeeGrantKeeper,
		&app.ConsensusParamsKeeper,
		&app.CircuitBreakerKeeper,
	); err != nil {
		panic(err)
	}

	// Below we could construct and set an application specific mempool and
	// ABCI 1.0 PrepareProposal and ProcessProposal handlers. These defaults are
	// already set in the SDK's BaseApp, this shows an example of how to override
	// them.
	//
	// Example:
	//
	// app.App = appBuilder.Build(...)
	// nonceMempool := mempool.NewSenderNonceMempool()
	// abciPropHandler := NewDefaultProposalHandler(nonceMempool, app.App.BaseApp)
	//
	// app.App.BaseApp.SetMempool(nonceMempool)
	// app.App.BaseApp.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// app.App.BaseApp.SetProcessProposal(abciPropHandler.ProcessProposalHandler())
	//
	// Alternatively, you can construct BaseApp options, append those to
	// baseAppOptions and pass them to the appBuilder.
	//
	// Example:
	//
	// prepareOpt = func(app *baseapp.BaseApp) {
	// 	abciPropHandler := baseapp.NewDefaultProposalHandler(nonceMempool, app)
	// 	app.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// }
	// baseAppOptions = append(baseAppOptions, prepareOpt)

	// create and set dummy vote extension handler
	voteExtOp := func(bApp *baseapp.BaseApp) {
		voteExtHandler := NewVoteExtensionHandler()
		voteExtHandler.SetHandlers(bApp)
	}
	baseAppOptions = append(baseAppOptions, voteExtOp, baseapp.SetOptimisticExecution())

	app.App = appBuilder.Build(db, traceStore, baseAppOptions...)

	/****  Module Options ****/

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	app.RegisterUpgradeHandlers()

	// add test gRPC service for testing gRPC queries in isolation
	testdata_pulsar.RegisterQueryServer(app.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AuthKeeper, &app.AccountsKeeper, authsims.RandomGenesisAccounts, nil),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	// A custom InitChainer can be set if extra pre-init-genesis logic is required.
	// By default, when using app wiring enabled module, this is not required.
	// For instance, the upgrade module will set automatically the module version map in its init genesis thanks to app wiring.
	// However, when registering a module manually (i.e. that does not support app wiring), the module version map
	// must be set manually as follow. The upgrade module will de-duplicate the module version map.
	//
	// app.SetInitChainer(func(ctx sdk.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	// 	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	// 	return app.App.InitChainer(ctx, req)
	// })

	// register custom snapshot extensions (if any)
	if manager := app.SnapshotManager(); manager != nil {
		if err := manager.RegisterExtensions(
			unorderedtx.NewSnapshotter(app.UnorderedTxManager),
		); err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %w", err))
		}
	}

	// set custom ante handlers
	app.setCustomAnteHandler()

	if err := app.Load(loadLatest); err != nil {
		panic(err)
	}

	return app
}

// setCustomAnteHandler overwrites default ante handlers with custom ante handlers
// set SkipAnteHandler to true in app config and set custom ante handler on baseapp
func (app *SimApp) setCustomAnteHandler() {
	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			ante.HandlerOptions{
				AccountKeeper:      app.AuthKeeper,
				BankKeeper:         app.BankKeeper,
				ConsensusKeeper:    app.ConsensusParamsKeeper,
				SignModeHandler:    app.txConfig.SignModeHandler(),
				FeegrantKeeper:     app.FeeGrantKeeper,
				SigGasConsumer:     ante.DefaultSigVerificationGasConsumer,
				UnorderedTxManager: app.UnorderedTxManager,
				Environment:        app.AuthKeeper.Environment,
			},
			&app.CircuitBreakerKeeper,
		},
	)
	if err != nil {
		panic(err)
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SimApp) LegacyAmino() *codec.LegacyAmino {
	switch cdc := app.legacyAmino.(type) {
	case *codec.LegacyAmino:
		return cdc
	default:
		panic("unexpected codec type")
	}
}

// AppCodec returns SimApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SimApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns SimApp's InterfaceRegistry.
func (app *SimApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns SimApp's TxConfig
func (app *SimApp) TxConfig() client.TxConfig {
	return app.txConfig
}

// SimulationManager implements the SimulationApp interface
func (app *SimApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *SimApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)
	// register swagger API in app.go so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// GetMaccPerms returns a copy of the module account permissions
//
// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	dup := make(map[string][]string)
	for _, perms := range moduleAccPerms {
		dup[perms.Account] = perms.Permissions
	}

	return dup
}

// BlockedAddresses returns all the app's blocked account addresses.
// This function takes an address.Codec parameter to maintain compatibility
// with the signature of the same function in appV1.
func BlockedAddresses(_ address.Codec) (map[string]bool, error) {
	result := make(map[string]bool)

	if len(blockAccAddrs) > 0 {
		for _, addr := range blockAccAddrs {
			result[addr] = true
		}
	} else {
		for addr := range GetMaccPerms() {
			result[addr] = true
		}
	}

	return result, nil
}
