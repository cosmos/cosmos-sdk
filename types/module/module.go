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
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AppModuleBasic is the standard form for basic non-dependant elements of an application module.
type AppModuleBasic interface {
	Name() string
	RegisterLegacyAminoCodec(*codec.LegacyAmino)
	RegisterInterfaces(codectypes.InterfaceRegistry)

	DefaultGenesis(codec.JSONCodec) json.RawMessage
	ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error

	// client functionality
	RegisterRESTRoutes(client.Context, *mux.Router)
	RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)
	GetTxCmd() *cobra.Command
	GetQueryCmd() *cobra.Command
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

// RegisterLegacyAminoCodec registers all module codecs
func (bm BasicManager) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	for _, b := range bm {
		b.RegisterLegacyAminoCodec(cdc)
	}
}

// RegisterInterfaces registers all module interface types
func (bm BasicManager) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	for _, m := range bm {
		m.RegisterInterfaces(registry)
	}
}

// DefaultGenesis provides default genesis information for all modules
func (bm BasicManager) DefaultGenesis(cdc codec.JSONCodec) map[string]json.RawMessage {
	genesis := make(map[string]json.RawMessage)
	for _, b := range bm {
		genesis[b.Name()] = b.DefaultGenesis(cdc)
	}

	return genesis
}

// ValidateGenesis performs genesis state validation for all modules
func (bm BasicManager) ValidateGenesis(cdc codec.JSONCodec, txEncCfg client.TxEncodingConfig, genesis map[string]json.RawMessage) error {
	for _, b := range bm {
		if err := b.ValidateGenesis(cdc, txEncCfg, genesis[b.Name()]); err != nil {
			return err
		}
	}

	return nil
}

// RegisterRESTRoutes registers all module rest routes
func (bm BasicManager) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	for _, b := range bm {
		b.RegisterRESTRoutes(clientCtx, rtr)
	}
}

// RegisterGRPCGatewayRoutes registers all module rest routes
func (bm BasicManager) RegisterGRPCGatewayRoutes(clientCtx client.Context, rtr *runtime.ServeMux) {
	for _, b := range bm {
		b.RegisterGRPCGatewayRoutes(clientCtx, rtr)
	}
}

// AddTxCommands adds all tx commands to the rootTxCmd.
//
// TODO: Remove clientCtx argument.
// REF: https://github.com/cosmos/cosmos-sdk/issues/6571
func (bm BasicManager) AddTxCommands(rootTxCmd *cobra.Command) {
	for _, b := range bm {
		if cmd := b.GetTxCmd(); cmd != nil {
			rootTxCmd.AddCommand(cmd)
		}
	}
}

// AddQueryCommands adds all query commands to the rootQueryCmd.
//
// TODO: Remove clientCtx argument.
// REF: https://github.com/cosmos/cosmos-sdk/issues/6571
func (bm BasicManager) AddQueryCommands(rootQueryCmd *cobra.Command) {
	for _, b := range bm {
		if cmd := b.GetQueryCmd(); cmd != nil {
			rootQueryCmd.AddCommand(cmd)
		}
	}
}

// AppModuleGenesis is the standard form for an application module genesis functions
type AppModuleGenesis interface {
	AppModuleBasic

	InitGenesis(sdk.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis(sdk.Context, codec.JSONCodec) json.RawMessage
}

// AppModule is the standard form for an application module
type AppModule interface {
	AppModuleGenesis

	// registers
	RegisterInvariants(sdk.InvariantRegistry)

	// Deprecated: use RegisterServices
	Route() sdk.Route

	// Deprecated: use RegisterServices
	QuerierRoute() string

	// Deprecated: use RegisterServices
	LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier

	// RegisterServices allows a module to register services
	RegisterServices(Configurator)

	// ConsensusVersion is a sequence number for state-breaking change of the
	// module. It should be incremented on each consensus-breaking change
	// introduced by the module. To avoid wrong/empty versions, the initial version
	// should be set to 1.
	ConsensusVersion() uint64

	// ABCI
	BeginBlock(sdk.Context, abci.RequestBeginBlock)
	EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
}

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
func (GenesisOnlyAppModule) Route() sdk.Route { return sdk.Route{} }

// QuerierRoute returns an empty module querier route
func (GenesisOnlyAppModule) QuerierRoute() string { return "" }

// LegacyQuerierHandler returns an empty module querier
func (gam GenesisOnlyAppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier { return nil }

// RegisterServices registers all services.
func (gam GenesisOnlyAppModule) RegisterServices(Configurator) {}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (gam GenesisOnlyAppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock returns an empty module begin-block
func (gam GenesisOnlyAppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {}

// EndBlock returns an empty module end-block
func (GenesisOnlyAppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

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

// RegisterInvariants registers all module invariants
func (m *Manager) RegisterInvariants(ir sdk.InvariantRegistry) {
	for _, module := range m.Modules {
		module.RegisterInvariants(ir)
	}
}

// RegisterRoutes registers all module routes and module querier routes
func (m *Manager) RegisterRoutes(router sdk.Router, queryRouter sdk.QueryRouter, legacyQuerierCdc *codec.LegacyAmino) {
	for _, module := range m.Modules {
		if r := module.Route(); !r.Empty() {
			router.AddRoute(r)
		}
		if r := module.QuerierRoute(); r != "" {
			queryRouter.AddRoute(r, module.LegacyQuerierHandler(legacyQuerierCdc))
		}
	}
}

// RegisterServices registers all module services
func (m *Manager) RegisterServices(cfg Configurator) {
	for _, module := range m.Modules {
		module.RegisterServices(cfg)
	}
}

// InitGenesis performs init genesis functionality for modules
func (m *Manager) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, genesisData map[string]json.RawMessage) abci.ResponseInitChain {
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
func (m *Manager) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) map[string]json.RawMessage {
	genesisData := make(map[string]json.RawMessage)
	for _, moduleName := range m.OrderExportGenesis {
		genesisData[moduleName] = m.Modules[moduleName].ExportGenesis(ctx, cdc)
	}

	return genesisData
}

// MigrationHandler is the migration function that each module registers.
type MigrationHandler func(sdk.Context) error

// VersionMap is a map of moduleName -> version, where version denotes the
// version from which we should perform the migration for each module.
type VersionMap map[string]uint64

// RunMigrations performs in-place store migrations for all modules. This
// function MUST be called insde an x/upgrade UpgradeHandler.
//
// Recall that in an upgrade handler, the `fromVM` VersionMap is retrieved from
// x/upgrade's store, and the function needs to return the target VersionMap
// that will in turn be persisted to the x/upgrade's store. In general,
// returning RunMigrations should be enough:
//
// Example:
//   cfg := module.NewConfigurator(...)
//   app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
//       return app.mm.RunMigrations(ctx, cfg, fromVM)
//   })
//
// Internally, RunMigrations will perform the following steps:
// - create an `updatedVM` VersionMap of module with their latest ConsensusVersion
// - make a diff of `fromVM` and `udpatedVM`, and for each module:
//    - if the module's `fromVM` version is less than its `updatedVM` version,
//      then run in-place store migrations for that module between those versions.
//    - if the module does not exist in the `fromVM` (which means that it's a new module,
//      because it was not in the previous x/upgrade's store), then run
//      `InitGenesis` on that module.
// - return the `updatedVM` to be persisted in the x/upgrade's store.
//
// As an app developer, if you wish to skip running InitGenesis for your new
// module "foo", you need to manually pass a `fromVM` argument to this function
// foo's module version set to its latest ConsensusVersion. That way, the diff
// between the function's `fromVM` and `udpatedVM` will be empty, hence not
// running anything for foo.
//
// Example:
//   cfg := module.NewConfigurator(...)
//   app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
//       // Assume "foo" is a new module.
//       // `fromVM` is fetched from existing x/upgrade store. Since foo didn't exist
//       // before this upgrade, `v, exists := fromVM["foo"]; exists == false`, and RunMigration will by default
//       // run InitGenesis on foo.
//       // To skip running foo's InitGenesis, you need set `fromVM`'s foo to its latest
//       // consensus version:
//       fromVM["foo"] = foo.AppModule{}.ConsensusVersion()
//
//       return app.mm.RunMigrations(ctx, cfg, fromVM)
//   })
//
// Please also refer to docs/core/upgrade.md for more information.
func (m Manager) RunMigrations(ctx sdk.Context, cfg Configurator, fromVM VersionMap) (VersionMap, error) {
	c, ok := cfg.(configurator)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", configurator{}, cfg)
	}

	updatedVM := make(VersionMap)
	for moduleName, module := range m.Modules {
		fromVersion, exists := fromVM[moduleName]
		toVersion := module.ConsensusVersion()

		// Only run migrations when the module exists in the fromVM.
		// Run InitGenesis otherwise.
		//
		// the module won't exist in the fromVM in two cases:
		// 1. A new module is added. In this case we run InitGenesis with an
		// empty genesis state.
		// 2. An existing chain is upgrading to v043 for the first time. In this case,
		// all modules have yet to be added to x/upgrade's VersionMap store.
		if exists {
			err := c.runModuleMigrations(ctx, moduleName, fromVersion, toVersion)
			if err != nil {
				return nil, err
			}
		} else {
			cfgtor, ok := cfg.(configurator)
			if !ok {
				// Currently, the only implementator of Configurator (the interface)
				// is configurator (the struct).
				return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", configurator{}, cfg)
			}

			moduleValUpdates := module.InitGenesis(ctx, cfgtor.cdc, module.DefaultGenesis(cfgtor.cdc))
			// The module manager assumes only one module will update the
			// validator set, and that it will not be by a new module.
			if len(moduleValUpdates) > 0 {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic, "validator InitGenesis updates already set by a previous module")
			}
		}

		updatedVM[moduleName] = toVersion
	}

	return updatedVM, nil
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

// GetVersionMap gets consensus version from all modules
func (m *Manager) GetVersionMap() VersionMap {
	vermap := make(VersionMap)
	for _, v := range m.Modules {
		version := v.ConsensusVersion()
		name := v.Name()
		vermap[name] = version
	}

	return vermap
}
