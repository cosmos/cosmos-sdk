package module

import (
	"context"
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/genesis"
	"cosmossdk.io/core/registry"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ appmodule.AppModule = coreAppModuleAdaptor{}

	_ HasName                         = coreAppModuleAdaptor{}
	_ HasAminoCodec                   = coreAppModuleAdaptor{}
	_ HasGRPCGateway                  = coreAppModuleAdaptor{}
	_ appmodule.HasRegisterInterfaces = coreAppModuleAdaptor{}
	_ HasABCIGenesis                  = coreAppModuleAdaptor{}
	_ HasServices                     = coreAppModuleAdaptor{}
)

// CoreAppModuleAdaptor wraps the core API module as an AppModule that this version of the SDK can use.
func CoreAppModuleAdaptor(name string, module appmodule.AppModule) AppModule {
	return coreAppModuleAdaptor{
		name:   name,
		module: module,
	}
}

type coreAppModuleAdaptor struct {
	name   string
	module appmodule.AppModule
}

// DefaultGenesis implements HasGenesis
func (c coreAppModuleAdaptor) DefaultGenesis() json.RawMessage {
	if mod, ok := c.module.(appmodule.HasGenesisAuto); ok {
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

	if mod, ok := c.module.(HasGenesisBasics); ok {
		return mod.DefaultGenesis()
	}

	if mod, ok := c.module.(HasGenesis); ok {
		return mod.DefaultGenesis()
	}

	return nil
}

// ValidateGenesis implements HasGenesis
func (c coreAppModuleAdaptor) ValidateGenesis(bz json.RawMessage) error {
	if mod, ok := c.module.(appmodule.HasGenesisAuto); ok {
		source, err := genesis.SourceFromRawJSON(bz)
		if err != nil {
			return err
		}

		if err := mod.ValidateGenesis(source); err != nil {
			return err
		}
	}

	if mod, ok := c.module.(HasGenesisBasics); ok {
		return mod.ValidateGenesis(bz)
	}

	if mod, ok := c.module.(HasGenesis); ok {
		return mod.ValidateGenesis(bz)
	}

	return nil
}

// ExportGenesis implements HasGenesis
func (c coreAppModuleAdaptor) ExportGenesis(ctx context.Context) json.RawMessage {
	if module, ok := c.module.(appmodule.HasGenesisAuto); ok {
		ctx := sdk.UnwrapSDKContext(ctx).WithGasMeter(storetypes.NewInfiniteGasMeter()) // avoid race conditions
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

	if mod, ok := c.module.(HasABCIGenesis); ok {
		return mod.ExportGenesis(ctx)
	}

	if mod, ok := c.module.(HasGenesis); ok {
		eg, err := mod.ExportGenesis(ctx)
		if err != nil {
			panic(err)
		}
		return eg
	}

	return nil
}

// InitGenesis implements HasGenesis
func (c coreAppModuleAdaptor) InitGenesis(ctx context.Context, bz json.RawMessage) []abci.ValidatorUpdate {
	if module, ok := c.module.(appmodule.HasGenesisAuto); ok {
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

	if mod, ok := c.module.(HasABCIGenesis); ok {
		return mod.InitGenesis(ctx, bz)
	}

	if mod, ok := c.module.(HasGenesis); ok {
		err := mod.InitGenesis(ctx, bz)
		if err != nil {
			panic(err)
		}

	}
	return nil
}

// Name implements HasName
func (c coreAppModuleAdaptor) Name() string {
	return c.name
}

func (c coreAppModuleAdaptor) GetQueryCmd() *cobra.Command {
	if mod, ok := c.module.(interface {
		GetQueryCmd() *cobra.Command
	}); ok {
		return mod.GetQueryCmd()
	}

	return nil
}

func (c coreAppModuleAdaptor) GetTxCmd() *cobra.Command {
	if mod, ok := c.module.(interface {
		GetTxCmd() *cobra.Command
	}); ok {
		return mod.GetTxCmd()
	}

	return nil
}

// RegisterGRPCGatewayRoutes implements HasGRPCGateway
func (c coreAppModuleAdaptor) RegisterGRPCGatewayRoutes(ctx client.Context, mux *runtime.ServeMux) {
	if mod, ok := c.module.(interface {
		RegisterGRPCGatewayRoutes(context client.Context, mux *runtime.ServeMux)
	}); ok {
		mod.RegisterGRPCGatewayRoutes(ctx, mux)
	}
}

// RegisterInterfaces implements HasRegisterInterfaces
func (c coreAppModuleAdaptor) RegisterInterfaces(reg registry.LegacyRegistry) {
	if mod, ok := c.module.(interface {
		RegisterInterfaces(registry.LegacyRegistry)
	}); ok {
		mod.RegisterInterfaces(reg)
	}
}

// RegisterLegacyAminoCodec implements HasAminoCodec
func (c coreAppModuleAdaptor) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	if mod, ok := c.module.(interface {
		RegisterLegacyAminoCodec(amino *codec.LegacyAmino)
	}); ok {
		mod.RegisterLegacyAminoCodec(amino)
	}
}

// RegisterServices implements HasServices
func (c coreAppModuleAdaptor) RegisterServices(cfg Configurator) {
	if module, ok := c.module.(appmodule.HasServices); ok {
		err := module.RegisterServices(cfg)
		if err != nil {
			panic(err)
		}
	}

	if module, ok := c.module.(appmodule.HasMigrations); ok {
		err := module.RegisterMigrations(cfg)
		if err != nil {
			panic(err)
		}
	}
}

func (c coreAppModuleAdaptor) IsOnePerModuleType() {}

func (c coreAppModuleAdaptor) IsAppModule() {}
