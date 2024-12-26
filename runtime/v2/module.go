package runtime

import (
	"fmt"
	"os"
	"slices"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/branch"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/store/v2/root"
)

var (
	_ appmodulev2.AppModule = appModule[transaction.Tx]{}
	_ hasServicesV1         = appModule[transaction.Tx]{}
)

type appModule[T transaction.Tx] struct {
	app *App[T]
}

func (m appModule[T]) IsOnePerModuleType() {}
func (m appModule[T]) IsAppModule()        {}

func (m appModule[T]) RegisterServices(registrar grpc.ServiceRegistrar) error {
	autoCliQueryService, err := services.NewAutoCLIQueryService(m.app.moduleManager.modules)
	if err != nil {
		return err
	}

	autocliv1.RegisterQueryServer(registrar, autoCliQueryService)

	reflectionSvc, err := services.NewReflectionService()
	if err != nil {
		return err
	}
	reflectionv1.RegisterReflectionServiceServer(registrar, reflectionSvc)

	return nil
}

func (m appModule[T]) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: appv1alpha1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Config",
					Short:     "Query the current app config",
				},
			},
			SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
				"autocli": {
					Service: autocliv1.Query_ServiceDesc.ServiceName,
					RpcCommandOptions: []*autocliv1.RpcCommandOptions{
						{
							RpcMethod: "AppOptions",
							Short:     "Query the custom autocli options",
						},
					},
				},
				"reflection": {
					Service: reflectionv1.ReflectionService_ServiceDesc.ServiceName,
					RpcCommandOptions: []*autocliv1.RpcCommandOptions{
						{
							RpcMethod: "FileDescriptors",
							Short:     "Query the app's protobuf file descriptors",
						},
					},
				},
			},
		},
	}
}

func init() {
	appconfig.Register(&runtimev2.Module{},
		appconfig.Provide(
			ProvideAppBuilder[transaction.Tx],
			ProvideModuleManager[transaction.Tx],
			ProvideEnvironment,
			ProvideKVService,
			ProvideModuleConfigMaps,
			ProvideModuleScopedConfigMap,
		),
		appconfig.Invoke(SetupAppBuilder),
	)
}

func ProvideAppBuilder[T transaction.Tx](
	interfaceRegistrar registry.InterfaceRegistrar,
	amino registry.AminoRegistrar,
	storeBuilder root.Builder,
	storeConfig *root.Config,
) (
	*AppBuilder[T],
	*stf.MsgRouterBuilder,
	appmodulev2.AppModule,
	protodesc.Resolver,
	protoregistry.MessageTypeResolver,
) {
	protoFiles := proto.HybridResolver
	protoTypes := protoregistry.GlobalTypes

	// At startup, check that all proto annotations are correct.
	if err := validateProtoAnnotations(protoFiles); err != nil {
		// Once we switch to using protoreflect-based ante handlers, we might
		// want to panic here instead of logging a warning.
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}

	msgRouterBuilder := stf.NewMsgRouterBuilder()
	app := &App[T]{
		interfaceRegistrar: interfaceRegistrar,
		amino:              amino,
		msgRouterBuilder:   msgRouterBuilder,
		queryRouterBuilder: stf.NewMsgRouterBuilder(), // TODO dedicated query router
		queryHandlers:      map[string]appmodulev2.Handler{},
		storeLoader:        DefaultStoreLoader,
	}
	appBuilder := &AppBuilder[T]{app: app, storeBuilder: storeBuilder, storeConfig: storeConfig}

	return appBuilder, msgRouterBuilder, appModule[T]{app}, protoFiles, protoTypes
}

type AppInputs struct {
	depinject.In

	StoreConfig        *root.Config
	Config             *runtimev2.Module
	AppBuilder         *AppBuilder[transaction.Tx]
	ModuleManager      *MM[transaction.Tx]
	InterfaceRegistrar registry.InterfaceRegistrar
	LegacyAmino        registry.AminoRegistrar
	Logger             log.Logger
	StoreBuilder       root.Builder
}

func SetupAppBuilder(inputs AppInputs) {
	app := inputs.AppBuilder.app
	app.config = inputs.Config
	app.logger = inputs.Logger
	app.moduleManager = inputs.ModuleManager
	app.moduleManager.RegisterInterfaces(inputs.InterfaceRegistrar)
	app.moduleManager.RegisterLegacyAminoCodec(inputs.LegacyAmino)
	// STF requires some state to run
	inputs.StoreBuilder.RegisterKey("stf")
}

func ProvideModuleManager[T transaction.Tx](
	logger log.Logger,
	config *runtimev2.Module,
	modules map[string]appmodulev2.AppModule,
) *MM[T] {
	return NewModuleManager[T](logger, config, modules)
}

func ProvideKVService(
	config *runtimev2.Module,
	key depinject.ModuleKey,
	kvFactory store.KVStoreServiceFactory,
	storeBuilder root.Builder,
) (store.KVStoreService, store.MemoryStoreService) {
	// skips modules that have no store
	if slices.Contains(config.SkipStoreKeys, key.Name()) {
		return &failingStoreService{}, &failingStoreService{}
	}
	var kvStoreKey string
	override := storeKeyOverride(config, key.Name())
	if override != nil {
		kvStoreKey = override.KvStoreKey
	} else {
		kvStoreKey = key.Name()
	}

	storeBuilder.RegisterKey(kvStoreKey)
	return kvFactory([]byte(kvStoreKey)), stf.NewMemoryStoreService([]byte(fmt.Sprintf("memory:%s", kvStoreKey)))
}

func storeKeyOverride(config *runtimev2.Module, moduleName string) *runtimev2.StoreKeyConfig {
	for _, cfg := range config.OverrideStoreKeys {
		if cfg.ModuleName == moduleName {
			return cfg
		}
	}
	return nil
}

// ProvideEnvironment provides the environment for keeper modules, while maintaining backward compatibility and provide services directly as well.
func ProvideEnvironment(
	logger log.Logger,
	key depinject.ModuleKey,
	kvService store.KVStoreService,
	memKvService store.MemoryStoreService,
	headerService header.Service,
	gasService gas.Service,
	eventService event.Service,
	branchService branch.Service,
	routerBuilder RouterServiceBuilder,
) appmodulev2.Environment {
	return appmodulev2.Environment{
		Logger:             logger,
		BranchService:      branchService,
		EventService:       eventService,
		GasService:         gasService,
		HeaderService:      headerService,
		QueryRouterService: routerBuilder.BuildQueryRouter(),
		MsgRouterService:   routerBuilder.BuildMsgRouter([]byte(key.Name())),
		TransactionService: services.NewContextAwareTransactionService(),
		KVStoreService:     kvService,
		MemStoreService:    memKvService,
	}
}

// RouterServiceBuilder builds the msg router and query router service during app initialization.
// this is mainly use for testing to override message router service in the environment and not in stf.
type RouterServiceBuilder interface {
	// BuildMsgRouter return a msg router service.
	// - actor is the module store key.
	BuildMsgRouter(actor []byte) router.Service
	BuildQueryRouter() router.Service
}

type RouterServiceFactory func([]byte) router.Service

// routerBuilder implements RouterServiceBuilder
type routerBuilder struct {
	msgRouterServiceFactory RouterServiceFactory
	queryRouter             router.Service
}

func NewRouterBuilder(
	msgRouterServiceFactory RouterServiceFactory,
	queryRouter router.Service,
) RouterServiceBuilder {
	return routerBuilder{
		msgRouterServiceFactory: msgRouterServiceFactory,
		queryRouter:             queryRouter,
	}
}

func (b routerBuilder) BuildMsgRouter(actor []byte) router.Service {
	return b.msgRouterServiceFactory(actor)
}

func (b routerBuilder) BuildQueryRouter() router.Service {
	return b.queryRouter
}

// DefaultServiceBindings provides default services for the following service interfaces:
// - store.KVStoreServiceFactory
// - header.Service
// - comet.Service
// - event.Service
// - store/v2/root.Builder
// - branch.Service
// - RouterServiceBuilder
//
// They are all required.  For most use cases these default services bindings should be sufficient.
// Power users (or tests) may wish to provide their own services bindings, in which case they must
// supply implementations for each of the above interfaces.
func DefaultServiceBindings() depinject.Config {
	var (
		kvServiceFactory store.KVStoreServiceFactory = func(actor []byte) store.KVStoreService {
			return services.NewGenesisKVService(
				actor,
				stf.NewKVStoreService(actor),
			)
		}
		routerBuilder RouterServiceBuilder = routerBuilder{
			msgRouterServiceFactory: stf.NewMsgRouterService,
			queryRouter:             stf.NewQueryRouterService(),
		}
		cometService  comet.Service = &services.ContextAwareCometInfoService{}
		headerService               = services.NewGenesisHeaderService(stf.HeaderService{})
		eventService                = services.NewGenesisEventService(stf.NewEventService())
		storeBuilder                = root.NewBuilder()
		branchService               = stf.BranchService{}
		gasService                  = stf.NewGasMeterService()
	)
	return depinject.Supply(
		kvServiceFactory,
		routerBuilder,
		headerService,
		cometService,
		eventService,
		storeBuilder,
		branchService,
		gasService,
	)
}
