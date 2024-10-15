package simapp

import (
	_ "embed"
	"fmt"

	"github.com/spf13/viper"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	serverstore "cosmossdk.io/server/v2/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/root"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	lockupdepinject "cosmossdk.io/x/accounts/defaults/lockup/depinject"
	multisigdepinject "cosmossdk.io/x/accounts/defaults/multisig/depinject"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	_ "github.com/cosmos/cosmos-sdk/x/genutil"
)

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome string

// SimApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type SimApp[T transaction.Tx] struct {
	*runtime.App[T]
	legacyAmino       registry.AminoRegistrar
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry
	store             store.RootStore

	// required keepers during wiring
	// others keepers are all in the app
	UpgradeKeeper *upgradekeeper.Keeper
	StakingKeeper *stakingkeeper.Keeper
}

func init() {
	var err error
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory(".simappv2")
	if err != nil {
		panic(err)
	}
}

// AppConfig returns the default app config.
func AppConfig() depinject.Config {
	return depinject.Configs(
		ModuleConfig, // Alternatively use appconfig.LoadYAML(AppConfigYAML)
		runtime.DefaultServiceBindings(),
		depinject.Provide(
			codec.ProvideInterfaceRegistry,
			codec.ProvideAddressCodec,
			codec.ProvideProtoCodec,
			codec.ProvideLegacyAmino,
			ProvideModuleScopedConfigMap,
			SanelyProvideModuleConfigMap,
			ProvideRootStoreConfig,
		),
		depinject.Invoke(
			std.RegisterInterfaces,
			std.RegisterLegacyAminoCodec,
		),
	)
}

func NewSimAppWithConfig[T transaction.Tx](
	config depinject.Config,
	outputs ...any,
) (*SimApp[T], error) {
	var (
		app          = &SimApp[T]{}
		appBuilder   *runtime.AppBuilder[T]
		storeBuilder root.Builder
		logger       log.Logger

		// merge the AppConfig and other configuration in one config
		appConfig = depinject.Configs(
			AppConfig(),
			config,
			depinject.Provide(
				multisigdepinject.ProvideAccount,
				basedepinject.ProvideAccount,
				lockupdepinject.ProvideAllLockupAccounts,
				basedepinject.ProvideSecp256K1PubKey,
			),
		)
	)

	outputs = append(outputs,
		&logger,
		&storeBuilder,
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.UpgradeKeeper,
		&app.StakingKeeper)

	if err := depinject.Inject(appConfig, outputs...); err != nil {
		return nil, err
	}

	var err error
	app.App, err = appBuilder.Build()
	if err != nil {
		return nil, err
	}

	app.store = storeBuilder.Get()
	if app.store == nil {
		return nil, fmt.Errorf("store builder not return a db")
	}

	/****  Module Options ****/

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	app.RegisterUpgradeHandlers()

	if err = app.LoadLatest(); err != nil {
		return nil, err
	}
	return app, nil
}

// NewSimApp returns a reference to an initialized SimApp.
func NewSimApp[T transaction.Tx](
	logger log.Logger,
	viper *viper.Viper,
) *SimApp[T] {
	var (
		app          = &SimApp[T]{}
		appBuilder   *runtime.AppBuilder[T]
		err          error
		storeBuilder root.Builder

		// merge the AppConfig and other configuration in one config
		appConfig = depinject.Configs(
			AppConfig(),
			depinject.Supply(
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
				// custom function that implements the minttypes.InflationCalculationFn
				// interface.
			),
			depinject.Provide(
				// inject desired account types:
				multisigdepinject.ProvideAccount,
				basedepinject.ProvideAccount,
				lockupdepinject.ProvideAllLockupAccounts,

				// provide base account options
				basedepinject.ProvideSecp256K1PubKey,
				// if you want to provide a custom public key you
				// can do it from here.
				// Example:
				// 		basedepinject.ProvideCustomPubkey[Ed25519PublicKey]()
				//
				// You can also provide a custom public key with a custom validation function:
				//
				// 		basedepinject.ProvideCustomPubKeyAndValidationFunc(func(pub Ed25519PublicKey) error {
				//			if len(pub.Key) != 64 {
				//				return fmt.Errorf("invalid pub key size")
				//			}
				// 		})
			),
		)
	)

	if err := depinject.Inject(appConfig,
		&storeBuilder,
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.UpgradeKeeper,
		&app.StakingKeeper,
	); err != nil {
		panic(err)
	}

	// store/v2 follows a slightly more eager config life cycle than server components
	storeConfig, err := serverstore.UnmarshalConfig(viper.AllSettings())
	if err != nil {
		panic(err)
	}

	app.store, err = storeBuilder.Build(logger, storeConfig)
	if err != nil {
		panic(err)
	}

	app.App, err = appBuilder.Build()
	if err != nil {
		panic(err)
	}

	/****  Module Options ****/

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	app.RegisterUpgradeHandlers()

	// TODO (here or in runtime/v2)
	// wire simulation manager
	// wire unordered tx manager

	if err := app.LoadLatest(); err != nil {
		panic(err)
	}
	return app
}

func (app *SimApp[T]) Build() error {
	return nil
}

// AppCodec returns SimApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SimApp[T]) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns SimApp's InterfaceRegistry.
func (app *SimApp[T]) InterfaceRegistry() server.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns SimApp's TxConfig.
func (app *SimApp[T]) TxConfig() client.TxConfig {
	return app.txConfig
}

func (app *SimApp[T]) GetStore() store.RootStore {
	return app.store
}

type GlobalConfig server.ConfigMap
type ModuleConfigMaps map[string]server.ConfigMap

// TODO combine below 2 functions
// - linear search for module name in provider is OK
// - move elsewhere, server/v2 or runtime/v2 ?

func SanelyProvideModuleConfigMap(
	moduleConfigs []server.ModuleConfigMap,
	globalConfig GlobalConfig,
) ModuleConfigMaps {
	moduleConfigMaps := make(ModuleConfigMaps)
	for _, moduleConfig := range moduleConfigs {
		cfg := moduleConfig.Config
		name := moduleConfig.Module
		moduleConfigMaps[name] = make(server.ConfigMap)
		for flag, df := range cfg {
			if val, ok := globalConfig[flag]; ok {
				moduleConfigMaps[name][flag] = val
			} else {
				moduleConfigMaps[name][flag] = df
			}
		}
	}
	return moduleConfigMaps
}

func ProvideModuleScopedConfigMap(
	key depinject.ModuleKey,
	moduleConfigs ModuleConfigMaps,
) server.ConfigMap {
	return moduleConfigs[key.Name()]
}

func ProvideRootStoreConfig(config GlobalConfig) (*root.Config, error) {
	return serverstore.UnmarshalConfig(config)
}
