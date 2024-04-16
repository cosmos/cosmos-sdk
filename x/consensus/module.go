package consensus

import (
	"context"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

// ConsensusVersion defines the current x/consensus module consensus version.
const ConsensusVersion = 1

var (
	_ module.HasName        = AppModule{}
	_ module.HasAminoCodec  = AppModule{}
	_ module.HasGRPCGateway = AppModule{}

	_ appmodule.AppModule             = AppModule{}
	_ appmodule.HasRegisterInterfaces = AppModule{}
)

// AppModule implements an application module
type AppModule struct {
	cdc    codec.Codec
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		cdc:    cdc,
		keeper: keeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the consensus module's name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the consensus module's types on the LegacyAmino codec.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModule) RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	types.RegisterInterfaces(registrar)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, am.keeper)
	types.RegisterQueryServer(registrar, am.keeper)
	return nil
}

// ConsensusVersion implements HasConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// RegisterConsensusMessages registers the consensus module's messages.
func (am AppModule) RegisterConsensusMessages(builder any) {
	// std.RegisterConsensusHandler(builder ,am.keeper.SetParams) // TODO uncomment when api is available
}
