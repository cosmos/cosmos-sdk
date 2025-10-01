package simapp

import (
	"fmt"
	"io"
	"maps"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cast"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	testdata_pulsar "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/epochs"
	epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool"
	protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const appName = "SimApp"

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// module account permissions
	maccPerms = map[string][]string{
		protocolpooltypes.ModuleName:                nil,
		protocolpooltypes.ProtocolPoolEscrowAccount: nil,
	}
)

var (
	_ runtime.AppI            = (*SimApp)(nil)
	_ servertypes.Application = (*SimApp)(nil)
)

// SimApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type SimApp struct {
	*SDKApp

	// supplementary keepers
	FeeGrantKeeper     feegrantkeeper.Keeper
	AuthzKeeper        authzkeeper.Keeper
	EpochsKeeper       epochskeeper.Keeper
	ProtocolPoolKeeper protocolpoolkeeper.Keeper
}

func init() {
	var err error
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory(".simapp")
	if err != nil {
		panic(err)
	}
}

// NewSimApp returns a reference to an initialized SimApp.
func NewSimApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	sdkApp := NewSDKApp(logger, db, traceStore, appOpts, baseAppOptions...)
	maps.Copy(maccPerms, defaultMaccPerms)

	app := &SimApp{
		SDKApp: sdkApp,
	}

	app.ProtocolPoolKeeper = protocolpoolkeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.Keys[protocolpooltypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	sdkApp.DistrKeeper = distrkeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.Keys[distrtypes.StoreKey]),
		sdkApp.AccountKeeper,
		sdkApp.BankKeeper,
		sdkApp.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		distrkeeper.WithExternalCommunityPool(app.ProtocolPoolKeeper),
	)

	sdkApp.SlashingKeeper = slashingkeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		sdkApp.EncodingConfig.LegacyAmino,
		runtime.NewKVStoreService(sdkApp.Keys[slashingtypes.StoreKey]),
		sdkApp.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.Keys[feegrant.StoreKey]),
		sdkApp.AccountKeeper,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	sdkApp.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			sdkApp.DistrKeeper.Hooks(),
			sdkApp.SlashingKeeper.Hooks(),
		),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(sdkApp.Keys[authzkeeper.StoreKey]),
		sdkApp.EncodingConfig.Codec,
		sdkApp.MsgServiceRouter(),
		sdkApp.AccountKeeper,
	)

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	// set the governance module account as the authority for conducting upgrades
	sdkApp.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(sdkApp.Keys[upgradetypes.StoreKey]),
		sdkApp.EncodingConfig.Codec,
		homePath,
		sdkApp.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler)
	govConfig := govtypes.DefaultConfig()
	/*
		Example of setting gov params:
		govConfig.MaxMetadataLen = 10000
	*/
	govKeeper := govkeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.Keys[govtypes.StoreKey]),
		sdkApp.AccountKeeper,
		sdkApp.BankKeeper,
		sdkApp.StakingKeeper,
		sdkApp.DistrKeeper,
		sdkApp.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		// govkeeper.WithCustomCalculateVoteResultsAndVotingPowerFn(...), // Add if you want to use a custom vote calculation function.
	)

	// Set legacy router for backwards compatibility with gov v1beta1
	govKeeper.SetLegacyRouter(govRouter)

	sdkApp.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.Keys[evidencetypes.StoreKey]),
		sdkApp.StakingKeeper,
		sdkApp.SlashingKeeper,
		sdkApp.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	sdkApp.EvidenceKeeper = *evidenceKeeper

	app.EpochsKeeper = epochskeeper.NewKeeper(
		runtime.NewKVStoreService(sdkApp.Keys[epochstypes.StoreKey]),
		sdkApp.EncodingConfig.Codec,
	)

	app.EpochsKeeper.SetHooks(
		epochstypes.NewMultiEpochHooks(
		// insert epoch hooks receivers here
		),
	)

	/****  Module Options ****/

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	sdkApp.ModuleManager = module.NewManager(
		genutil.NewAppModule(
			sdkApp.AccountKeeper, sdkApp.StakingKeeper, sdkApp,
			sdkApp.TxConfig(),
		),
		auth.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.AccountKeeper, authsims.RandomGenesisAccounts, nil),
		vesting.NewAppModule(sdkApp.AccountKeeper, sdkApp.BankKeeper),
		bank.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.BankKeeper, sdkApp.AccountKeeper, nil),
		feegrantmodule.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.AccountKeeper, sdkApp.BankKeeper, app.FeeGrantKeeper, sdkApp.EncodingConfig.InterfaceRegistry),
		gov.NewAppModule(sdkApp.EncodingConfig.Codec, &sdkApp.GovKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, nil),
		mint.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.MintKeeper, sdkApp.AccountKeeper, nil, nil),
		slashing.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.SlashingKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, sdkApp.StakingKeeper, nil, app.EncodingConfig.InterfaceRegistry),
		distr.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.DistrKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, sdkApp.StakingKeeper, nil),
		staking.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.StakingKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, nil),
		upgrade.NewAppModule(sdkApp.UpgradeKeeper, sdkApp.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(sdkApp.EvidenceKeeper),
		authzmodule.NewAppModule(sdkApp.EncodingConfig.Codec, app.AuthzKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, sdkApp.EncodingConfig.InterfaceRegistry),
		consensus.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.ConsensusParamsKeeper),
		epochs.NewAppModule(app.EpochsKeeper),
		protocolpool.NewAppModule(app.ProtocolPoolKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper),
	)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependent module elements, such as codec registration and genesis verification.
	// By default, it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	sdkApp.BasicModuleManager = module.NewBasicManagerFromManager(
		sdkApp.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{},
			),
		})
	sdkApp.BasicModuleManager.RegisterLegacyAminoCodec(sdkApp.EncodingConfig.LegacyAmino)
	sdkApp.BasicModuleManager.RegisterInterfaces(sdkApp.EncodingConfig.InterfaceRegistry)

	// NOTE: upgrade module is required to be prioritized
	sdkApp.ModuleManager.SetOrderPreBlockers(
		upgradetypes.ModuleName,
		authtypes.ModuleName,
	)
	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	sdkApp.ModuleManager.SetOrderBeginBlockers(
		minttypes.ModuleName,
		distrtypes.ModuleName,
		protocolpooltypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		epochstypes.ModuleName,
	)
	sdkApp.ModuleManager.SetOrderEndBlockers(
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		protocolpooltypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	genesisModuleOrder := []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		epochstypes.ModuleName,
		protocolpooltypes.ModuleName,
	}

	exportModuleOrder := []string{
		consensusparamtypes.ModuleName,
		authtypes.ModuleName,
		protocolpooltypes.ModuleName, // Must be exported before bank
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		epochstypes.ModuleName,
	}

	sdkApp.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	sdkApp.ModuleManager.SetOrderExportGenesis(exportModuleOrder...)

	// Uncomment if you want to set a custom migration order here.
	// app.ModuleManager.SetOrderMigrations(custom order)

	sdkApp.configurator = module.NewConfigurator(sdkApp.EncodingConfig.Codec, sdkApp.MsgServiceRouter(), sdkApp.GRPCQueryRouter())
	err := sdkApp.ModuleManager.RegisterServices(sdkApp.configurator)
	if err != nil {
		panic(err)
	}

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	// Make sure it's called after `app.ModuleManager` and `app.Configurator` are set.
	app.RegisterUpgradeHandlers()

	autocliv1.RegisterQueryServer(sdkApp.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(sdkApp.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(sdkApp.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	testdata_pulsar.RegisterQueryServer(sdkApp.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.AccountKeeper, authsims.RandomGenesisAccounts, nil),
	}
	sdkApp.simulationManager = module.NewSimulationManagerFromAppModules(sdkApp.ModuleManager.Modules, overrideModules)

	sdkApp.simulationManager.RegisterStoreDecoders()

	// initialize stores
	sdkApp.MountKVStores(sdkApp.Keys)

	// initialize BaseApp
	sdkApp.SetInitChainer(app.InitChainer)
	sdkApp.SetPreBlocker(app.PreBlocker)
	sdkApp.SetBeginBlocker(app.BeginBlocker)
	sdkApp.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler(sdkApp.TxConfig())

	// In v0.46, the SDK introduces _postHandlers_. PostHandlers are like
	// antehandlers, but are run _after_ the `runMsgs` execution. They are also
	// defined as a chain, and have the same signature as antehandlers.
	//
	// In baseapp, postHandlers are run in the same store branch as `runMsgs`,
	// meaning that both `runMsgs` and `postHandler` state will be committed if
	// both are successful, and both will be reverted if any of the two fails.
	//
	// The SDK exposes a default postHandlers chain
	//
	// Please note that changing any of the anteHandler or postHandler chain is
	// likely to be a state-machine breaking change, which needs a coordinated
	// upgrade.
	app.setPostHandler()

	if loadLatest {
		if err := sdkApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return app
}

func (app *SimApp) setAnteHandler(txConfig client.TxConfig) {
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

func (app *SimApp) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// GetMaccPerms returns a copy of the module account permissions
//
// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	return maps.Clone(maccPerms)
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}
