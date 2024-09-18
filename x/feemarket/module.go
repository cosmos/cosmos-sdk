package feemarket

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/spf13/cobra"

	"cosmossdk.io/x/feemarket/client/cli"
	"cosmossdk.io/x/feemarket/keeper"
	"cosmossdk.io/x/feemarket/types"
)

// ConsensusVersion is the x/feemarket module's consensus version identifier.
const ConsensusVersion = 1

var (
	_ module.HasGenesis     = AppModule{}
	_ module.AppModuleBasic = AppModule{}
	_ module.HasServices    = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasEndBlocker = AppModule{}
)

// AppModule represents an application module for the x/feemarket module.
type AppModule struct {
	cdc codec.Codec
	k   keeper.Keeper
}

// NewAppModule returns an application module for the x/feemarket module.
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		cdc: cdc,
		k:   k,
	}
}

// Name returns the name of x/feemarket module.
func (amb AppModule) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the necessary types from the x/feemarket module for amino
// serialization.
func (amb AppModule) RegisterLegacyAminoCodec(cdc registry.AminoRegistrar) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the necessary implementations / interfaces in the x/feemarket
// module w/ the interface-registry.
func (amb AppModule) RegisterInterfaces(ir codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(ir)
}

// RegisterGRPCGatewayRoutes registers the necessary REST routes for the GRPC-gateway to
// the x/feemarket module QueryService on mux. This method panics on failure.
func (amb AppModule) RegisterGRPCGatewayRoutes(cliCtx client.Context, mux *runtime.ServeMux) {
	// Register the gate-way routes w/ the provided mux.
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(cliCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd is a no-op, as no txs are registered for submission (apart from messages that
// can only be executed by governance).
func (amb AppModule) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the x/feemarket module base query cli-command.
func (amb AppModule) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// EndBlock returns an endblocker for the x/feemarket module.
func (am AppModule) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return am.k.EndBlock(sdkCtx)
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// RegisterServices registers the module's services with the app's module configurator.
func (am AppModule) RegisterServices(cfc module.Configurator) {
	types.RegisterMsgServer(cfc.MsgServer(), keeper.NewMsgServer(&am.k))
	types.RegisterQueryServer(cfc.QueryServer(), keeper.NewQueryServer(am.k))
}

// DefaultGenesis returns default genesis state as raw bytes for the feemarket
// module.
func (am AppModule) DefaultGenesis() json.RawMessage {
	return am.cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the feemarket module.
func (am AppModule) ValidateGenesis(bz json.RawMessage) error {
	var gs types.GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &gs); err != nil {
		return err
	}

	return gs.ValidateBasic()
}

// InitGenesis performs the genesis initialization for the x/feemarket module. This method returns
// no validator set updates. This method panics on any errors.
func (am AppModule) InitGenesis(ctx context.Context, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &gs); err != nil {
		return err
	}

	return am.k.InitGenesis(ctx, gs)
}

// ExportGenesis returns the feemarket module's exported genesis state as raw
// JSON bytes. This method panics on any error.
func (am AppModule) ExportGenesis(ctx context.Context) (json.RawMessage, error) {
	gs, err := am.k.ExportGenesis(ctx)
	if err != nil {
		return nil, err
	}
	return am.cdc.MarshalJSON(gs)
}
