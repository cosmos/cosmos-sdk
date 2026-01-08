package app

import (
	"fmt"

	"github.com/spf13/cast"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
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
	"github.com/cosmos/cosmos-sdk/x/gov"
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

func (app *SDKApp) mustGetStoreKey(keyName string) *storetypes.KVStoreKey {
	storeKey, found := app.storeKeys[keyName]
	if !found {
		panic(fmt.Sprintf("store key %s not found, make sure it is initialized in your application", keyName))
	}

	return storeKey
}

// CONTRACT:
// - Account Keeper is initialized
// - Staking Keeper is initialized
func (app *SDKApp) initGenutilModule(_ SDKAppConfig) {
	app.Logger().Info("initializing genutil module")

	module := genutil.NewAppModule(
		app.AccountKeeper,
		app.StakingKeeper,
		app,
		app.TxConfig(),
	)

	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - Account Keeper is initialized
// - Bank Keeper is initialized
func (app *SDKApp) initVestingModule(_ SDKAppConfig) {
	app.Logger().Info("initializing vesting module")

	module := vesting.NewAppModule(app.AccountKeeper, app.BankKeeper)

	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - baseapp must be configured
// - storeKeys are populated
func (app *SDKApp) initConsensusModule(cfg SDKAppConfig) {
	app.Logger().Info("initializing consensus keeper")

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		app.encodingConfig.Codec,
		runtime.NewKVStoreService(app.mustGetStoreKey(consensusparamtypes.StoreKey)),
		cfg.ModuleAuthority,
		runtime.EventService{},
	)
	app.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	module := consensus.NewAppModule(
		app.encodingConfig.Codec,
		app.ConsensusParamsKeeper,
	)
	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - moduleAccountPerms are populated
// - storeKeys are populated
func (app *SDKApp) initAccountModule(cfg SDKAppConfig) {
	app.Logger().Info("initializing account keeper")

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		app.encodingConfig.Codec,
		runtime.NewKVStoreService(app.mustGetStoreKey(authtypes.StoreKey)),
		authtypes.ProtoBaseAccount,
		app.moduleAccountPerms,
		authcodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		cfg.ModuleAuthority,
		authkeeper.WithUnorderedTransactions(cfg.WithUnorderedTx),
	)

	module := auth.NewAppModule(
		app.encodingConfig.Codec,
		app.AccountKeeper,
		authsims.RandomGenesisAccounts,
		nil,
	)
	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - moduleAccountPerms are populated
// - Account Keeper is initialized
// - storeKeys are populated
func (app *SDKApp) initBankModule(cfg SDKAppConfig) {
	app.Logger().Info("initializing bank keeper")

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		app.encodingConfig.Codec,
		runtime.NewKVStoreService(app.mustGetStoreKey(banktypes.StoreKey)),
		app.AccountKeeper,
		app.BlockedAddresses(),
		cfg.ModuleAuthority,
		app.Logger(),
	)

	module := bank.NewAppModule(
		app.encodingConfig.Codec,
		app.BankKeeper,
		app.AccountKeeper,
		nil,
	)

	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - Account Keeper is initialized
// - Bank Keeper is initialized
// - storeKeys are populated
func (app *SDKApp) initStakingModule(cfg SDKAppConfig) {
	app.Logger().Info("initializing staking keeper")

	app.StakingKeeper = stakingkeeper.NewKeeper(
		app.encodingConfig.Codec,
		runtime.NewKVStoreService(app.mustGetStoreKey(stakingtypes.StoreKey)),
		app.AccountKeeper,
		app.BankKeeper,
		cfg.ModuleAuthority,
		authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr), // TODO config option
		authcodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)

	module := staking.NewAppModule(
		app.encodingConfig.Codec,
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		nil,
	)

	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - Account Keeper is initialized
// - Bank Keeper is initialized
// - Staking Keeper is initialized
// - storeKeys are populated
func (app *SDKApp) initMintModule(cfg SDKAppConfig) {
	if cfg.WithMint {
		app.Logger().Info("initializing mint keeper")

		// TODO pipe in mintfn etc
		mintKeeper := mintkeeper.NewKeeper(
			app.encodingConfig.Codec,
			runtime.NewKVStoreService(app.mustGetStoreKey(minttypes.StoreKey)),
			app.StakingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			authtypes.FeeCollectorName,
			cfg.ModuleAuthority,
			// mintkeeper.WithMintFn(mintkeeper.DefaultMintFn(minttypes.DefaultInflationCalculationFn)), custom mintFn can be added here
		)
		app.MintKeeper = &mintKeeper
		app.optionalModules = append(app.optionalModules, mint.NewAppModule(
			app.encodingConfig.Codec,
			*app.MintKeeper,
			app.AccountKeeper,
			nil,
			nil,
		))
	}
}

// CONTRACT:
// - Account Keeper is initialized
// - Bank Keeper is initialized
// - Staking Keeper is initialized
// - storeKeys are populated
func (app *SDKApp) initDistrModules(cfg SDKAppConfig) {
	var distrOpts []distrkeeper.InitOption
	if cfg.WithProtocolPool {
		app.Logger().Info("initializing protocol pool keeper")

		protocolPoolKeeper := protocolpoolkeeper.NewKeeper(
			app.encodingConfig.Codec,
			runtime.NewKVStoreService(app.mustGetStoreKey(protocolpooltypes.StoreKey)),
			app.AccountKeeper,
			app.BankKeeper,
			cfg.ModuleAuthority,
		)
		app.ProtocolPoolKeeper = &protocolPoolKeeper
		distrOpts = append(distrOpts, distrkeeper.WithExternalCommunityPool(app.ProtocolPoolKeeper))
		app.optionalModules = append(app.optionalModules, protocolpool.NewAppModule(
			*app.ProtocolPoolKeeper,
			app.AccountKeeper,
			app.BankKeeper,
		))
	}

	app.Logger().Info("initializing distribution keeper")

	// TODO optional?
	app.DistrKeeper = distrkeeper.NewKeeper(
		app.encodingConfig.Codec,
		runtime.NewKVStoreService(app.mustGetStoreKey(distrtypes.StoreKey)),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		cfg.ModuleAuthority,
		distrOpts...,
	)

	module := distr.NewAppModule(
		app.encodingConfig.Codec,
		app.DistrKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		nil,
	)

	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - Staking Keeper is initialized
// - storeKeys are populated
func (app *SDKApp) initSlashingModule(cfg SDKAppConfig) {
	app.Logger().Info("initializing slashing keeper")

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		app.encodingConfig.Codec,
		app.encodingConfig.LegacyAmino,
		runtime.NewKVStoreService(app.mustGetStoreKey(slashingtypes.StoreKey)),
		app.StakingKeeper,
		cfg.ModuleAuthority,
	)

	module := slashing.NewAppModule(
		app.encodingConfig.Codec,
		app.SlashingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		nil,
		app.encodingConfig.InterfaceRegistry,
	)

	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - Account Keeper is initialized
// - Bank Keeper is initialized
// - storeKeys are populated
func (app *SDKApp) initFeeGrantModule(cfg SDKAppConfig) {
	if cfg.WithFeeGrant {
		app.Logger().Info("initializing fee grant keeper")

		feeGrantKeeper := feegrantkeeper.NewKeeper(
			app.encodingConfig.Codec,
			runtime.NewKVStoreService(app.mustGetStoreKey(feegrant.StoreKey)),
			app.AccountKeeper,
		)
		app.FeeGrantKeeper = &feeGrantKeeper
		app.optionalModules = append(app.optionalModules, feegrantmodule.NewAppModule(
			app.encodingConfig.Codec,
			app.AccountKeeper,
			app.BankKeeper,
			*app.FeeGrantKeeper,
			app.encodingConfig.InterfaceRegistry,
		))
	}
}

func (app *SDKApp) processHooks() {
	// set staking hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
		),
	)
}

// CONTRACT:
// - Account Keeper is initialized
// - Bank Keeper is initialized
// - storeKeys are populated
func (app *SDKApp) initAuthzModule(cfg SDKAppConfig) {
	if cfg.WithAuthz {
		app.Logger().Info("initializing authz keeper")

		authzKeeper := authzkeeper.NewKeeper(
			runtime.NewKVStoreService(app.mustGetStoreKey(authzkeeper.StoreKey)),
			app.encodingConfig.Codec,
			app.MsgServiceRouter(),
			app.AccountKeeper,
		)
		app.AuthzKeeper = &authzKeeper
		app.optionalModules = append(app.optionalModules, authzmodule.NewAppModule(
			app.encodingConfig.Codec,
			*app.AuthzKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.encodingConfig.InterfaceRegistry,
		))
	}
}

// CONTRACT:
// - BaseApp is initialized
// - storeKeys are populated
func (app *SDKApp) initUpgradeModule(cfg SDKAppConfig) {
	app.Logger().Info("initializing upgrade keeper")

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(cfg.AppOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(cfg.AppOpts.Get(flags.FlagHome))
	// set the governance module account as the authority for conducting upgrades
	app.upgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(app.mustGetStoreKey(upgradetypes.StoreKey)),
		app.encodingConfig.Codec,
		homePath,
		app.BaseApp,
		cfg.ModuleAuthority,
	)

	module := upgrade.NewAppModule(
		app.upgradeKeeper,
		app.AccountKeeper.AddressCodec(),
	)

	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - Account Keeper is initialized
// - Bank Keeper is initialized
// - Staking Keeper is initialized
// - Distribution Keeper is initialized
// - storeKeys are populated
func (app *SDKApp) initGovModule(cfg SDKAppConfig) {
	// TODO add optionality and gov configs
	// TODO should this be an option?
	app.Logger().Info("initializing gov keeper")

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
		app.encodingConfig.Codec,
		runtime.NewKVStoreService(app.mustGetStoreKey(govtypes.StoreKey)),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		app.MsgServiceRouter(),
		govConfig,
		cfg.ModuleAuthority,
		govkeeper.NewDefaultCalculateVoteResultsAndVotingPower(app.StakingKeeper),
	)

	// Set legacy router for backwards compatibility with gov v1beta1
	govKeeper.SetLegacyRouter(govRouter)

	app.GovKeeper = *govKeeper.SetHooks(govtypes.NewMultiGovHooks())

	module := gov.NewAppModule(
		app.encodingConfig.Codec,
		&app.GovKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		nil,
	)

	app.requiredModules = append(app.requiredModules, module)
}

// CONTRACT:
// - Account Keeper is initialized
// - SlashingKeeper Keeper is initialized
// - Staking Keeper is initialized
// - storeKeys are populated
func (app *SDKApp) initEvidenceModule(cfg SDKAppConfig) {
	app.Logger().Info("initializing evidence module")

	// create evidence keeper with router
	app.EvidenceKeeper = evidencekeeper.NewKeeper(
		app.encodingConfig.Codec,
		runtime.NewKVStoreService(app.mustGetStoreKey(evidencetypes.StoreKey)),
		app.StakingKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)

	module := evidence.NewAppModule(*app.EvidenceKeeper)

	app.requiredModules = append(app.requiredModules, module)
}

func (app *SDKApp) initEpochsModule(cfg SDKAppConfig) {
	if cfg.WithEpochs {
		app.Logger().Info("initializing epochs module")

		epochsKeeper := epochskeeper.NewKeeper(
			runtime.NewKVStoreService(app.mustGetStoreKey(epochstypes.StoreKey)),
			app.encodingConfig.Codec,
		)
		app.EpochsKeeper = &epochsKeeper

		app.EpochsKeeper.SetHooks(
			epochstypes.NewMultiEpochHooks(
			// insert epoch hooks receivers here
			),
		)
		app.optionalModules = append(app.optionalModules, epochs.NewAppModule(app.EpochsKeeper))
	}
}
