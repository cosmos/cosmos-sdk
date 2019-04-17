package types

import (
	"encoding/json"

	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ModuleClients helps modules provide a standard interface for exporting client functionality
type ModuleClients interface {
	GetQueryCmd() *cobra.Command
	GetTxCmd() *cobra.Command
}

// AppModule is the standard form for an application module
type AppModule interface {

	// app name
	Name() string

	// registers
	RegisterInvariants(InvariantRouter)

	// routes
	Route() string
	NewHandler() Handler
	QuerierRoute() string
	NewQuerierHandler() Querier

	// genesis
	InitGenesis(Context, json.RawMessage) []abci.ValidatorUpdate
	ValidateGenesis(json.RawMessage) error
	ExportGenesis(Context) json.RawMessage

	BeginBlock(Context, abci.RequestBeginBlock) Tags
	EndBlock(Context, abci.RequestEndBlock) ([]abci.ValidatorUpdate, Tags)
}

// module manananaager
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

// validate all genesis information
func (mm *ModuleManager) ValidateGenesis(genesisData map[string]json.RawMessage) error {
	for _, module := range mm.Modules {
		err := module.ValidateGenesis(genesisData[module.Name()])
		if err != nil {
			return err
		}
	}
	return nil
}

//// default genesis state for modules
//func (mm *ModuleManager) DefaultGenesisState() map[string]json.RawMessage {
//defaultGenesisState := make(map[string]json.RawMessage)
//for _, module := range mm.Modules {
//defaultGenesisState[module.Name()] = module.DefaultGenesisState()
//}
//return defaultGenesisState
//}

// perform init genesis functionality for modules
func (mm *ModuleManager) InitGenesis(ctx Context, genesisData map[string]json.RawMessage) []abci.ValidatorUpdate {
	moduleNames := mm.OrderInitGenesis

	var validatorUpdates []abci.ValidatorUpdate
	for _, moduleName := range moduleNames {
		moduleValUpdates := mm.Modules[moduleName].InitGenesis(ctx, genesisData[moduleName])

		// overwrite validator updates if provided
		if len(moduleValUpdates) > 0 {
			validatorUpdates = moduleValUpdates
		}
	}
	return validatorUpdates
}

// perform export genesis functionality for modules
func (mm *ModuleManager) ExportGenesis(ctx Context) (genesisData map[string]json.RawMessage) {
	moduleNames := mm.OrderExportGenesis

	for _, moduleName := range moduleNames {
		mm.Modules[moduleName].ExportGenesis(ctx)
	}
	return genesisData
}

// perform begin block functionality for modules
func (mm *ModuleManager) BeginBlock(ctx Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	moduleNames := mm.OrderBeginBlockers

	tags := EmptyTags()
	for _, moduleName := range moduleNames {
		moduleTags := mm.Modules[moduleName].BeginBlock(ctx, req)
		tags = tags.AppendTags(moduleTags)
	}

	return abci.ResponseBeginBlock{
		Tags: tags.ToKVPairs(),
	}
}

// perform end block functionality for modules
func (mm *ModuleManager) EndBlock(ctx Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	moduleNames := mm.OrderEndBlockers

	validatorUpdates := []abci.ValidatorUpdate{}
	tags := EmptyTags()
	for _, moduleName := range moduleNames {
		moduleValUpdates, moduleTags := mm.Modules[moduleName].EndBlock(ctx, req)
		tags = tags.AppendTags(moduleTags)

		// overwrite validator updates if provided
		if len(moduleValUpdates) > 0 {
			validatorUpdates = moduleValUpdates
		}
	}

	return abci.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		Tags:             tags,
	}
}

// DONTCOVER
