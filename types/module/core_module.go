package module

import (
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/genesis"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ AppModuleBasic = coreAppModuleBasicAdapator{}
	_ HasGenesis     = coreAppModuleBasicAdapator{}
)

// CoreAppModuleBasicAdaptor wraps the core API module as an AppModule that this version
// of the SDK can use.
func CoreAppModuleBasicAdaptor(name string, module appmodule.AppModule) AppModuleBasic {
	return coreAppModuleBasicAdapator{
		name:   name,
		module: module,
	}
}

type coreAppModuleBasicAdapator struct {
	name   string
	module appmodule.AppModule
}

// DefaultGenesis implements HasGenesis
func (c coreAppModuleBasicAdapator) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	if mod, ok := c.module.(appmodule.HasGenesis); ok {
		target := genesis.RawJSONTarget{}
		err := mod.DefaultGenesis(target.Target())
		if err != nil {
			panic(err)
		}

		res, err := target.JSON()
		if err != nil {
			panic(err)
		}

		return res
	}
	return nil
}

// ValidateGenesis implements HasGenesis
func (c coreAppModuleBasicAdapator) ValidateGenesis(cdc codec.JSONCodec, txConfig client.TxEncodingConfig, bz json.RawMessage) error {
	if mod, ok := c.module.(appmodule.HasGenesis); ok {
		source, err := genesis.SourceFromRawJSON(bz)
		if err != nil {
			return err
		}

		if err := mod.ValidateGenesis(source); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis implements HasGenesis
func (c coreAppModuleBasicAdapator) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	if module, ok := c.module.(appmodule.HasGenesis); ok {
		ctx := ctx.WithGasMeter(storetypes.NewInfiniteGasMeter()) // avoid race conditions
		target := genesis.RawJSONTarget{}
		err := module.ExportGenesis(ctx, target.Target())
		if err != nil {
			panic(err)
		}

		rawJSON, err := target.JSON()
		if err != nil {
			panic(err)
		}

		return rawJSON
	}
	return nil
}

// InitGenesis implements HasGenesis
func (c coreAppModuleBasicAdapator) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, bz json.RawMessage) []abci.ValidatorUpdate {
	if module, ok := c.module.(appmodule.HasGenesis); ok {
		// core API genesis
		source, err := genesis.SourceFromRawJSON(bz)
		if err != nil {
			panic(err)
		}

		err = module.InitGenesis(ctx, source)
		if err != nil {
			panic(err)
		}
	}
	return nil
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
		RegisterInterfaces(registry codectypes.InterfaceRegistry)
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
