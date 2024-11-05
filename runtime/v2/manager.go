package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	cosmosmsg "cosmossdk.io/api/cosmos/msg/v1"
	"cosmossdk.io/core/appmodule"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/server/v2/stf"
)

type MM[T transaction.Tx] struct {
	logger             log.Logger
	config             *runtimev2.Module
	modules            map[string]appmodulev2.AppModule
	migrationRegistrar *migrationRegistrar
}

// NewModuleManager is the constructor for the module manager
// It handles all the interactions between the modules and the application
func NewModuleManager[T transaction.Tx](
	logger log.Logger,
	config *runtimev2.Module,
	modules map[string]appmodulev2.AppModule,
) *MM[T] {
	// good defaults for the module manager order
	modulesName := slices.Sorted(maps.Keys(modules))
	if len(config.PreBlockers) == 0 {
		config.PreBlockers = modulesName
	}
	if len(config.BeginBlockers) == 0 {
		config.BeginBlockers = modulesName
	}
	if len(config.EndBlockers) == 0 {
		config.EndBlockers = modulesName
	}
	if len(config.TxValidators) == 0 {
		config.TxValidators = modulesName
	}
	if len(config.InitGenesis) == 0 {
		config.InitGenesis = modulesName
	}
	if len(config.ExportGenesis) == 0 {
		config.ExportGenesis = modulesName
	}
	if len(config.OrderMigrations) == 0 {
		config.OrderMigrations = defaultMigrationsOrder(modulesName)
	}

	mm := &MM[T]{
		logger:             logger,
		config:             config,
		modules:            modules,
		migrationRegistrar: newMigrationRegistrar(),
	}

	if err := mm.validateConfig(); err != nil {
		panic(err)
	}

	return mm
}

// Modules returns the modules registered in the module manager
func (m *MM[T]) Modules() map[string]appmodulev2.AppModule {
	return m.modules
}

// RegisterLegacyAminoCodec registers all module codecs
func (m *MM[T]) RegisterLegacyAminoCodec(registrar registry.AminoRegistrar) {
	for _, b := range m.modules {
		if mod, ok := b.(appmodule.HasAminoCodec); ok {
			mod.RegisterLegacyAminoCodec(registrar)
		}
	}
}

// RegisterInterfaces registers all module interface types
func (m *MM[T]) RegisterInterfaces(registry registry.InterfaceRegistrar) {
	for _, b := range m.modules {
		if mod, ok := b.(appmodulev2.HasRegisterInterfaces); ok {
			mod.RegisterInterfaces(registry)
		}
	}
}

// DefaultGenesis provides default genesis information for all modules
func (m *MM[T]) DefaultGenesis() map[string]json.RawMessage {
	genesisData := make(map[string]json.RawMessage)
	for name, b := range m.modules {
		if mod, ok := b.(appmodule.HasGenesisBasics); ok {
			genesisData[name] = mod.DefaultGenesis()
		} else if mod, ok := b.(appmodulev2.HasGenesis); ok {
			genesisData[name] = mod.DefaultGenesis()
		} else {
			genesisData[name] = []byte("{}")
		}
	}

	return genesisData
}

// ValidateGenesis performs genesis state validation for all modules
func (m *MM[T]) ValidateGenesis(genesisData map[string]json.RawMessage) error {
	for name, b := range m.modules {
		if mod, ok := b.(appmodule.HasGenesisBasics); ok {
			if err := mod.ValidateGenesis(genesisData[name]); err != nil {
				return err
			}
		} else if mod, ok := b.(appmodulev2.HasGenesis); ok {
			if err := mod.ValidateGenesis(genesisData[name]); err != nil {
				return err
			}
		}
	}

	return nil
}

// InitGenesisJSON performs init genesis functionality for modules from genesis data in JSON format
func (m *MM[T]) InitGenesisJSON(
	ctx context.Context,
	genesisData map[string]json.RawMessage,
	txHandler func(json.RawMessage) error,
) error {
	m.logger.Info("initializing blockchain state from genesis.json", "order", m.config.InitGenesis)
	var seenValUpdates bool
	for _, moduleName := range m.config.InitGenesis {
		if genesisData[moduleName] == nil {
			continue
		}

		mod := m.modules[moduleName]

		// we might get an adapted module, a native core API module or a legacy module
		switch module := mod.(type) {
		case appmodule.HasGenesisAuto:
			panic(fmt.Sprintf("module %s isn't server/v2 compatible", moduleName))
		case appmodulev2.GenesisDecoder: // GenesisDecoder needs to supersede HasGenesis and HasABCIGenesis.
			genTxs, err := module.DecodeGenesisJSON(genesisData[moduleName])
			if err != nil {
				return err
			}
			for _, jsonTx := range genTxs {
				if err := txHandler(jsonTx); err != nil {
					return fmt.Errorf("failed to handle genesis transaction: %w", err)
				}
			}
		case appmodulev2.HasGenesis:
			m.logger.Debug("running initialization for module", "module", moduleName)
			if err := module.InitGenesis(ctx, genesisData[moduleName]); err != nil {
				return fmt.Errorf("init module %s: %w", moduleName, err)
			}
		case appmodulev2.HasABCIGenesis:
			m.logger.Debug("running initialization for module", "module", moduleName)
			moduleValUpdates, err := module.InitGenesis(ctx, genesisData[moduleName])
			if err != nil {
				return err
			}

			// use these validator updates if provided, the module manager assumes
			// only one module will update the validator set
			if len(moduleValUpdates) > 0 {
				if seenValUpdates {
					return fmt.Errorf("validator InitGenesis updates already set by a previous module: current module %s", moduleName)
				} else {
					seenValUpdates = true
				}
			}
		}

	}
	return nil
}

// ExportGenesisForModules performs export genesis functionality for modules
func (m *MM[T]) ExportGenesisForModules(
	ctx context.Context,
	stateFactory func() store.WriterMap,
	modulesToExport ...string,
) (map[string]json.RawMessage, error) {
	if len(modulesToExport) == 0 {
		modulesToExport = m.config.ExportGenesis
	}
	// verify modules exists in app, so that we don't panic in the middle of an export
	if err := m.checkModulesExists(modulesToExport); err != nil {
		return nil, err
	}

	type genesisResult struct {
		bz  json.RawMessage
		err error
	}

	type ModuleI interface {
		ExportGenesis(ctx context.Context) (json.RawMessage, error)
	}

	channels := make(map[string]chan genesisResult)
	for _, moduleName := range modulesToExport {
		mod := m.modules[moduleName]
		var moduleI ModuleI
		if module, hasGenesis := mod.(appmodulev2.HasGenesis); hasGenesis {
			moduleI = module.(ModuleI)
		} else if module, hasABCIGenesis := mod.(appmodulev2.HasABCIGenesis); hasABCIGenesis {
			moduleI = module.(ModuleI)
		} else {
			continue
		}

		channels[moduleName] = make(chan genesisResult)
		go func(moduleI ModuleI, ch chan genesisResult) {
			genesisCtx := services.NewGenesisContext(stateFactory())
			err := genesisCtx.Read(ctx, func(ctx context.Context) error {
				jm, err := moduleI.ExportGenesis(ctx)
				if err != nil {
					return err
				}
				ch <- genesisResult{jm, nil}
				return nil
			})
			if err != nil {
				ch <- genesisResult{nil, err}
			}
		}(moduleI, channels[moduleName])
	}

	genesisData := make(map[string]json.RawMessage)
	for moduleName := range channels {
		res := <-channels[moduleName]
		if res.err != nil {
			return nil, fmt.Errorf("genesis export error in %s: %w", moduleName, res.err)
		}

		genesisData[moduleName] = res.bz
	}

	return genesisData, nil
}

// checkModulesExists verifies that all modules in the list exist in the app
func (m *MM[T]) checkModulesExists(moduleName []string) error {
	for _, name := range moduleName {
		if _, ok := m.modules[name]; !ok {
			return fmt.Errorf("module %s does not exist", name)
		}
	}

	return nil
}

// BeginBlock runs the begin-block logic of all modules
func (m *MM[T]) BeginBlock() func(ctx context.Context) error {
	return func(ctx context.Context) error {
		for _, moduleName := range m.config.BeginBlockers {
			if module, ok := m.modules[moduleName].(appmodulev2.HasBeginBlocker); ok {
				if err := module.BeginBlock(ctx); err != nil {
					return fmt.Errorf("failed to run beginblocker for %s: %w", moduleName, err)
				}
			}
		}

		return nil
	}
}

// hasABCIEndBlock is the legacy EndBlocker implemented by x/staking in the CosmosSDK
type hasABCIEndBlock interface {
	EndBlock(context.Context) ([]appmodulev2.ValidatorUpdate, error)
}

// EndBlock runs the end-block logic of all modules and tx validator updates
func (m *MM[T]) EndBlock() (
	endBlockFunc func(ctx context.Context) error,
	valUpdateFunc func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error),
) {
	var validatorUpdates []appmodulev2.ValidatorUpdate
	endBlockFunc = func(ctx context.Context) error {
		for _, moduleName := range m.config.EndBlockers {
			if module, ok := m.modules[moduleName].(appmodulev2.HasEndBlocker); ok {
				err := module.EndBlock(ctx)
				if err != nil {
					return fmt.Errorf("failed to run endblock for %s: %w", moduleName, err)
				}
			} else if module, ok := m.modules[moduleName].(hasABCIEndBlock); ok { // we need to keep this for our module compatibility promise
				moduleValUpdates, err := module.EndBlock(ctx)
				if err != nil {
					return fmt.Errorf("failed to run enblock for %s: %w", moduleName, err)
				}
				// use these validator updates if provided, the module manager assumes
				// only one module will update the validator set
				if len(moduleValUpdates) > 0 {
					if len(validatorUpdates) > 0 {
						return errors.New("validator end block updates already set by a previous module")
					}

					validatorUpdates = append(validatorUpdates, moduleValUpdates...)
				}
			}
		}

		return nil
	}

	valUpdateFunc = func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error) {
		// get validator updates of modules implementing directly the new HasUpdateValidators interface
		for _, v := range m.modules {
			if module, ok := v.(appmodulev2.HasUpdateValidators); ok {
				moduleValUpdates, err := module.UpdateValidators(ctx)
				if err != nil {
					return nil, err
				}

				if len(moduleValUpdates) > 0 {
					if len(validatorUpdates) > 0 {
						return nil, errors.New("validator end block updates already set by a previous module")
					}

					validatorUpdates = append(validatorUpdates, moduleValUpdates...)
				}
			}
		}

		// Reset validatorUpdates
		res := validatorUpdates
		validatorUpdates = []appmodulev2.ValidatorUpdate{}

		return res, nil
	}

	return endBlockFunc, valUpdateFunc
}

// PreBlocker runs the pre-block logic of all modules
func (m *MM[T]) PreBlocker() func(ctx context.Context, txs []T) error {
	return func(ctx context.Context, txs []T) error {
		for _, moduleName := range m.config.PreBlockers {
			if module, ok := m.modules[moduleName].(appmodulev2.HasPreBlocker); ok {
				if err := module.PreBlock(ctx); err != nil {
					return fmt.Errorf("failed to run preblock for %s: %w", moduleName, err)
				}
			}
		}

		return nil
	}
}

// TxValidators validates incoming transactions
func (m *MM[T]) TxValidators() func(ctx context.Context, tx T) error {
	return func(ctx context.Context, tx T) error {
		for _, moduleName := range m.config.TxValidators {
			if module, ok := m.modules[moduleName].(appmodulev2.HasTxValidator[T]); ok {
				if err := module.TxValidator(ctx, tx); err != nil {
					return fmt.Errorf("failed to run tx validator for %s: %w", moduleName, err)
				}
			}
		}

		return nil
	}
}

// RunMigrations performs in-place store migrations for all modules. This
// function MUST be called inside an x/upgrade UpgradeHandler.
//
// Recall that in an upgrade handler, the `fromVM` VersionMap is retrieved from
// x/upgrade's store, and the function needs to return the target VersionMap
// that will in turn be persisted to the x/upgrade's store. In general,
// returning RunMigrations should be enough:
//
// Example:
//
//	app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx context.Context, plan upgradetypes.Plan, fromVM appmodule.VersionMap) (appmodule.VersionMap, error) {
//	    return app.ModuleManager().RunMigrations(ctx, fromVM)
//	})
//
// Internally, RunMigrations will perform the following steps:
//   - create an `updatedVM` VersionMap of module with their latest ConsensusVersion
//   - if module implements `HasConsensusVersion` interface get the consensus version as `toVersion`,
//     if not `toVersion` is set to 0.
//   - get `fromVersion` from `fromVM` with module's name.
//   - if the module's name exists in `fromVM` map, then run in-place store migrations
//     for that module between `fromVersion` and `toVersion`.
//   - if the module does not exist in the `fromVM` (which means that it's a new module,
//     because it was not in the previous x/upgrade's store), then run
//     `InitGenesis` on that module.
//
// - return the `updatedVM` to be persisted in the x/upgrade's store.
//
// Migrations are run in an order defined by `mm.config.OrderMigrations`.
//
// As an app developer, if you wish to skip running InitGenesis for your new
// module "foo", you need to manually pass a `fromVM` argument to this function
// foo's module version set to its latest ConsensusVersion. That way, the diff
// between the function's `fromVM` and `udpatedVM` will be empty, hence not
// running anything for foo.
//
// Example:
//
//	app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
//	    // Assume "foo" is a new module.
//	    // `fromVM` is fetched from existing x/upgrade store. Since foo didn't exist
//	    // before this upgrade, `v, exists := fromVM["foo"]; exists == false`, and RunMigration will by default
//	    // run InitGenesis on foo.
//	    // To skip running foo's InitGenesis, you need set `fromVM`'s foo to its latest
//	    // consensus version:
//	    fromVM["foo"] = foo.AppModule{}.ConsensusVersion()
//
//	    return app.ModuleManager().RunMigrations(ctx, fromVM)
//	})
//
// Please also refer to https://docs.cosmos.network/main/core/upgrade for more information.
func (m *MM[T]) RunMigrations(ctx context.Context, fromVM appmodulev2.VersionMap) (appmodulev2.VersionMap, error) {
	updatedVM := appmodulev2.VersionMap{}
	for _, moduleName := range m.config.OrderMigrations {
		module := m.modules[moduleName]
		fromVersion, exists := fromVM[moduleName]
		toVersion := uint64(0)
		if module, ok := module.(appmodulev2.HasConsensusVersion); ok {
			toVersion = module.ConsensusVersion()
		}

		// We run migration if the module is specified in `fromVM`.
		// Otherwise we run InitGenesis.
		//
		// The module won't exist in the fromVM in two cases:
		// 1. A new module is added. In this case we run InitGenesis with an
		// empty genesis state.
		// 2. An existing chain is upgrading from version < 0.43 to v0.43+ for the first time.
		// In this case, all modules have yet to be added to x/upgrade's VersionMap store.
		if exists {
			m.logger.Info(fmt.Sprintf("migrating module %s from version %d to version %d", moduleName, fromVersion, toVersion))
			if err := m.migrationRegistrar.RunModuleMigrations(ctx, moduleName, fromVersion, toVersion); err != nil {
				return nil, err
			}
		} else {
			m.logger.Info(fmt.Sprintf("adding a new module: %s", moduleName))
			if mod, ok := m.modules[moduleName].(appmodule.HasGenesis); ok {
				if err := mod.InitGenesis(ctx, mod.DefaultGenesis()); err != nil {
					return nil, fmt.Errorf("failed to run InitGenesis for %s: %w", moduleName, err)
				}
			}
			if mod, ok := m.modules[moduleName].(appmodulev2.HasABCIGenesis); ok {
				moduleValUpdates, err := mod.InitGenesis(ctx, mod.DefaultGenesis())
				if err != nil {
					return nil, err
				}

				// The module manager assumes only one module will update the validator set, and it can't be a new module.
				if len(moduleValUpdates) > 0 {
					return nil, errors.New("validator InitGenesis update is already set by another module")
				}
			}
		}

		updatedVM[moduleName] = toVersion
	}

	return updatedVM, nil
}

// RegisterServices registers all module services.
func (m *MM[T]) RegisterServices(app *App[T]) error {
	for _, module := range m.modules {
		// register msg + query
		if err := registerServices(module, app, protoregistry.GlobalFiles); err != nil {
			return err
		}

		// register migrations
		if module, ok := module.(appmodulev2.HasMigrations); ok {
			if err := module.RegisterMigrations(m.migrationRegistrar); err != nil {
				return err
			}
		}

		// register pre and post msg
		if module, ok := module.(appmodulev2.HasPreMsgHandlers); ok {
			module.RegisterPreMsgHandlers(app.msgRouterBuilder)
		}

		if module, ok := module.(appmodulev2.HasPostMsgHandlers); ok {
			module.RegisterPostMsgHandlers(app.msgRouterBuilder)
		}
	}

	return nil
}

// validateConfig validates the module manager configuration
// it asserts that all modules are defined in the configuration and that no modules are forgotten
func (m *MM[T]) validateConfig() error {
	if err := m.assertNoForgottenModules("PreBlockers", m.config.PreBlockers, func(moduleName string) bool {
		module := m.modules[moduleName]
		_, hasPreBlock := module.(appmodulev2.HasPreBlocker)
		return !hasPreBlock
	}); err != nil {
		return err
	}

	if err := m.assertNoForgottenModules("BeginBlockers", m.config.BeginBlockers, func(moduleName string) bool {
		module := m.modules[moduleName]
		_, hasBeginBlock := module.(appmodulev2.HasBeginBlocker)
		return !hasBeginBlock
	}); err != nil {
		return err
	}

	if err := m.assertNoForgottenModules("EndBlockers", m.config.EndBlockers, func(moduleName string) bool {
		module := m.modules[moduleName]
		if _, hasEndBlock := module.(appmodulev2.HasEndBlocker); hasEndBlock {
			return !hasEndBlock
		}

		_, hasABCIEndBlock := module.(hasABCIEndBlock)
		return !hasABCIEndBlock
	}); err != nil {
		return err
	}

	if err := m.assertNoForgottenModules("TxValidators", m.config.TxValidators, func(moduleName string) bool {
		module := m.modules[moduleName]
		_, hasTxValidator := module.(appmodulev2.HasTxValidator[T])
		return !hasTxValidator
	}); err != nil {
		return err
	}

	if err := m.assertNoForgottenModules("InitGenesis", m.config.InitGenesis, func(moduleName string) bool {
		module := m.modules[moduleName]
		if _, hasGenesis := module.(appmodule.HasGenesisAuto); hasGenesis {
			panic(fmt.Sprintf("module %s isn't server/v2 compatible", moduleName))
		}

		if _, hasGenesis := module.(appmodulev2.HasGenesis); hasGenesis {
			return !hasGenesis
		}

		_, hasABCIGenesis := module.(appmodulev2.HasABCIGenesis)
		return !hasABCIGenesis
	}); err != nil {
		return err
	}

	if err := m.assertNoForgottenModules("ExportGenesis", m.config.ExportGenesis, func(moduleName string) bool {
		module := m.modules[moduleName]
		if _, hasGenesis := module.(appmodule.HasGenesisAuto); hasGenesis {
			panic(fmt.Sprintf("module %s isn't server/v2 compatible", moduleName))
		}

		if _, hasGenesis := module.(appmodulev2.HasGenesis); hasGenesis {
			return !hasGenesis
		}

		_, hasABCIGenesis := module.(appmodulev2.HasABCIGenesis)
		return !hasABCIGenesis
	}); err != nil {
		return err
	}

	if err := m.assertNoForgottenModules("OrderMigrations", m.config.OrderMigrations, nil); err != nil {
		return err
	}

	return nil
}

// assertNoForgottenModules checks that we didn't forget any modules in the *runtimev2.Module config.
// `pass` is a closure which allows one to omit modules from `moduleNames`.
// If you provide non-nil `pass` and it returns true, the module would not be subject of the assertion.
func (m *MM[T]) assertNoForgottenModules(
	setOrderFnName string,
	moduleNames []string,
	pass func(moduleName string) bool,
) error {
	ms := make(map[string]bool)
	for _, m := range moduleNames {
		ms[m] = true
	}
	var missing []string
	for m := range m.modules {
		if pass != nil && pass(m) {
			continue
		}

		if !ms[m] {
			missing = append(missing, m)
		}
	}

	if len(missing) != 0 {
		sort.Strings(missing)
		return fmt.Errorf("all modules must be defined when setting %s, missing: %v", setOrderFnName, missing)
	}

	return nil
}

func registerServices[T transaction.Tx](s appmodulev2.AppModule, app *App[T], registry *protoregistry.Files) error {
	// case module with services
	if services, ok := s.(hasServicesV1); ok {
		c := &configurator{
			queryHandlers:  map[string]appmodulev2.Handler{},
			stfQueryRouter: app.queryRouterBuilder,
			stfMsgRouter:   app.msgRouterBuilder,
			registry:       registry,
			err:            nil,
		}
		if err := services.RegisterServices(c); err != nil {
			return fmt.Errorf("unable to register services: %w", err)
		}

		if c.err != nil {
			app.logger.Warn("error registering services", "error", c.err)
		}

		// merge maps
		for path, decoder := range c.queryHandlers {
			app.queryHandlers[path] = decoder
		}
	}

	// if module implements register msg handlers
	if module, ok := s.(appmodulev2.HasMsgHandlers); ok {
		wrapper := newStfRouterWrapper(app.msgRouterBuilder)
		module.RegisterMsgHandlers(&wrapper)
		if wrapper.error != nil {
			return fmt.Errorf("unable to register handlers: %w", wrapper.error)
		}
	}

	// if module implements register query handlers
	if module, ok := s.(appmodulev2.HasQueryHandlers); ok {
		wrapper := newStfRouterWrapper(app.queryRouterBuilder)
		module.RegisterQueryHandlers(&wrapper)

		for path, handler := range wrapper.handlers {
			app.queryHandlers[path] = handler
		}
	}

	return nil
}

var _ grpc.ServiceRegistrar = (*configurator)(nil)

type configurator struct {
	queryHandlers map[string]appmodulev2.Handler

	stfQueryRouter *stf.MsgRouterBuilder
	stfMsgRouter   *stf.MsgRouterBuilder
	registry       *protoregistry.Files
	err            error
}

func (c *configurator) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	// first we check if it's a msg server
	prefSd, err := c.registry.FindDescriptorByName(protoreflect.FullName(sd.ServiceName))
	if err != nil {
		c.err = fmt.Errorf("register service: unable to find protov2 service descriptor: please make sure protov2 API counterparty is imported: %s", sd.ServiceName)
		return
	}

	if !proto.HasExtension(prefSd.(protoreflect.ServiceDescriptor).Options(), cosmosmsg.E_Service) {
		err = c.registerQueryHandlers(sd, ss)
		if err != nil {
			c.err = err
		}
	} else {
		err = c.registerMsgHandlers(sd, ss)
		if err != nil {
			c.err = err
		}
	}
}

func (c *configurator) registerQueryHandlers(sd *grpc.ServiceDesc, ss interface{}) error {
	for _, md := range sd.Methods {
		// TODO(tip): what if a query is not deterministic?

		handler, err := grpcHandlerToAppModuleHandler(sd, md, ss)
		if err != nil {
			return fmt.Errorf("unable to make a appmodulev2.HandlerFunc from gRPC handler (%s, %s): %w", sd.ServiceName, md.MethodName, err)
		}

		// register to stf query router.
		err = c.stfQueryRouter.RegisterHandler(gogoproto.MessageName(handler.MakeMsg()), handler.Func)
		if err != nil {
			return fmt.Errorf("unable to register handler to stf router (%s, %s): %w", sd.ServiceName, md.MethodName, err)
		}

		// register query handler using the same mapping used in stf
		c.queryHandlers[gogoproto.MessageName(handler.MakeMsg())] = handler
	}
	return nil
}

func (c *configurator) registerMsgHandlers(sd *grpc.ServiceDesc, ss interface{}) error {
	for _, md := range sd.Methods {
		handler, err := grpcHandlerToAppModuleHandler(sd, md, ss)
		if err != nil {
			return err
		}
		err = c.stfMsgRouter.RegisterHandler(gogoproto.MessageName(handler.MakeMsg()), handler.Func)
		if err != nil {
			return fmt.Errorf("unable to register msg handler %s.%s: %w", sd.ServiceName, md.MethodName, err)
		}
	}
	return nil
}

// grpcHandlerToAppModuleHandler converts a gRPC handler into an appmodulev2.HandlerFunc.
func grpcHandlerToAppModuleHandler(
	sd *grpc.ServiceDesc,
	md grpc.MethodDesc,
	ss interface{},
) (appmodulev2.Handler, error) {
	requestName, responseName, err := requestFullNameFromMethodDesc(sd, md)
	if err != nil {
		return appmodulev2.Handler{}, err
	}

	requestTyp := gogoproto.MessageType(string(requestName))
	if requestTyp == nil {
		return appmodulev2.Handler{}, fmt.Errorf("no proto message found for %s", requestName)
	}
	responseTyp := gogoproto.MessageType(string(responseName))
	if responseTyp == nil {
		return appmodulev2.Handler{}, fmt.Errorf("no proto message found for %s", responseName)
	}

	handlerFunc := func(
		ctx context.Context,
		msg transaction.Msg,
	) (resp transaction.Msg, err error) {
		res, err := md.Handler(ss, ctx, noopDecoder, messagePassingInterceptor(msg))
		if err != nil {
			return nil, err
		}
		return res.(transaction.Msg), nil
	}

	return appmodulev2.Handler{
		Func: handlerFunc,
		MakeMsg: func() transaction.Msg {
			return reflect.New(requestTyp.Elem()).Interface().(transaction.Msg)
		},
		MakeMsgResp: func() transaction.Msg {
			return reflect.New(responseTyp.Elem()).Interface().(transaction.Msg)
		},
	}, nil
}

func noopDecoder(_ interface{}) error { return nil }

func messagePassingInterceptor(msg transaction.Msg) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		return handler(ctx, msg)
	}
}

// requestFullNameFromMethodDesc returns the fully-qualified name of the request message and response of the provided service's method.
func requestFullNameFromMethodDesc(sd *grpc.ServiceDesc, method grpc.MethodDesc) (
	protoreflect.FullName, protoreflect.FullName, error,
) {
	methodFullName := protoreflect.FullName(fmt.Sprintf("%s.%s", sd.ServiceName, method.MethodName))
	desc, err := gogoproto.HybridResolver.FindDescriptorByName(methodFullName)
	if err != nil {
		return "", "", fmt.Errorf("cannot find method descriptor %s", methodFullName)
	}
	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return "", "", fmt.Errorf("invalid method descriptor %s", methodFullName)
	}
	return methodDesc.Input().FullName(), methodDesc.Output().FullName(), nil
}

// defaultMigrationsOrder returns a default migrations order: ascending alphabetical by module name,
// except x/auth which will run last, see:
// https://github.com/cosmos/cosmos-sdk/issues/10591
func defaultMigrationsOrder(modules []string) []string {
	const authName = "auth"
	out := make([]string, 0, len(modules))
	hasAuth := false
	for _, m := range modules {
		if m == authName {
			hasAuth = true
		} else {
			out = append(out, m)
		}
	}
	sort.Strings(out)
	if hasAuth {
		out = append(out, authName)
	}
	return out
}

// hasServicesV1 is the interface for registering service in baseapp Cosmos SDK.
// This API is part of core/appmodule but commented out for dependencies.
type hasServicesV1 interface {
	RegisterServices(grpc.ServiceRegistrar) error
}

var _ appmodulev2.MsgRouter = (*stfRouterWrapper)(nil)

// stfRouterWrapper wraps the stf router and implements the core appmodulev2.MsgRouter
// interface.
// The difference between this type and stf router is that the stf router expects
// us to provide it the msg name, but the core router interface does not have
// such requirement.
type stfRouterWrapper struct {
	stfRouter *stf.MsgRouterBuilder

	error error

	handlers map[string]appmodulev2.Handler
}

func newStfRouterWrapper(stfRouterBuilder *stf.MsgRouterBuilder) stfRouterWrapper {
	wrapper := stfRouterWrapper{stfRouter: stfRouterBuilder}
	wrapper.error = nil
	wrapper.handlers = map[string]appmodulev2.Handler{}
	return wrapper
}

func (s *stfRouterWrapper) RegisterHandler(handler appmodulev2.Handler) {
	req := handler.MakeMsg()
	requestName := gogoproto.MessageName(req)
	if requestName == "" {
		s.error = errors.Join(s.error, fmt.Errorf("unable to extract request name for type: %T", req))
	}

	// register handler to stf router
	err := s.stfRouter.RegisterHandler(requestName, handler.Func)
	s.error = errors.Join(s.error, err)

	// also make the decoder
	if s.handlers == nil {
		s.handlers = map[string]appmodulev2.Handler{}
	}
	s.handlers[requestName] = handler
}
