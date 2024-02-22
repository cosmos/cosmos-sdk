package runtime

import (
	"fmt"
	"os"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	stakingmodulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/server/v2/stf"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	_ appmodule.AppModule   = appModule{}
	_ appmodule.HasServices = appModule{}
)

type appModule struct {
	app *App
}

func (m appModule) RegisterServices(registar grpc.ServiceRegistrar) error {
	autoCliQueryService, err := services.NewAutoCLIQueryService(m.app.moduleManager.modules)
	if err != nil {
		return err
	}

	autocliv1.RegisterQueryServer(registar, autoCliQueryService)

	reflectionSvc, err := services.NewReflectionService()
	if err != nil {
		return err
	}
	reflectionv1.RegisterReflectionServiceServer(registar, reflectionSvc)

	return nil
}

func (m appModule) IsOnePerModuleType() {}
func (m appModule) IsAppModule()        {}

func init() {
	appconfig.Register(&runtimev2.Module{},
		appconfig.Provide(
			ProvideApp,
			ProvideInterfaceRegistry,
			ProvideKVStoreKey,
			ProvideEnvironment,
			ProvideModuleManager,
			ProvideTransientStoreKey,
			ProvideMemoryStoreKey,
			ProvideAddressCodec,
		),
		appconfig.Invoke(SetupAppBuilder),
	)
}

func ProvideApp(interfaceRegistry codectypes.InterfaceRegistry) (
	codec.Codec,
	*codec.LegacyAmino,
	*AppBuilder,
	*stf.MsgRouterBuilder,
	appmodule.AppModule,
	protodesc.Resolver,
	protoregistry.MessageTypeResolver,
	error,
) {
	protoFiles := proto.HybridResolver
	protoTypes := protoregistry.GlobalTypes

	// At startup, check that all proto annotations are correct.
	if err := msgservice.ValidateProtoAnnotations(protoFiles); err != nil {
		// Once we switch to using protoreflect-based ante handlers, we might
		// want to panic here instead of logging a warning.
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}

	amino := codec.NewLegacyAmino()

	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(amino)

	cdc := codec.NewProtoCodec(interfaceRegistry)
	msgRouterBuilder := stf.NewMsgRouterBuilder()
	app := &App{
		storeKeys:         nil,
		interfaceRegistry: interfaceRegistry,
		cdc:               cdc,
		amino:             amino,
		msgRouterBuilder:  msgRouterBuilder,
	}
	appBuilder := &AppBuilder{app: app}

	return cdc, amino, appBuilder, msgRouterBuilder, appModule{app}, protoFiles, protoTypes, nil
}

type AppInputs struct {
	depinject.In

	AppConfig         *appv1alpha1.Config
	Config            *runtimev2.Module
	AppBuilder        *AppBuilder
	ModuleManager     *MM
	InterfaceRegistry codectypes.InterfaceRegistry
	LegacyAmino       *codec.LegacyAmino
	Logger            log.Logger
}

func SetupAppBuilder(inputs AppInputs) {
	app := inputs.AppBuilder.app
	app.config = inputs.Config
	app.appConfig = inputs.AppConfig
	app.logger = inputs.Logger
	app.moduleManager = inputs.ModuleManager
	app.moduleManager.RegisterInterfaces(inputs.InterfaceRegistry)
	app.moduleManager.RegisterLegacyAminoCodec(inputs.LegacyAmino)
}

func ProvideModuleManager(logger log.Logger, cdc codec.Codec, config *runtimev2.Module, modules map[string]appmodule.AppModule) *MM {
	return NewModuleManager(logger, cdc, config, modules)
}

func ProvideInterfaceRegistry(addressCodec address.Codec, validatorAddressCodec ValidatorAddressCodec, customGetSigners []signing.CustomGetSigner) (codectypes.InterfaceRegistry, error) {
	signingOptions := signing.Options{
		AddressCodec:          addressCodec,
		ValidatorAddressCodec: validatorAddressCodec,
	}
	for _, signer := range customGetSigners {
		signingOptions.DefineCustomGetSigners(signer.MsgType, signer.Fn)
	}

	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	if err != nil {
		return nil, err
	}

	if err := interfaceRegistry.SigningContext().Validate(); err != nil {
		return nil, err
	}

	return interfaceRegistry, nil
}

func registerStoreKey(wrapper *AppBuilder, key storetypes.StoreKey) {
	wrapper.app.storeKeys = append(wrapper.app.storeKeys, key)
}

func storeKeyOverride(config *runtimev2.Module, moduleName string) *runtimev2.StoreKeyConfig {
	for _, cfg := range config.OverrideStoreKeys {
		if cfg.ModuleName == moduleName {
			return cfg
		}
	}
	return nil
}

func ProvideKVStoreKey(config *runtimev2.Module, key depinject.ModuleKey, app *AppBuilder) *storetypes.KVStoreKey {
	override := storeKeyOverride(config, key.Name())

	var storeKeyName string
	if override != nil {
		storeKeyName = override.KvStoreKey
	} else {
		storeKeyName = key.Name()
	}

	storeKey := storetypes.NewKVStoreKey(storeKeyName)
	registerStoreKey(app, storeKey)
	return storeKey
}

func ProvideTransientStoreKey(key depinject.ModuleKey, app *AppBuilder) *storetypes.TransientStoreKey {
	storeKey := storetypes.NewTransientStoreKey(fmt.Sprintf("transient:%s", key.Name()))
	registerStoreKey(app, storeKey)
	return storeKey
}

func ProvideMemoryStoreKey(key depinject.ModuleKey, app *AppBuilder) *storetypes.MemoryStoreKey {
	storeKey := storetypes.NewMemoryStoreKey(fmt.Sprintf("memory:%s", key.Name()))
	registerStoreKey(app, storeKey)
	return storeKey
}

// ProvideEnvironment provides the environment for keeper modules, while maintaining backward compatibility and provide services directly as well.
func ProvideEnvironment(config *runtimev2.Module, key depinject.ModuleKey, app *AppBuilder) (
	appmodule.Environment,
	store.KVStoreService,
	store.MemoryStoreService,
	store.TransientStoreService,
) {
	kvStoreKey := ProvideKVStoreKey(config, key, app)
	kvService := stf.NewKVStoreService([]byte(kvStoreKey.Name()))

	memStoreKey := ProvideMemoryStoreKey(key, app)
	memService := stf.NewMemoryStoreService([]byte(memStoreKey.Name()))

	transientStoreKey := ProvideTransientStoreKey(key, app)
	transientService := stf.NewTransientStoreService([]byte(transientStoreKey.Name()))

	env := appmodule.Environment{
		BranchService:   nil,
		EventService:    stf.NewEventService(),
		GasService:      stf.NewGasMeterService(),
		HeaderService:   nil,
		KVStoreService:  kvService,
		MemStoreService: memService,
	}

	return env, kvService, memService, transientService
}

type (
	// ValidatorAddressCodec is an alias for address.Codec for validator addresses.
	ValidatorAddressCodec address.Codec

	// ConsensusAddressCodec is an alias for address.Codec for validator consensus addresses.
	ConsensusAddressCodec address.Codec
)

type AddressCodecInputs struct {
	depinject.In

	AuthConfig    *authmodulev1.Module    `optional:"true"`
	StakingConfig *stakingmodulev1.Module `optional:"true"`

	AddressCodecFactory          func() address.Codec         `optional:"true"`
	ValidatorAddressCodecFactory func() ValidatorAddressCodec `optional:"true"`
	ConsensusAddressCodecFactory func() ConsensusAddressCodec `optional:"true"`
}

// ProvideAddressCodec provides an address.Codec to the container for any
// modules that want to do address string <> bytes conversion.
func ProvideAddressCodec(in AddressCodecInputs) (address.Codec, ValidatorAddressCodec, ConsensusAddressCodec) {
	if in.AddressCodecFactory != nil && in.ValidatorAddressCodecFactory != nil && in.ConsensusAddressCodecFactory != nil {
		return in.AddressCodecFactory(), in.ValidatorAddressCodecFactory(), in.ConsensusAddressCodecFactory()
	}

	if in.AuthConfig == nil || in.AuthConfig.Bech32Prefix == "" {
		panic("auth config bech32 prefix cannot be empty if no custom address codec is provided")
	}

	if in.StakingConfig == nil {
		in.StakingConfig = &stakingmodulev1.Module{}
	}

	if in.StakingConfig.Bech32PrefixValidator == "" {
		in.StakingConfig.Bech32PrefixValidator = fmt.Sprintf("%svaloper", in.AuthConfig.Bech32Prefix)
	}

	if in.StakingConfig.Bech32PrefixConsensus == "" {
		in.StakingConfig.Bech32PrefixConsensus = fmt.Sprintf("%svalcons", in.AuthConfig.Bech32Prefix)
	}

	return addresscodec.NewBech32Codec(in.AuthConfig.Bech32Prefix),
		addresscodec.NewBech32Codec(in.StakingConfig.Bech32PrefixValidator),
		addresscodec.NewBech32Codec(in.StakingConfig.Bech32PrefixConsensus)
}
