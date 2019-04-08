package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
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
	RegisterCodec(*codec.Codec)
	RegisterInvariants(InvariantRouter)

	// routes
	Route() string
	NewHandler() sdk.Handler
	QuerierRoute() string
	NewQuerierHandler() sdk.Querier

	// genesis
	DefaultGenesisState() json.RawMessage
	ValidateGenesis(json.RawMessage) error
	InitGenesis(sdk.Context, json.RawMessage) ([]abci.ValidatorUpdate, error)
	ExportGenesis(sdk.Context) json.RawMessage

	BeginBlock(sdk.Context) error
	EndBlock(sdk.Context) (Tags, error)
}

// module manananaager
type ModuleManager struct {
	Modules            map[string]AppModule
	OrderInitGenesis   []string
	OrderExportGenesis []string
	OrderEndBlockers   []string
	OrderBeginBlockers []string
}

// NewModuleManager creates a new ModuleManager object
func NewModuleManager(modules ...AppModule) ModuleManager {

	moduleMap := make(map[string]AppModule)
	for _, module := range modules {
		moduleMap[module.Name()] = module
	}

	return ModuleManager{
		Modules:            moduleMap,
		OrderInitGenesis:   []string{},
		OrderExportGenesis: []string{},
		OrderEndBlockers:   []string{},
		OrderBeginBlockers: []string{},
	}
}

// set the order of init genesis calls
func (mm ModuleManager) SetOrderInitGenesis(moduleNames ...string) {
	mm.OrderInitGenesis = moduleNames
}

// set the order of export genesis calls
func (mm ModuleManager) SetOrderExportGenesis(moduleNames ...string) {
	mm.OrderExportGenesis = moduleNames
}

// set the order of set end-blocker calls
func (mm ModuleManager) SetOrderEndBlockers(moduleNames ...string) {
	mm.OrderEndBlockers = moduleNames
}

// set the order of set begin-blocker calls
func (mm ModuleManager) SetOrderBeginBlockers(moduleNames ...string) {
	mm.OrderBeginBlockers = moduleNames
}

// register all module codecs
func (mm ModuleManager) RegisterCodecs(*codec.Codec) {
	for _, module := range mm.Modules {
		module.RegisterCodec(*codec.Codec)
	}
}

// register all module routes and module querier routes
func (mm ModuleManager) RegisterInvariants(invarRouter InvariantRouter) {
	for _, module := range mm.Modules {
		module.RegisterInvariants(ck)
	}
}

// register all module routes and module querier routes
func (mm ModuleManager) RegisterRoutes(router baseapp.Router, querierRouter baseapp.QuerierRouter) {
	for _, module := range mm.Modules {
		router.AddRoute(module.Route(), module.NewHandler())
		querierRouter.AddRoute(module.QuerierRoute(), module.NewQuerierHandler())
	}
}

// validate all genesis information
func (mm ModuleManager) ValidateGenesis(genesisData map[string]json.RawMessage) error {
	for _, module := range mm.Modules {
		err := module.ValidateGenesis(genesisDate[module.Name()])
		if err != nil {
			return err
		}
	}
	return nil
}

// default genesis state for modules
func (mm ModuleManager) DefaultGenesisState() map[string]json.RawMessage {
	defaultGenesisState := make(map[string]json.RawMessage)
	for _, module := range mm.Modules {
		defaultGenesisState[module.Name()] = module.DefaultGenesisState()
	}
	return defaultGenesisState
}

func (mm ModuleManager) moduleNames() (names []string) {
	for _, module := range mm.Modules {
		names = append(names, module.Name())
	}
	return names
}

// perform init genesis functionality for modules
func (mm ModuleManager) InitGenesis(ctx sdk.Context, genesisData map[string]json.RawMessage) ([]abci.ValidatorUpdate, error) {
	var moduleNames []string
	if len(OrderInitGenesis) > 0 {
		moduleNames = OrderInitGenesis
	} else {
		moduleNames = moduleNames()
	}

	var validatorUpdates []abci.ValidatorUpdate
	for _, moduleName := range moduleNames {
		moduleValUpdates, err := mm.Modules[moduleName].InitGenesis(ctx, genesisDate[module.Name()])
		if err != nil {
			return []abci.ValidatorUpdate{}, err
		}

		// overwrite validator updates if provided
		if len(moduleValUpdates) > 0 {
			validatorUpdates = moduleValUpdates
		}
	}
	return validatorUpdates, nil
}

// perform export genesis functionality for modules
func (mm ModuleManager) ExportGenesis(ctx sdk.Context) (genesisData map[string]json.RawMessage) {
	var moduleNames []string
	if len(OrderExportGenesis) > 0 {
		moduleNames = OrderExportGenesis
	} else {
		moduleNames = moduleNames()
	}

	for _, moduleName := range moduleNames {
		mm.Modules[moduleName].ExportGenesis(ctx)
	}
	return genesisData
}

// perform begin block functionality for modules
func (mm ModuleManager) BeginBlock(ctx sdk.Context) error {
	var moduleNames []string
	if len(OrderBeginBlock) > 0 {
		moduleNames = OrderBeginBlock
	} else {
		moduleNames = moduleNames()
	}

	for _, moduleName := range moduleNames {
		err := mm.Modules[moduleName].BeginBlock(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// perform end block functionality for modules
func (mm ModuleManager) EndBlock(ctx sdk.Context) (Tags, error) {
	var moduleNames []string
	if len(OrderEndBlock) > 0 {
		moduleNames = OrderEndBlock
	} else {
		moduleNames = moduleNames()
	}

	tags := EmptyTags()
	for _, moduleName := range moduleNames {
		moduleTags, err := mm.Modules[moduleName].EndBlock(ctx)
		if err != nil {
			return tags, err
		}
		tags = tags.AppendTags(moduleTags)
	}
	return tags, nil
}

// DONTCOVER
