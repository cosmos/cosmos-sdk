package module

import (
	"context"
	"encoding/json"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/errors"
	"cosmossdk.io/x/nft"
	"cosmossdk.io/x/nft/keeper"
	"cosmossdk.io/x/nft/simulation"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

var (
	_ module.HasName                  = AppModule{}
	_ module.HasGRPCGateway           = AppModule{}
	_ appmodule.HasRegisterInterfaces = AppModule{}
	_ module.AppModuleSimulation      = AppModule{}
	_ module.HasGenesis               = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

const ConsensusVersion = 1

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	cdc      codec.Codec
	registry cdctypes.InterfaceRegistry

	keeper        keeper.Keeper
	accountKeeper nft.AccountKeeper
	bankKeeper    nft.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, ak nft.AccountKeeper, bk nft.BankKeeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		cdc:           cdc,
		keeper:        keeper,
		accountKeeper: ak,
		bankKeeper:    bk,
		registry:      registry,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the nft module's name.
func (AppModule) Name() string {
	return nft.ModuleName
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	nft.RegisterMsgServer(registrar, am.keeper)
	nft.RegisterQueryServer(registrar, am.keeper)
	return nil
}

// RegisterInterfaces registers the nft module's interface types
func (AppModule) RegisterInterfaces(registry registry.LegacyRegistry) {
	nft.RegisterInterfaces(registry)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the nft module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *gwruntime.ServeMux) {
	if err := nft.RegisterQueryHandlerClient(context.Background(), mux, nft.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// DefaultGenesis returns default genesis state as raw bytes for the nft module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(nft.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the nft module.
func (am AppModule) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data nft.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return errors.Wrapf(err, "failed to unmarshal %s genesis state", nft.ModuleName)
	}

	return nft.ValidateGenesis(data, am.accountKeeper.AddressCodec())
}

// InitGenesis performs genesis initialization for the nft module.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState nft.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the nft module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the nft module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState, am.accountKeeper.AddressCodec())
}

// RegisterStoreDecoder registers a decoder for nft module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[keeper.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the nft module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		am.registry,
		simState.AppParams, simState.Cdc, simState.TxConfig,
		am.accountKeeper, am.bankKeeper, am.keeper,
	)
}
