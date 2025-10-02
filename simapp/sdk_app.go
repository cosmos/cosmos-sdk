package simapp

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cast"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	testdata_pulsar "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
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
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
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
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool"
	protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var defaultMaccPerms = map[string][]string{
	authtypes.FeeCollectorName:                  nil,
	distrtypes.ModuleName:                       nil,
	minttypes.ModuleName:                        {authtypes.Minter},
	stakingtypes.BondedPoolName:                 {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:              {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:                         {authtypes.Burner},
	protocolpooltypes.ModuleName:                nil,
	protocolpooltypes.ProtocolPoolEscrowAccount: nil,
}

type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             *codec.ProtoCodec
	LegacyAmino       *codec.LegacyAmino
	TxConfig          client.TxConfig
}

func NewEncodingConfigFromOptions(opts types.InterfaceRegistryOptions) EncodingConfig {
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(opts)
	if err != nil {
		panic(err)
	}

	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	if err := interfaceRegistry.SigningContext().Validate(); err != nil {
		panic(err)
	}

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             appCodec,
		LegacyAmino:       legacyAmino,
		TxConfig:          txConfig,
	}
}

type SDKAppConfig struct {
	AppName string

	AppOpts        servertypes.AppOptions
	BaseAppOptions []func(*baseapp.BaseApp)

	InterfaceRegistryOptions types.InterfaceRegistryOptions

	WithProtocolPool bool
	WithAuthz        bool
	WithEpochs       bool
	WithFeeGrant     bool
	WithMint         bool
	// TODO gov optional?
	// TODO any other optional modules?

	WithUnorderedTx bool

	Keys               []string
	OrderPreBlockers   []string
	OrderBeginBlockers []string
	OrderEndBlockers   []string
	OrderInitGenesis   []string
	OrderExportGenesis []string

	ModuleAccountPerms map[string][]string

	Mempool mempool.Mempool

	VerifyVoteExtensionHandler sdk.VerifyVoteExtensionHandler
	PrepareProposalHandler     sdk.PrepareProposalHandler
	ProcessProposalHandler     sdk.ProcessProposalHandler
	ExtendVoteHandler          sdk.ExtendVoteHandler
}

var (
	defaultKeys = []string{
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		consensusparamtypes.StoreKey,
		upgradetypes.StoreKey,
		feegrant.StoreKey,
		evidencetypes.StoreKey,
		authzkeeper.StoreKey,
		epochstypes.StoreKey,
		protocolpooltypes.StoreKey,
	}

	// NOTE: upgrade module is required to be prioritized
	defaultOrderPreBlockers = []string{
		upgradetypes.ModuleName,
		authtypes.ModuleName,
	}

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	defaultOrderBeginBlockers = []string{
		minttypes.ModuleName,
		distrtypes.ModuleName,
		protocolpooltypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		epochstypes.ModuleName,
	}

	defaultOrderEndBlockers = []string{
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		protocolpooltypes.ModuleName,
	}

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	defaultOrderInitGenesis = []string{
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

	defaultOrderExportGenesis = []string{
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

	defaultInterfaceRegistryOptions = types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	}
)

func DefaultSDKAppConfig(
	name string,
	opts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) SDKAppConfig {
	defaultOptions := server.DefaultBaseappOptions(opts)

	// TODO - populate if nil to fix any issues

	baseAppOptions = append(defaultOptions, baseAppOptions...)

	return SDKAppConfig{
		AppName: name,

		InterfaceRegistryOptions: defaultInterfaceRegistryOptions,

		AppOpts:          opts,
		BaseAppOptions:   baseAppOptions,
		WithProtocolPool: true,
		WithAuthz:        true,
		WithEpochs:       true,
		WithFeeGrant:     true,
		WithMint:         true,

		WithUnorderedTx: true,

		ModuleAccountPerms: defaultMaccPerms,

		OrderPreBlockers:   defaultOrderPreBlockers,
		OrderBeginBlockers: defaultOrderBeginBlockers,
		OrderEndBlockers:   defaultOrderEndBlockers,
		OrderInitGenesis:   defaultOrderInitGenesis,
		OrderExportGenesis: defaultOrderExportGenesis,

		Mempool:                    mempool.NoOpMempool{},
		VerifyVoteExtensionHandler: baseapp.NoOpVerifyVoteExtensionHandler(),
		ExtendVoteHandler:          baseapp.NoOpExtendVote(),
		// leave these as nil for construction later in baseapp by default
		PrepareProposalHandler: nil,
		ProcessProposalHandler: nil,
	}
}

type SDKApp struct {
	loaded sync.Once

	cfg SDKAppConfig

	*baseapp.BaseApp
	EncodingConfig EncodingConfig

	// StoreKeys to access the substores
	StoreKeys map[string]*storetypes.KVStoreKey

	// the module manager
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

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
	UpgradeKeeper         *upgradekeeper.Keeper
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

	requiredModules []module.AppModule
	optionalModules []module.AppModule
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

	bApp := baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	bApp.SetTxEncoder(encodingConfig.TxConfig.TxEncoder())

	return bApp
}

func processOptionalModules(appConfig SDKAppConfig) {
	checkForModuleInclusion := func(moduleName string) func(string) bool {
		return func(s string) bool {
			return moduleName == s
		}
	}

	deleteModuleFromOrdering := func(moduleName string) {
		defaultOrderPreBlockers = slices.DeleteFunc(defaultOrderPreBlockers, checkForModuleInclusion(moduleName))
		defaultOrderBeginBlockers = slices.DeleteFunc(defaultOrderBeginBlockers, checkForModuleInclusion(moduleName))
		defaultOrderEndBlockers = slices.DeleteFunc(defaultOrderEndBlockers, checkForModuleInclusion(moduleName))
		defaultOrderInitGenesis = slices.DeleteFunc(defaultOrderInitGenesis, checkForModuleInclusion(moduleName))
		defaultOrderExportGenesis = slices.DeleteFunc(defaultOrderExportGenesis, checkForModuleInclusion(moduleName))
	}

	if !appConfig.WithProtocolPool {
		// remove from macc permissions
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case protocolpooltypes.ModuleName:
				return true
			case protocolpooltypes.ProtocolPoolEscrowAccount:
				return true
			default:
				return false
			}
		})

		deleteModuleFromOrdering(protocolpooltypes.ModuleName)
	}

	if !appConfig.WithAuthz {
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case authz.ModuleName:
				return true
			default:
				return false
			}
		})

		deleteModuleFromOrdering(authz.ModuleName)
	}

	if !appConfig.WithFeeGrant {
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case feegrant.ModuleName:
				return true
			default:
				return false
			}
		})

		deleteModuleFromOrdering(feegrant.ModuleName)
	}

	if !appConfig.WithMint {
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case minttypes.ModuleName:
				return true
			default:
				return false
			}
		})
	}

	if !appConfig.WithEpochs {
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case epochstypes.ModuleName:
				return true
			default:
				return false
			}
		})

		deleteModuleFromOrdering(epochstypes.ModuleName)
	}
}

func NewSDKApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appConfig SDKAppConfig,
) *SDKApp {
	processOptionalModules(appConfig)

	encodingConfig := NewEncodingConfigFromOptions(appConfig.InterfaceRegistryOptions)
	bApp := initBaseApp(logger, db, traceStore, encodingConfig, appConfig)

	storeKeys := storetypes.NewKVStoreKeys(
		defaultKeys...,
	)

	sdkApp := &SDKApp{
		cfg:                appConfig,
		BaseApp:            bApp,
		EncodingConfig:     encodingConfig,
		StoreKeys:          storeKeys,
		orderPreBlockers:   appConfig.OrderPreBlockers,
		orderBeginBlockers: appConfig.OrderBeginBlockers,
		orderEndBlockers:   appConfig.OrderEndBlockers,
		orderInitGenesis:   appConfig.OrderInitGenesis,
		orderExportGenesis: appConfig.OrderExportGenesis,
		moduleAccountPerms: appConfig.ModuleAccountPerms,
	}

	var optionalModules []module.AppModule

	// set the BaseApp's parameter store
	sdkApp.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.StoreKeys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)
	sdkApp.SetParamStore(sdkApp.ConsensusParamsKeeper.ParamsStore)

	// add keepers
	sdkApp.AccountKeeper = authkeeper.NewAccountKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.StoreKeys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		sdkApp.moduleAccountPerms,
		authcodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authkeeper.WithUnorderedTransactions(appConfig.WithUnorderedTx),
	)

	sdkApp.BankKeeper = bankkeeper.NewBaseKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.StoreKeys[banktypes.StoreKey]),
		sdkApp.AccountKeeper,
		sdkApp.BlockedAddresses(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	// TODO probably just eliminate this and remove textual signing
	// optional: enable sign mode textual by overwriting the default tx config (after setting the bank keeper)
	enabledSignModes := append(authtx.DefaultSignModes, sigtypes.SignMode_SIGN_MODE_TEXTUAL)
	txConfigOpts := authtx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: txmodule.NewBankKeeperCoinMetadataQueryFn(sdkApp.BankKeeper),
	}
	txConfig, err := authtx.NewTxConfigWithOptions(
		sdkApp.EncodingConfig.Codec,
		txConfigOpts,
	)
	if err != nil {
		panic(err)
	}
	sdkApp.EncodingConfig.TxConfig = txConfig

	sdkApp.StakingKeeper = stakingkeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.StoreKeys[stakingtypes.StoreKey]),
		sdkApp.AccountKeeper,
		sdkApp.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr),
		authcodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)

	if appConfig.WithMint {
		mintKeeper := mintkeeper.NewKeeper(
			sdkApp.EncodingConfig.Codec,
			runtime.NewKVStoreService(sdkApp.StoreKeys[minttypes.StoreKey]),
			sdkApp.StakingKeeper,
			sdkApp.AccountKeeper,
			sdkApp.BankKeeper,
			authtypes.FeeCollectorName,
			authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			// mintkeeper.WithMintFn(mintkeeper.DefaultMintFn(minttypes.DefaultInflationCalculationFn)), custom mintFn can be added here
		)
		sdkApp.MintKeeper = &mintKeeper
		optionalModules = append(optionalModules, mint.NewAppModule(sdkApp.EncodingConfig.Codec, *sdkApp.MintKeeper, sdkApp.AccountKeeper, nil, nil))
	}

	var distrOpts []distrkeeper.InitOption
	if appConfig.WithProtocolPool {
		protocolPoolKeeper := protocolpoolkeeper.NewKeeper(
			sdkApp.EncodingConfig.Codec,
			runtime.NewKVStoreService(sdkApp.StoreKeys[protocolpooltypes.StoreKey]),
			sdkApp.AccountKeeper,
			sdkApp.BankKeeper,
			authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		)
		sdkApp.ProtocolPoolKeeper = &protocolPoolKeeper
		distrOpts = append(distrOpts, distrkeeper.WithExternalCommunityPool(sdkApp.ProtocolPoolKeeper))
		optionalModules = append(optionalModules, protocolpool.NewAppModule(*sdkApp.ProtocolPoolKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper))
	}

	sdkApp.DistrKeeper = distrkeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.StoreKeys[distrtypes.StoreKey]),
		sdkApp.AccountKeeper,
		sdkApp.BankKeeper,
		sdkApp.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		distrOpts...,
	)

	sdkApp.SlashingKeeper = slashingkeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		sdkApp.EncodingConfig.LegacyAmino,
		runtime.NewKVStoreService(sdkApp.StoreKeys[slashingtypes.StoreKey]),
		sdkApp.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	if appConfig.WithFeeGrant {
		feeGrantKeeper := feegrantkeeper.NewKeeper(
			sdkApp.EncodingConfig.Codec,
			runtime.NewKVStoreService(sdkApp.StoreKeys[feegrant.StoreKey]),
			sdkApp.AccountKeeper,
		)
		sdkApp.FeeGrantKeeper = &feeGrantKeeper
		optionalModules = append(optionalModules, feegrantmodule.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.AccountKeeper, sdkApp.BankKeeper, *sdkApp.FeeGrantKeeper, sdkApp.EncodingConfig.InterfaceRegistry))
	}

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	sdkApp.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			sdkApp.DistrKeeper.Hooks(),
			sdkApp.SlashingKeeper.Hooks(),
		),
	)

	if appConfig.WithAuthz {
		authzKeeper := authzkeeper.NewKeeper(
			runtime.NewKVStoreService(sdkApp.StoreKeys[authzkeeper.StoreKey]),
			sdkApp.EncodingConfig.Codec,
			sdkApp.MsgServiceRouter(),
			sdkApp.AccountKeeper,
		)
		sdkApp.AuthzKeeper = &authzKeeper
		optionalModules = append(optionalModules, authzmodule.NewAppModule(sdkApp.EncodingConfig.Codec, *sdkApp.AuthzKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, sdkApp.EncodingConfig.InterfaceRegistry))
	}

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appConfig.AppOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appConfig.AppOpts.Get(flags.FlagHome))
	// set the governance module account as the authority for conducting upgrades
	sdkApp.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(sdkApp.StoreKeys[upgradetypes.StoreKey]),
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
		runtime.NewKVStoreService(sdkApp.StoreKeys[govtypes.StoreKey]),
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

	sdkApp.GovKeeper = *govKeeper.SetHooks(govtypes.NewMultiGovHooks())
	//	govtypes.NewMultiGovHooks(
	//		// register the governance hooks
	//	),
	//)

	// create evidence keeper with router
	sdkApp.EvidenceKeeper = evidencekeeper.NewKeeper(
		sdkApp.EncodingConfig.Codec,
		runtime.NewKVStoreService(sdkApp.StoreKeys[evidencetypes.StoreKey]),
		sdkApp.StakingKeeper,
		sdkApp.SlashingKeeper,
		sdkApp.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)

	if appConfig.WithEpochs {
		epochsKeeper := epochskeeper.NewKeeper(
			runtime.NewKVStoreService(sdkApp.StoreKeys[epochstypes.StoreKey]),
			sdkApp.EncodingConfig.Codec,
		)
		sdkApp.EpochsKeeper = &epochsKeeper

		sdkApp.EpochsKeeper.SetHooks(
			epochstypes.NewMultiEpochHooks(
				// insert epoch hooks receivers here
			),
		)
		optionalModules = append(optionalModules, epochs.NewAppModule(*sdkApp.EpochsKeeper))
	}

	requiredModules := []module.AppModule{
		genutil.NewAppModule(
			sdkApp.AccountKeeper, sdkApp.StakingKeeper, sdkApp,
			sdkApp.TxConfig(),
		),
		auth.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.AccountKeeper, authsims.RandomGenesisAccounts, nil),
		vesting.NewAppModule(sdkApp.AccountKeeper, sdkApp.BankKeeper),
		bank.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.BankKeeper, sdkApp.AccountKeeper, nil),
		// todo optional???
		gov.NewAppModule(sdkApp.EncodingConfig.Codec, &sdkApp.GovKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, nil),
		slashing.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.SlashingKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, sdkApp.StakingKeeper, nil, sdkApp.EncodingConfig.InterfaceRegistry),
		distr.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.DistrKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, sdkApp.StakingKeeper, nil),
		staking.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.StakingKeeper, sdkApp.AccountKeeper, sdkApp.BankKeeper, nil),
		upgrade.NewAppModule(sdkApp.UpgradeKeeper, sdkApp.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(*sdkApp.EvidenceKeeper),
		consensus.NewAppModule(sdkApp.EncodingConfig.Codec, sdkApp.ConsensusParamsKeeper),
	}

	sdkApp.requiredModules = requiredModules
	sdkApp.optionalModules = optionalModules

	return sdkApp
}

type AppModule struct {
	module.AppModule
	storeKeys map[string]*storetypes.KVStoreKey
	name      string
	maccPerms map[string][]string
}

func (app *SDKApp) AddModule(module AppModule) error {
	// update maccPerms
	for moduleAcc, perms := range module.maccPerms {
		if _, found := app.moduleAccountPerms[moduleAcc]; found {
			return fmt.Errorf("module account %s already exists in app: %v", moduleAcc, app.moduleAccountPerms)
		}

		app.moduleAccountPerms[moduleAcc] = perms
	}

	// add to store key list
	for name, storeKey := range module.storeKeys {
		if _, found := app.StoreKeys[name]; found {
			return fmt.Errorf("module store key %s already exists in app: %v", module.name, app.StoreKeys)
		}
		app.StoreKeys[name] = storeKey
	}

	// append actual module
	app.optionalModules = append(app.optionalModules, module)

	// append to order of genesis etc
	app.orderPreBlockers = append(app.orderPreBlockers, module.name)
	app.orderBeginBlockers = append(app.orderBeginBlockers, module.name)
	app.orderEndBlockers = append(app.orderEndBlockers, module.name)
	app.orderInitGenesis = append(app.orderInitGenesis, module.name)
	app.orderExportGenesis = append(app.orderExportGenesis, module.name)

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
	app.ModuleManager = module.NewManager(
		append(app.requiredModules, app.optionalModules...)...,
	)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependent module elements, such as codec registration and genesis verification.
	// By default, it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{},
			),
		})
	app.BasicModuleManager.RegisterLegacyAminoCodec(app.EncodingConfig.LegacyAmino)
	app.BasicModuleManager.RegisterInterfaces(app.EncodingConfig.InterfaceRegistry)

	app.ModuleManager.SetOrderPreBlockers(app.orderPreBlockers...)
	app.ModuleManager.SetOrderBeginBlockers(app.orderBeginBlockers...)
	app.ModuleManager.SetOrderEndBlockers(app.orderEndBlockers...)
	app.ModuleManager.SetOrderInitGenesis(app.orderInitGenesis...)
	app.ModuleManager.SetOrderExportGenesis(app.orderExportGenesis...)

	app.configurator = module.NewConfigurator(app.EncodingConfig.Codec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err := app.ModuleManager.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	testdata_pulsar.RegisterQueryServer(app.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.EncodingConfig.Codec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
	}
	app.simulationManager = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)

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
	app.MountKVStores(app.StoreKeys)
}

// Name returns the name of the App
func (app *SDKApp) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *SDKApp) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *SDKApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *SDKApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

func (app *SDKApp) Configurator() module.Configurator {
	return app.configurator
}

// InitChainer application update at chain initialization
func (app *SDKApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	_ = app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	return app.ModuleManager.InitGenesis(ctx, app.EncodingConfig.Codec, genesisState)
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
	return app.EncodingConfig.LegacyAmino
}

// AppCodec returns SimApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SDKApp) AppCodec() codec.Codec {
	return app.EncodingConfig.Codec
}

// InterfaceRegistry returns SimApp's InterfaceRegistry
func (app *SDKApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.EncodingConfig.InterfaceRegistry
}

// TxConfig returns SimApp's TxConfig
func (app *SDKApp) TxConfig() client.TxConfig {
	return app.EncodingConfig.TxConfig
}

// AutoCliOpts returns the autocli options for the app.
func (app *SDKApp) AutoCliOpts() autocli.AppOptions {
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

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *SDKApp) DefaultGenesis() map[string]json.RawMessage {
	return app.BasicModuleManager.DefaultGenesis(app.EncodingConfig.Codec)
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *SDKApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.StoreKeys[storeKey]
}

// GetStoreKeys returns all the stored store Keys.
func (app *SDKApp) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.StoreKeys))
	for _, key := range app.StoreKeys {
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
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *SDKApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.EncodingConfig.InterfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *SDKApp) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.EncodingConfig.InterfaceRegistry,
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
