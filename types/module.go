/*
Package types contains application module patterns and associated "manager" functionality.
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
ModuleBasicManager.
*/
package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ModuleClient helps modules provide a standard interface for exporting client functionality
type ModuleClient interface {
	GetQueryCmd() *cobra.Command
	GetTxCmd() *cobra.Command
}

//__________________________________________________________________________________________
// AppModuleBasic is the standard form for basic non-dependant elements of an application module.
type AppModuleBasic interface {
	Name() string
	RegisterCodec(*codec.Codec)
	DefaultGenesis() json.RawMessage
	ValidateGenesis(json.RawMessage) error
}

// collections of AppModuleBasic
type ModuleBasicManager []AppModuleBasic

func NewModuleBasicManager(modules ...AppModuleBasic) ModuleBasicManager {
	return modules
}

// RegisterCodecs registers all module codecs
func (mbm ModuleBasicManager) RegisterCodec(cdc *codec.Codec) {
	for _, mb := range mbm {
		mb.RegisterCodec(cdc)
	}
}

// Provided default genesis information for all modules
func (mbm ModuleBasicManager) DefaultGenesis() map[string]json.RawMessage {
	genesis := make(map[string]json.RawMessage)
	for _, mb := range mbm {
		genesis[mb.Name()] = mb.DefaultGenesis()
	}
	return genesis
}

// Provided default genesis information for all modules
func (mbm ModuleBasicManager) ValidateGenesis(genesis map[string]json.RawMessage) error {
	for _, mb := range mbm {
		if err := mb.ValidateGenesis(genesis[mb.Name()]); err != nil {
			return err
		}
	}
	return nil
}

//_________________________________________________________
// AppModuleGenesis is the standard form for an application module genesis functions
type AppModuleGenesis interface {
	AppModuleBasic
	InitGenesis(Context, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis(Context) json.RawMessage
}

// AppModule is the standard form for an application module
type AppModule interface {
	AppModuleGenesis

	// registers
	RegisterInvariants(InvariantRouter)

	// routes
	Route() string
	NewHandler() Handler
	QuerierRoute() string
	NewQuerierHandler() Querier

	BeginBlock(Context, abci.RequestBeginBlock) Tags
	EndBlock(Context, abci.RequestEndBlock) ([]abci.ValidatorUpdate, Tags)
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
func (GenesisOnlyAppModule) RegisterInvariants(_ InvariantRouter) {}

// module message route ngame
func (GenesisOnlyAppModule) Route() string { return "" }

// module handler
func (GenesisOnlyAppModule) NewHandler() Handler { return nil }

// module querier route ngame
func (GenesisOnlyAppModule) QuerierRoute() string { return "" }

// module querier
func (gam GenesisOnlyAppModule) NewQuerierHandler() Querier { return nil }

// module begin-block
func (gam GenesisOnlyAppModule) BeginBlock(ctx Context, req abci.RequestBeginBlock) Tags {
	return EmptyTags()
}

// module end-block
func (GenesisOnlyAppModule) EndBlock(_ Context, _ abci.RequestEndBlock) ([]abci.ValidatorUpdate, Tags) {
	return []abci.ValidatorUpdate{}, EmptyTags()
}

//____________________________________________________________________________
// module manager provides the high level utility for managing and executing
// operations for a group of modules
type ModuleManager struct {
	Modules            map[string]AppModule
	OrderInitGenesis   []string
	OrderExportGenesis []string
	OrderBeginBlockers []string
	OrderEndBlockers   []string
}

// NewModuleManager creates a new ModuleManager object
func NewModuleManager(modules ...AppModule) *ModuleManager {

	moduleMap := make(map[string]AppModule)
	var modulesStr []string
	for _, module := range modules {
		moduleMap[module.Name()] = module
		modulesStr = append(modulesStr, module.Name())
	}

	return &ModuleManager{
		Modules:            moduleMap,
		OrderInitGenesis:   modulesStr,
		OrderExportGenesis: modulesStr,
		OrderBeginBlockers: modulesStr,
		OrderEndBlockers:   modulesStr,
	}
}

// set the order of init genesis calls
func (mm *ModuleManager) SetOrderInitGenesis(moduleNames ...string) {
	mm.OrderInitGenesis = moduleNames
}

// set the order of export genesis calls
func (mm *ModuleManager) SetOrderExportGenesis(moduleNames ...string) {
	mm.OrderExportGenesis = moduleNames
}

// set the order of set begin-blocker calls
func (mm *ModuleManager) SetOrderBeginBlockers(moduleNames ...string) {
	mm.OrderBeginBlockers = moduleNames
}

// set the order of set end-blocker calls
func (mm *ModuleManager) SetOrderEndBlockers(moduleNames ...string) {
	mm.OrderEndBlockers = moduleNames
}

// register all module routes and module querier routes
func (mm *ModuleManager) RegisterInvariants(invarRouter InvariantRouter) {
	for _, module := range mm.Modules {
		module.RegisterInvariants(invarRouter)
	}
}

// register all module routes and module querier routes
func (mm *ModuleManager) RegisterRoutes(router Router, queryRouter QueryRouter) {
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
func (mm *ModuleManager) InitGenesis(ctx Context, genesisData map[string]json.RawMessage) abci.ResponseInitChain {
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
func (mm *ModuleManager) ExportGenesis(ctx Context) map[string]json.RawMessage {
	genesisData := make(map[string]json.RawMessage)
	for _, moduleName := range mm.OrderExportGenesis {
		genesisData[moduleName] = mm.Modules[moduleName].ExportGenesis(ctx)
	}
	return genesisData
}

// perform begin block functionality for modules
func (mm *ModuleManager) BeginBlock(ctx Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	tags := EmptyTags()
	for _, moduleName := range mm.OrderBeginBlockers {
		moduleTags := mm.Modules[moduleName].BeginBlock(ctx, req)
		tags = tags.AppendTags(moduleTags)
	}

	return abci.ResponseBeginBlock{
		Tags: tags.ToKVPairs(),
	}
}

// perform end block functionality for modules
func (mm *ModuleManager) EndBlock(ctx Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	validatorUpdates := []abci.ValidatorUpdate{}
	tags := EmptyTags()
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
