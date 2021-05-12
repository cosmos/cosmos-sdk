package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogogrpc "github.com/gogo/protobuf/grpc"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Manager is the server module manager
type Manager struct {
	baseApp                   *baseapp.BaseApp
	cdc                       *codec.ProtoCodec
	keys                      map[string]ModuleKey
	router                    *router
	requiredServices          map[reflect.Type]bool
	initGenesisHandlers       map[string]module.InitGenesisHandler
	exportGenesisHandlers     map[string]module.ExportGenesisHandler
	registerInvariantsHandler map[string]RegisterInvariantsHandler
}

// RegisterInvariants registers all module routes and module querier routes
func (mm *Manager) RegisterInvariants(ir sdk.InvariantRegistry) {
	for _, invariant := range mm.registerInvariantsHandler {
		if invariant != nil {
			invariant(ir)
		}
	}
}

// NewManager creates a new Manager
func NewManager(baseApp *baseapp.BaseApp, cdc *codec.ProtoCodec) *Manager {
	return &Manager{
		baseApp:                   baseApp,
		cdc:                       cdc,
		keys:                      map[string]ModuleKey{},
		registerInvariantsHandler: map[string]RegisterInvariantsHandler{},
		initGenesisHandlers:       map[string]module.InitGenesisHandler{},
		exportGenesisHandlers:     map[string]module.ExportGenesisHandler{},
		router: &router{
			handlers:         map[string]handler{},
			providedServices: map[reflect.Type]bool{},
			antiReentryMap:   map[string]bool{},
		},
	}
}

// RegisterModules registers modules with the Manager and registers their services.
func (mm *Manager) RegisterModules(modules []module.Module) error {
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
		mm.baseApp.MountStore(key, sdk.StoreTypeIAVL)

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
			router:           mm.baseApp.Router(), // TODO: remove once #225 addressed
		}

		serverMod.RegisterServices(cfg)
		mm.registerInvariantsHandler[name] = cfg.registerInvariantsHandler
		mm.initGenesisHandlers[name] = cfg.initGenesisHandler
		mm.exportGenesisHandlers[name] = cfg.exportGenesisHandler

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

// AuthorizationMiddleware is a function that allows for more complex authorization than the default authorization scheme,
// such as delegated permissions. It will be called only if the default authorization fails.
type AuthorizationMiddleware func(ctx sdk.Context, methodName string, req sdk.MsgRequest, signer sdk.AccAddress) bool

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

// InitGenesis performs init genesis functionality for modules.
// We pass in existing validatorUpdates from the sdk module Manager.InitGenesis.
func (mm *Manager) InitGenesis(ctx sdk.Context, genesisData map[string]json.RawMessage, validatorUpdates []abci.ValidatorUpdate) abci.ResponseInitChain {
	res, err := initGenesis(ctx, mm.cdc, genesisData, validatorUpdates, mm.initGenesisHandlers)
	if err != nil {
		panic(err)
	}
	return res
}

func initGenesis(ctx sdk.Context, cdc codec.JSONMarshaler,
	genesisData map[string]json.RawMessage, validatorUpdates []abci.ValidatorUpdate,
	initGenesisHandlers map[string]module.InitGenesisHandler) (abci.ResponseInitChain, error) {
	for name, initGenesisHandler := range initGenesisHandlers {
		if genesisData[name] == nil || initGenesisHandler == nil {
			continue
		}
		moduleValUpdates, err := initGenesisHandler(types.Context{Context: ctx}, cdc, genesisData[name])
		if err != nil {
			return abci.ResponseInitChain{}, err
		}

		// use these validator updates if provided, the module manager assumes
		// only one module will update the validator set
		if len(moduleValUpdates) > 0 {
			if len(validatorUpdates) > 0 {
				return abci.ResponseInitChain{}, errors.New("validator InitGenesis updates already set by a previous module")
			}
			validatorUpdates = moduleValUpdates
		}
	}

	return abci.ResponseInitChain{
		Validators: validatorUpdates,
	}, nil
}

// ExportGenesis performs export genesis functionality for modules.
func (mm *Manager) ExportGenesis(ctx sdk.Context) map[string]json.RawMessage {
	genesisData, err := exportGenesis(ctx, mm.cdc, mm.exportGenesisHandlers)
	if err != nil {
		panic(err)
	}

	return genesisData
}

func exportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, exportGenesisHandlers map[string]module.ExportGenesisHandler) (map[string]json.RawMessage, error) {
	var err error
	genesisData := make(map[string]json.RawMessage)
	for name, exportGenesisHandler := range exportGenesisHandlers {
		if exportGenesisHandler == nil {
			continue
		}
		genesisData[name], err = exportGenesisHandler(types.Context{Context: ctx}, cdc)
		if err != nil {
			return genesisData, err
		}
	}

	return genesisData, nil
}

type RegisterInvariantsHandler func(ir sdk.InvariantRegistry)

type configurator struct {
	msgServer                 gogogrpc.Server
	queryServer               gogogrpc.Server
	key                       *rootModuleKey
	cdc                       codec.Marshaler
	requiredServices          map[reflect.Type]bool
	router                    sdk.Router
	initGenesisHandler        module.InitGenesisHandler
	exportGenesisHandler      module.ExportGenesisHandler
	registerInvariantsHandler RegisterInvariantsHandler
}

var _ Configurator = &configurator{}

func (c *configurator) MsgServer() gogogrpc.Server {
	return c.msgServer
}

func (c *configurator) QueryServer() gogogrpc.Server {
	return c.queryServer
}

func (c *configurator) RegisterInvariantsHandler(registry RegisterInvariantsHandler) {
	c.registerInvariantsHandler = registry
}

func (c *configurator) RegisterGenesisHandlers(initGenesisHandler module.InitGenesisHandler, exportGenesisHandler module.ExportGenesisHandler) {
	c.initGenesisHandler = initGenesisHandler
	c.exportGenesisHandler = exportGenesisHandler
}

func (c *configurator) ModuleKey() RootModuleKey {
	return c.key
}

func (c *configurator) Marshaler() codec.Marshaler {
	return c.cdc
}

// Router is temporarily added here to use in the group module.
// TODO: remove once #225 addressed
func (c *configurator) Router() sdk.Router {
	return c.router
}

func (c *configurator) RequireServer(serverInterface interface{}) {
	c.requiredServices[reflect.TypeOf(serverInterface)] = true
}
