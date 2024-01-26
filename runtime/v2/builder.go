package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	cosmosmsg "cosmossdk.io/api/cosmos/msg/v1"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/runtime/v2/protocompat"
	coreappmanager "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/stf"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"
)

type branchFunc func(state store.GetReader) store.GetWriter

// AppBuilder is a type that is injected into a container by the runtime module
// (as *AppBuilder) which can be used to create an app which is compatible with
// the existing app.go initialization conventions.
type AppBuilder struct {
	app *App

	// options for building the app
	branch      branchFunc
	txValidator func(ctx context.Context, tx transaction.Tx) error
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *AppBuilder) DefaultGenesis() map[string]json.RawMessage {
	panic("not implemented")
}

// RegisterModules registers the provided modules with the module manager and
// the basic module manager. This is the primary hook for integrating with
// modules which are not registered using the app config.
func (a *AppBuilder) RegisterModules(modules ...module.AppModule) error {
	for _, appModule := range modules {
		name := appModule.Name()
		if _, ok := a.app.moduleManager.modules[name]; ok {
			return fmt.Errorf("AppModule named %q already exists", name)
		}

		a.app.moduleManager.modules[name] = appModule
		appModule.RegisterInterfaces(a.app.interfaceRegistry)
		appModule.RegisterLegacyAminoCodec(a.app.amino)
		// register msg + query
		if services, ok := appModule.(module.HasServices); ok {
			err := registerServices(services, a.app, protoregistry.GlobalFiles)
			if err != nil {
				return err
			}
		}
		// TODO: register pre and post msg
	}

	return nil
}

// Build builds an *App instance.
func (a *AppBuilder) Build(store store.Store, opts ...AppBuilderOption) (*App, error) {
	for _, opt := range opts {
		opt(a)
	}

	// default branch
	if a.branch == nil {
		// a.branch = func(state store.GetReader) store.GetWriter {
		// 	return branch.NewStore(state)
		// }
	}

	// default tx validator
	if a.txValidator == nil {
		a.txValidator = a.app.moduleManager.TxValidation()
	}

	if err := a.app.moduleManager.RegisterMsgs(a.app.msgRouterBuilder); err != nil {
		return nil, err
	}

	stfMsgHandler, err := a.app.msgRouterBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build STF message handler: %w", err)
	}

	stfQueryHandler, err := a.app.queryRouterBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build STF query handler: %w", err)
	}

	endBlocker, valUpdate := a.app.moduleManager.EndBlock()

	a.app.stf = stf.NewSTF[transaction.Tx](
		stfMsgHandler,
		stfQueryHandler,
		a.app.moduleManager.PreBlocker(),
		a.app.moduleManager.BeginBlock(),
		endBlocker,
		a.txValidator,
		valUpdate,
		a.branch,
	)
	a.app.store = store

	return a.app, nil
}

// AppBuilderOption is a function that can be passed to AppBuilder.Build to
// customize the resulting app.
type AppBuilderOption func(*AppBuilder)

func AppBuilderWithMempool(mempool mempool.Mempool[transaction.Tx]) AppBuilderOption {
	return func(a *AppBuilder) {
		a.app.mempool = mempool
	}
}

func AppBuilderWithPrepareBlockHandler(handler coreappmanager.PrepareHandler[transaction.Tx]) AppBuilderOption {
	return func(a *AppBuilder) {
		a.app.prepareBlockHandler = handler
	}
}

func AppBuilderWithVerifyBlockHandler(handler coreappmanager.ProcessHandler[transaction.Tx]) AppBuilderOption {
	return func(a *AppBuilder) {
		a.app.verifyBlockHandler = handler
	}
}

func AppBuilderWithBranch(branch branchFunc) AppBuilderOption {
	return func(a *AppBuilder) {
		a.branch = branch
	}
}

// AppBuilderWithTxValidator sets the tx validator for the app.
// It overrides the default tx validator from all modules.
func AppBuilderWithTxValidator(validator func(ctx context.Context, tx transaction.Tx) error) AppBuilderOption {
	return func(a *AppBuilder) {
		a.txValidator = validator
	}
}

func registerServices(s module.HasServices, app *App, registry *protoregistry.Files) error {
	c := configurator{
		stfQueryRouter: app.queryRouterBuilder,
		stfMsgRouter:   app.msgRouterBuilder,
	}
	s.RegisterServices(&configurator{
		cdc:            app.cdc,
		stfQueryRouter: app.queryRouterBuilder,
		stfMsgRouter:   app.msgRouterBuilder,
		registry:       registry,
		err:            nil,
	})
	return c.Error()
}

var _ module.Configurator = (*configurator)(nil)

type configurator struct {
	cdc            codec.BinaryCodec
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

	if !protobuf.HasExtension(prefSd.(protoreflect.ServiceDescriptor).Options(), cosmosmsg.E_Service) {
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

func (c *configurator) Error() error {
	return c.err
}

func (c *configurator) MsgServer() grpc1.Server {
	panic("do not call me")
}

func (c *configurator) QueryServer() grpc1.Server {
	panic("do not call me")
}

func (c *configurator) RegisterMigration(moduleName string, fromVersion uint64, handler module.MigrationHandler) error {
	// TODO implement me
	panic("implement me")
}

func (c *configurator) registerQueryHandlers(sd *grpc.ServiceDesc, ss interface{}) error {
	for _, md := range sd.Methods {
		err := registerMethod(c.cdc, c.stfQueryRouter, sd, md, ss)
		if err != nil {
			return fmt.Errorf("unable to register query handler %s: %w", md.MethodName, err)
		}
	}
	return nil
}

func (c *configurator) registerMsgHandlers(sd *grpc.ServiceDesc, ss interface{}) error {
	for _, md := range sd.Methods {
		err := registerMethod(c.cdc, c.stfMsgRouter, sd, md, ss)
		if err != nil {
			return fmt.Errorf("unable to register msg handler %s: %w", md.MethodName, err)
		}
	}
	return nil
}

func registerMethod(cdc codec.BinaryCodec, stfRouter *stf.MsgRouterBuilder, sd *grpc.ServiceDesc, md grpc.MethodDesc, ss interface{}) error {
	requestName, err := protocompat.RequestFullNameFromMethodDesc(sd, md)
	if err != nil {
		return err
	}

	responseName, err := protocompat.ResponseFullNameFromMethodDesc(sd, md)
	if err != nil {
		return err
	}

	// now we create the hybrid handler
	hybridHandler, err := protocompat.MakeHybridHandler(cdc, sd, md, ss)
	if err != nil {
		return err
	}

	responseV2Type, err := protoregistry.GlobalTypes.FindMessageByName(responseName)
	if err != nil {
		return err
	}

	return stfRouter.RegisterHandler(string(requestName), func(ctx context.Context, msg transaction.Type) (resp transaction.Type, err error) {
		resp = responseV2Type.New().Interface()
		return resp, hybridHandler(ctx, msg.(protoiface.MessageV1), resp.(protoiface.MessageV1))
	})
}
