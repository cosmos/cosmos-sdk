/*
Package module contains application module patterns and associated "manager" functionality.
The module pattern has been broken down by:
 - independent module functionality (AppModuleBasic)
 - inter-dependent module genesis functionality (AppModuleGenesis)
 - inter-dependent module simulation functionality (AppModuleSimulation)
 - inter-dependent module full functionality (AppModule)

inter-dependent module functionality is module functionality which somehow
depends on other modules, typically through the module keeper.  Many of the
module keepers are dependent on each other, thus in order to access the full
set of module functionality we need to define all the keepers/params-store/keys
etc. This full set of advanced functionality is defined by the AppModule interface.

Independent module functions are separated to allow for the construction of the
basic application structures required early on in the application definition
and used to enable the definition of full module functionality later in the
process. This separation is necessary, however we still want to allow for a
high level pattern for modules to follow - for instance, such that we don't
have to manually register all of the codecs for all the modules. This basic
procedure as well as other basic patterns are handled through the use of
BasicManager.

Lastly the interface for genesis functionality (AppModuleGenesis) has been
separated out from full module functionality (AppModule) so that modules which
are only used for genesis can take advantage of the Module patterns without
needlessly defining many placeholder functions
*/
package module

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//__________________________________________________________________________________________

// AppModuleBasic is the standard form for basic non-dependant elements of an application module.
type AppModuleBasic interface {
	Name() string
	RegisterCodec(*codec.Codec)

	DefaultGenesis(codec.JSONMarshaler) json.RawMessage
	ValidateGenesis(codec.JSONMarshaler, json.RawMessage) error

	// client functionality
	RegisterRESTRoutes(context.CLIContext, *mux.Router)
	GetTxCmd(*codec.Codec) *cobra.Command
	GetQueryCmd(*codec.Codec) *cobra.Command
}

// BasicManager is a collection of AppModuleBasic
type BasicManager map[string]AppModuleBasic

// NewBasicManager creates a new BasicManager object
func NewBasicManager(modules ...AppModuleBasic) BasicManager {
	moduleMap := make(map[string]AppModuleBasic)
	for _, module := range modules {
		moduleMap[module.Name()] = module
	}
	return moduleMap
}

// RegisterCodec registers all module codecs
func (bm BasicManager) RegisterCodec(cdc *codec.Codec) {
	for _, b := range bm {
		b.RegisterCodec(cdc)
	}
}

// DefaultGenesis provides default genesis information for all modules
func (bm BasicManager) DefaultGenesis(cdc codec.JSONMarshaler) map[string]json.RawMessage {
	genesis := make(map[string]json.RawMessage)
	for _, b := range bm {
		genesis[b.Name()] = b.DefaultGenesis(cdc)
	}

	return genesis
}

// ValidateGenesis performs genesis state validation for all modules
func (bm BasicManager) ValidateGenesis(cdc codec.JSONMarshaler, genesis map[string]json.RawMessage) error {
	for _, b := range bm {
		if err := b.ValidateGenesis(cdc, genesis[b.Name()]); err != nil {
			return err
		}
	}

	return nil
}

// RegisterRESTRoutes registers all module rest routes
func (bm BasicManager) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	for _, b := range bm {
		b.RegisterRESTRoutes(ctx, rtr)
	}
}

// AddTxCommands adds all tx commands to the rootTxCmd
func (bm BasicManager) AddTxCommands(rootTxCmd *cobra.Command, cdc *codec.Codec) {
	for _, b := range bm {
		if cmd := b.GetTxCmd(cdc); cmd != nil {
			rootTxCmd.AddCommand(cmd)
		}
	}
}

// AddQueryCommands adds all query commands to the rootQueryCmd
func (bm BasicManager) AddQueryCommands(rootQueryCmd *cobra.Command, cdc *codec.Codec) {
	for _, b := range bm {
		if cmd := b.GetQueryCmd(cdc); cmd != nil {
			rootQueryCmd.AddCommand(cmd)
		}
	}
}

//_________________________________________________________

// AppModuleGenesis is the standard form for an application module genesis functions
type AppModuleGenesis interface {
	AppModuleBasic

	InitGenesis(sdk.Context, codec.JSONMarshaler, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis(sdk.Context, codec.JSONMarshaler) json.RawMessage
}

// AppModule is the standard form for an application module
type AppModule interface {
	AppModuleGenesis

	// registers
	RegisterInvariants(sdk.InvariantRegistry)

	// routes
	Route() string
	NewHandler() sdk.Handler
	QuerierRoute() string
	NewQuerierHandler() sdk.Querier

	// ABCI
	BeginBlock(sdk.Context, abci.RequestBeginBlock)
	EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
}

//___________________________

// GenesisOnlyAppModule is an AppModule that only has import/export functionality
type GenesisOnlyAppModule struct {
	AppModuleGenesis
}

// NewGenesisOnlyAppModule creates a new GenesisOnlyAppModule object
func NewGenesisOnlyAppModule(amg AppModuleGenesis) AppModule {
	return GenesisOnlyAppModule{
		AppModuleGenesis: amg,
	}
}

// RegisterInvariants is a placeholder function register no invariants
func (GenesisOnlyAppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route empty module message route
func (GenesisOnlyAppModule) Route() string { return "" }

// NewHandler returns an empty module handler
func (GenesisOnlyAppModule) NewHandler() sdk.Handler { return nil }

// QuerierRoute returns an empty module querier route
func (GenesisOnlyAppModule) QuerierRoute() string { return "" }

// NewQuerierHandler returns an empty module querier
func (gam GenesisOnlyAppModule) NewQuerierHandler() sdk.Querier { return nil }

// BeginBlock returns an empty module begin-block
func (gam GenesisOnlyAppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {}

// EndBlock returns an empty module end-block
func (GenesisOnlyAppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

//____________________________________________________________________________

// Manager defines a module manager that provides the high level utility for managing and executing
// operations for a group of modules
type Manager struct {
	Modules            map[string]AppModule
	OrderInitGenesis   []string
	OrderExportGenesis []string
	OrderBeginBlockers []string
	OrderEndBlockers   []string
}

// NewManager creates a new Manager object
func NewManager(modules ...AppModule) *Manager {

	moduleMap := make(map[string]AppModule)
	modulesStr := make([]string, 0, len(modules))
	for _, module := range modules {
		moduleMap[module.Name()] = module
		modulesStr = append(modulesStr, module.Name())
	}

	return &Manager{
		Modules:            moduleMap,
		OrderInitGenesis:   modulesStr,
		OrderExportGenesis: modulesStr,
		OrderBeginBlockers: modulesStr,
		OrderEndBlockers:   modulesStr,
	}
}

// SetOrderInitGenesis sets the order of init genesis calls
func (m *Manager) SetOrderInitGenesis(moduleNames ...string) {
	m.OrderInitGenesis = moduleNames
}

// SetOrderExportGenesis sets the order of export genesis calls
func (m *Manager) SetOrderExportGenesis(moduleNames ...string) {
	m.OrderExportGenesis = moduleNames
}

// SetOrderBeginBlockers sets the order of set begin-blocker calls
func (m *Manager) SetOrderBeginBlockers(moduleNames ...string) {
	m.OrderBeginBlockers = moduleNames
}

// SetOrderEndBlockers sets the order of set end-blocker calls
func (m *Manager) SetOrderEndBlockers(moduleNames ...string) {
	m.OrderEndBlockers = moduleNames
}

// RegisterInvariants registers all module routes and module querier routes
func (m *Manager) RegisterInvariants(ir sdk.InvariantRegistry) {
	for _, module := range m.Modules {
		module.RegisterInvariants(ir)
	}
}

// RegisterRoutes registers all module routes and module querier routes
func (m *Manager) RegisterRoutes(router sdk.Router, queryRouter sdk.QueryRouter) {
	for _, module := range m.Modules {
		if module.Route() != "" {
			router.AddRoute(module.Route(), module.NewHandler())
		}
		if module.QuerierRoute() != "" {
			queryRouter.AddRoute(module.QuerierRoute(), module.NewQuerierHandler())
		}
	}
}

// InitGenesis performs init genesis functionality for modules
func (m *Manager) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, genesisData map[string]json.RawMessage) abci.ResponseInitChain {
	var validatorUpdates []abci.ValidatorUpdate
	for _, moduleName := range m.OrderInitGenesis {
		if genesisData[moduleName] == nil {
			continue
		}

		moduleValUpdates := m.Modules[moduleName].InitGenesis(ctx, cdc, genesisData[moduleName])

		// use these validator updates if provided, the module manager assumes
		// only one module will update the validator set
		if len(moduleValUpdates) > 0 {
			if len(validatorUpdates) > 0 {
				panic("validator InitGenesis updates already set by a previous module")
			}
			validatorUpdates = moduleValUpdates
		}
	}

	return abci.ResponseInitChain{
		Validators: validatorUpdates,
	}
}

// ExportGenesis performs export genesis functionality for modules
func (m *Manager) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) map[string]json.RawMessage {
	genesisData := make(map[string]json.RawMessage)
	for _, moduleName := range m.OrderExportGenesis {
		genesisData[moduleName] = m.Modules[moduleName].ExportGenesis(ctx, cdc)
	}

	return genesisData
}

// BeginBlock performs begin block functionality for all modules. It creates a
// child context with an event manager to aggregate events emitted from all
// modules.
func (m *Manager) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	for _, moduleName := range m.OrderBeginBlockers {
		m.Modules[moduleName].BeginBlock(ctx, req)
	}

	return abci.ResponseBeginBlock{
		Events: ctx.EventManager().ABCIEvents(),
	}
}

// EndBlock performs end block functionality for all modules. It creates a
// child context with an event manager to aggregate events emitted from all
// modules.
func (m *Manager) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	validatorUpdates := []abci.ValidatorUpdate{}

	for _, moduleName := range m.OrderEndBlockers {
		moduleValUpdates := m.Modules[moduleName].EndBlock(ctx, req)

		// use these validator updates if provided, the module manager assumes
		// only one module will update the validator set
		if len(moduleValUpdates) > 0 {
			if len(validatorUpdates) > 0 {
				panic("validator EndBlock updates already set by a previous module")
			}

			validatorUpdates = moduleValUpdates
		}
	}

	return abci.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		Events:           ctx.EventManager().ABCIEvents(),
	}
}
