package module

import (
	"cosmossdk.io/core/appmodule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// UseCoreAPIModule wraps the core API module as an AppModule that this version
// of the SDK can use.
func UseCoreAPIModule(name string, module appmodule.AppModule) AppModuleBasic {
	return coreAppModuleBasicAdapator{
		name:   name,
		module: module,
	}
}

type coreAppModuleBasicAdapator struct {
	name   string
	module appmodule.AppModule
}

// Name implements AppModuleBasic
func (c coreAppModuleBasicAdapator) Name() string {
	return c.name
}

// GetQueryCmd implements AppModuleBasic
func (c coreAppModuleBasicAdapator) GetQueryCmd() *cobra.Command {
	if mod, ok := c.module.(interface {
		GetQueryCmd() *cobra.Command
	}); ok {
		return mod.GetQueryCmd()
	}

	return nil
}

// GetTxCmd implements AppModuleBasic
func (c coreAppModuleBasicAdapator) GetTxCmd() *cobra.Command {
	if mod, ok := c.module.(interface {
		GetTxCmd() *cobra.Command
	}); ok {
		return mod.GetTxCmd()
	}

	return nil
}

// RegisterGRPCGatewayRoutes implements AppModuleBasic
func (c coreAppModuleBasicAdapator) RegisterGRPCGatewayRoutes(ctx client.Context, mux *runtime.ServeMux) {
	if mod, ok := c.module.(interface {
		RegisterGRPCGatewayRoutes(context client.Context, mux *runtime.ServeMux)
	}); ok {
		mod.RegisterGRPCGatewayRoutes(ctx, mux)
	}
}

// RegisterInterfaces implements AppModuleBasic
func (c coreAppModuleBasicAdapator) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	if mod, ok := c.module.(interface {
		RegisterInterfaces(registry types.InterfaceRegistry)
	}); ok {
		mod.RegisterInterfaces(registry)
	}
}

// RegisterLegacyAminoCodec implements AppModuleBasic
func (c coreAppModuleBasicAdapator) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	if mod, ok := c.module.(interface {
		RegisterLegacyAminoCodec(amino *codec.LegacyAmino)
	}); ok {
		mod.RegisterLegacyAminoCodec(amino)
	}
}
