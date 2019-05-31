/*
Package module contains application module patterns and associated "manager" functionality.
The module pattern has been broken down by:
 - independent module functionality (AppModuleBasic)
 - inter-dependent module functionality (AppModule)

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

	// genesis
	DefaultGenesis() json.RawMessage
	ValidateGenesis(json.RawMessage) error

	// client functionality
	RegisterRESTRoutes(context.CLIContext, *mux.Router, *codec.Codec)
	GetTxCmd() *cobra.Command
	GetQueryCmd() *cobra.Command
}

// collections of AppModuleBasic
type BasicManager []AppModuleBasic

func NewModuleBasicManager(modules ...AppModuleBasic) BasicManager {
	return modules
}

// RegisterCodecs registers all module codecs
func (mbm BasicManager) RegisterCodec(cdc *codec.Codec) {
	for _, mb := range mbm {
		mb.RegisterCodec(cdc)
	}
}

// Provided default genesis information for all modules
func (mbm BasicManager) DefaultGenesis() map[string]json.RawMessage {
	genesis := make(map[string]json.RawMessage)
	for _, mb := range mbm {
		genesis[mb.Name()] = mb.DefaultGenesis()
	}
	return genesis
}

// Provided default genesis information for all modules
func (mbm BasicManager) ValidateGenesis(genesis map[string]json.RawMessage) error {
	for _, mb := range mbm {
		if err := mb.ValidateGenesis(genesis[mb.Name()]); err != nil {
			return err
		}
	}
	return nil
}

// RegisterRestRoutes registers all module rest routes
func (mbm BasicManager) RegisterRESTRoutes(
	ctx context.CLIContext, rtr *mux.Router, cdc *codec.Codec) {

	for _, mb := range mbm {
		mb.RegisterRESTRoutes(ctx, rtr, cdc)
	}
}

// add all tx commands to the rootTxCmd
func (mbm BasicManager) AddTxCommands(rootTxCmd *cobra.Command) {
	for _, mb := range mbm {
		if cmd := mb.GetTxCmd(); cmd != nil {
			rootTxCmd.AddCommand(cmd)
		}
	}
}

// add all query commands to the rootQueryCmd
func (mbm BasicManager) AddQueryCommands(rootQueryCmd *cobra.Command) {
	for _, mb := range mbm {
		if cmd := mb.GetQueryCmd(); cmd != nil {
			rootQueryCmd.AddCommand(cmd)
		}
	}
}

//_________________________________________________________
// AppModuleGenesis is the standard form for an application module genesis functions
type AppModuleGenesis interface {
	AppModuleBasic
	InitGenesis(sdk.Context, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis(sdk.Context) json.RawMessage
}

// AppModule is the standard form for an application module
type AppModule interface {
	AppModuleGenesis

	// registers
	RegisterInvariants(sdk.InvariantRouter)

	// routes
	Route() string
	NewHandler() sdk.Handler
	QuerierRoute() string
	NewQuerierHandler() sdk.Querier

	BeginBlock(sdk.Context, abci.RequestBeginBlock) sdk.Tags
	EndBlock(sdk.Context, abci.RequestEndBlock) ([]abci.ValidatorUpdate, sdk.Tags)
}

//___________________________
// app module
type GenesisOnlyAppModule struct {
	AppModuleGenesis
}

// NewGenesisOnlyAppModule creates a new GenesisOnlyAppModule object
func NewGenesisOnlyAppModule(amg AppModuleGenesis) AppModule {
	return GenesisOnlyAppModule{
		AppModuleGenesis: amg,
	}
}

// register invariants
func (GenesisOnlyAppModule) RegisterInvariants(_ sdk.InvariantRouter) {}

// module message route ngame
func (GenesisOnlyAppModule) Route() string { return "" }

// module handler
func (GenesisOnlyAppModule) NewHandler() sdk.Handler { return nil }

// module querier route ngame
func (GenesisOnlyAppModule) QuerierRoute() string { return "" }

// module querier
func (gam GenesisOnlyAppModule) NewQuerierHandler() sdk.Querier { return nil }

// module begin-block
func (gam GenesisOnlyAppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) sdk.Tags {
	return sdk.EmptyTags()
}

// module end-block
func (GenesisOnlyAppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) ([]abci.ValidatorUpdate, sdk.Tags) {
	return []abci.ValidatorUpdate{}, sdk.EmptyTags()
}

//____________________________________________________________________________
// module manager provides the high level utility for managing and executing
// operations for a group of modules
type Manager struct {
	Modules            map[string]AppModule
	OrderInitGenesis   []string
	OrderExportGenesis []string
	OrderBeginBlockers []string
	OrderEndBlockers   []string
}

// NewModuleManager creates a new Manager object
func NewModuleManager(modules ...AppModule) *Manager {

	moduleMap := make(map[string]AppModule)
	var modulesStr []string
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

// set the order of init genesis calls
func (mm *Manager) SetOrderInitGenesis(moduleNames ...string) {
	mm.OrderInitGenesis = moduleNames
}

// set the order of export genesis calls
func (mm *Manager) SetOrderExportGenesis(moduleNames ...string) {
	mm.OrderExportGenesis = moduleNames
}

// set the order of set begin-blocker calls
func (mm *Manager) SetOrderBeginBlockers(moduleNames ...string) {
	mm.OrderBeginBlockers = moduleNames
}

// set the order of set end-blocker calls
func (mm *Manager) SetOrderEndBlockers(moduleNames ...string) {
	mm.OrderEndBlockers = moduleNames
}

// register all module routes and module querier routes
func (mm *Manager) RegisterInvariants(invarRouter sdk.InvariantRouter) {
	for _, module := range mm.Modules {
		module.RegisterInvariants(invarRouter)
	}
}

// register all module routes and module querier routes
func (mm *Manager) RegisterRoutes(router sdk.Router, queryRouter sdk.QueryRouter) {
	for _, module := range mm.Modules {
		if module.Route() != "" {
			router.AddRoute(module.Route(), module.NewHandler())
		}
		if module.QuerierRoute() != "" {
			queryRouter.AddRoute(module.QuerierRoute(), module.NewQuerierHandler())
		}
	}
}

// perform init genesis functionality for modules
func (mm *Manager) InitGenesis(ctx sdk.Context, genesisData map[string]json.RawMessage) abci.ResponseInitChain {
	var validatorUpdates []abci.ValidatorUpdate
	for _, moduleName := range mm.OrderInitGenesis {
		if genesisData[moduleName] == nil {
			continue
		}
		moduleValUpdates := mm.Modules[moduleName].InitGenesis(ctx, genesisData[moduleName])

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

// perform export genesis functionality for modules
func (mm *Manager) ExportGenesis(ctx sdk.Context) map[string]json.RawMessage {
	genesisData := make(map[string]json.RawMessage)
	for _, moduleName := range mm.OrderExportGenesis {
		genesisData[moduleName] = mm.Modules[moduleName].ExportGenesis(ctx)
	}
	return genesisData
}

// perform begin block functionality for modules
func (mm *Manager) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	tags := sdk.EmptyTags()
	for _, moduleName := range mm.OrderBeginBlockers {
		moduleTags := mm.Modules[moduleName].BeginBlock(ctx, req)
		tags = tags.AppendTags(moduleTags)
	}

	return abci.ResponseBeginBlock{
		Tags: tags.ToKVPairs(),
	}
}

// perform end block functionality for modules
func (mm *Manager) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	validatorUpdates := []abci.ValidatorUpdate{}
	tags := sdk.EmptyTags()
	for _, moduleName := range mm.OrderEndBlockers {
		moduleValUpdates, moduleTags := mm.Modules[moduleName].EndBlock(ctx, req)
		tags = tags.AppendTags(moduleTags)

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
		Tags:             tags,
	}
}

// DONTCOVER
