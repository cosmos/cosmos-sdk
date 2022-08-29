package runtime

import (
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// AppModuleWrapper is a type used for injecting a module.AppModule into a
// container so that it can be used by the runtime module.
type AppModuleWrapper struct {
	module.AppModule
}

// WrapAppModule wraps a module.AppModule so that it can be injected into
// a container for use by the runtime module.
func WrapAppModule(appModule module.AppModule) AppModuleWrapper {
	return AppModuleWrapper{AppModule: appModule}
}

// IsOnePerModuleType identifies this type as a depinject.OnePerModuleType.
func (AppModuleWrapper) IsOnePerModuleType() {}

// AppModuleBasicWrapper is a type used for injecting a module.AppModuleBasic
// into a container so that it can be used by the runtime module.
type AppModuleBasicWrapper struct {
	module.AppModuleBasic
}

// WrapAppModuleBasic wraps a module.AppModuleBasic so that it can be injected into
// a container for use by the runtime module.
func WrapAppModuleBasic(basic module.AppModuleBasic) AppModuleBasicWrapper {
	return AppModuleBasicWrapper{AppModuleBasic: basic}
}

// IsOnePerModuleType identifies this type as a depinject.OnePerModuleType.
func (AppModuleBasicWrapper) IsOnePerModuleType() {}

type handlerAppModuleWrapper struct {
	moduleName string
	appmodule.Handler
}

func (h handlerAppModuleWrapper) Name() string {
	return h.moduleName
}

func (h handlerAppModuleWrapper) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {}

func (h handlerAppModuleWrapper) RegisterInterfaces(registry types.InterfaceRegistry) {}

func (h handlerAppModuleWrapper) DefaultGenesis(codec codec.JSONCodec) json.RawMessage {
	//TODO implement me
	panic("implement me")
}

func (h handlerAppModuleWrapper) ValidateGenesis(codec codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	//TODO implement me
	panic("implement me")
}

func (h handlerAppModuleWrapper) RegisterGRPCGatewayRoutes(context client.Context, mux *runtime.ServeMux) {
}

func (h handlerAppModuleWrapper) GetTxCmd() *cobra.Command {
	return nil
}

func (h handlerAppModuleWrapper) GetQueryCmd() *cobra.Command {
	return nil
}

func (h handlerAppModuleWrapper) InitGenesis(context sdk.Context, codec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

func (h handlerAppModuleWrapper) ExportGenesis(context sdk.Context, codec codec.JSONCodec) json.RawMessage {
	return nil
}

func (h handlerAppModuleWrapper) RegisterInvariants(registry sdk.InvariantRegistry) {}

func (h handlerAppModuleWrapper) RegisterServices(configurator module.Configurator) {}

func (h handlerAppModuleWrapper) ConsensusVersion() uint64 {
	return 0
}

var _ module.AppModule = handlerAppModuleWrapper{}
