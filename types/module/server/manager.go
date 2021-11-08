package server

import (
	"fmt"
	"reflect"

	gogogrpc "github.com/gogo/protobuf/grpc"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

// Manager is the server module manager
type Manager struct {
	baseApp                    *baseapp.BaseApp
	cdc                        *codec.ProtoCodec
	keys                       map[string]ModuleKey
	modules                    []module.Module
	router                     *router
	requiredServices           map[reflect.Type]bool
	registerInvariantsHandler  map[string]RegisterInvariantsHandler
	weightedOperationsHandlers map[string]WeightedOperationsHandler
}

// NewManager creates a new Manager
func NewManager(baseApp *baseapp.BaseApp, cdc *codec.ProtoCodec) *Manager {
	return &Manager{
		baseApp: baseApp,
		cdc:     cdc,
		keys:    map[string]ModuleKey{},
		router: &router{
			handlers:         map[string]handler{},
			providedServices: map[reflect.Type]bool{},
			msgServiceRouter: baseApp.MsgServiceRouter(),
		},
		requiredServices: map[reflect.Type]bool{},
		// weightedOperationsHandlers: map[string]WeightedOperationsHandler{},
	}
}

func (mm *Manager) GetWeightedOperationsHandlers() map[string]WeightedOperationsHandler {
	return mm.weightedOperationsHandlers
}

// RegisterModules registers modules with the Manager and registers their services.
func (mm *Manager) RegisterModules(modules []module.Module) error {
	mm.modules = modules
	// First we register all interface types. This is done for all modules first before registering
	// any services in case there are any weird dependencies that will cause service initialization to fail.
	for _, mod := range modules {
		// check if we actually have a server module, otherwise skip
		serverMod, ok := mod.(Module)
		if !ok {
			continue
		}

		serverMod.RegisterInterfaces(mm.cdc.InterfaceRegistry())
	}

	// Next we register services
	for _, mod := range modules {
		// check if we actually have a server module, otherwise skip
		serverMod, ok := mod.(Module)
		if !ok {
			continue
		}

		name := serverMod.Name()

		invokerFactory := mm.router.invokerFactory(name)

		key := &rootModuleKey{
			moduleName:     name,
			invokerFactory: invokerFactory,
		}

		if _, found := mm.keys[name]; found {
			return fmt.Errorf("module named %s defined twice", name)
		}

		mm.keys[name] = key
		mm.baseApp.MountStore(key, types.StoreTypeIAVL)

		msgRegistrar := registrar{
			router:       mm.router,
			baseServer:   mm.baseApp.MsgServiceRouter(),
			commitWrites: true,
			moduleName:   name,
		}

		queryRegistrar := registrar{
			router:       mm.router,
			baseServer:   mm.baseApp.GRPCQueryRouter(),
			commitWrites: false,
			moduleName:   name,
		}

		cfg := &configurator{
			msgServer:        msgRegistrar,
			queryServer:      queryRegistrar,
			key:              key,
			cdc:              mm.cdc,
			requiredServices: map[reflect.Type]bool{},
		}

		serverMod.RegisterServices(cfg)

		if cfg.weightedOperationHandler != nil {
			mm.weightedOperationsHandlers[name] = cfg.weightedOperationHandler
		}

		// If mod implements LegacyRouteModule, register module route.
		// This is currently used for the group module as part of #218.
		routeMod, ok := mod.(LegacyRouteModule)
		if ok {
			if r := routeMod.Route(cfg); !r.Empty() {
				mm.baseApp.Router().AddRoute(r)
			}
		}

		for typ := range cfg.requiredServices {
			mm.requiredServices[typ] = true
		}

	}

	return nil
}

// WeightedOperations returns all the modules' weighted operations of an application
func (mm *Manager) WeightedOperations(state sdkmodule.SimulationState, modules []sdkmodule.AppModuleSimulation) []simulation.WeightedOperation {
	wOps := make([]simulation.WeightedOperation, 0, len(modules)+len(mm.weightedOperationsHandlers))
	// adding non ADR-33 modules weighted operations
	for _, m := range modules {
		wOps = append(wOps, m.WeightedOperations(state)...)
	}

	// adding ADR-33 modules weighted operations
	for _, weightedOperationHandler := range mm.weightedOperationsHandlers {
		wOps = append(wOps, weightedOperationHandler(state)...)
	}

	return wOps
}

// AuthorizationMiddleware is a function that allows for more complex authorization than the default authorization scheme,
// such as delegated permissions. It will be called only if the default authorization fails.
type AuthorizationMiddleware func(ctx sdk.Context, methodName string, req sdk.Msg, signer sdk.AccAddress) bool

// SetAuthorizationMiddleware sets AuthorizationMiddleware for the Manager.
func (mm *Manager) SetAuthorizationMiddleware(authzFunc AuthorizationMiddleware) {
	mm.router.authzMiddleware = authzFunc
}

// CompleteInitialization should be the last function on the Manager called before the application starts to perform
// any necessary validation and initialization.
func (mm *Manager) CompleteInitialization() error {
	for typ := range mm.requiredServices {
		if _, found := mm.router.providedServices[typ]; !found {
			return fmt.Errorf("initialization error, service %s was required, but not provided", typ)
		}

	}

	return nil
}

type RegisterInvariantsHandler func(ir sdk.InvariantRegistry)

type configurator struct {
	sdkmodule.Configurator
	msgServer                gogogrpc.Server
	queryServer              gogogrpc.Server
	key                      *rootModuleKey
	cdc                      codec.Codec
	requiredServices         map[reflect.Type]bool
	weightedOperationHandler WeightedOperationsHandler
}

var _ Configurator = &configurator{}

func (c *configurator) RegisterWeightedOperationsHandler(operationsHandler WeightedOperationsHandler) {
	c.weightedOperationHandler = operationsHandler
}

func (c *configurator) MsgServer() gogogrpc.Server {
	return c.msgServer
}

func (c *configurator) QueryServer() gogogrpc.Server {
	return c.queryServer
}

func (c *configurator) ModuleKey() RootModuleKey {
	return c.key
}

func (c *configurator) Marshaler() codec.Codec {
	return c.cdc
}

func (c *configurator) RequireServer(serverInterface interface{}) {
	c.requiredServices[reflect.TypeOf(serverInterface)] = true
}

type WeightedOperationsHandler func(simstate sdkmodule.SimulationState) []simulation.WeightedOperation
