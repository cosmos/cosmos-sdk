package circuit

import (
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/x/circuit/keeper"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
)

// ConsensusVersion defines the current x/bank module consensus version.
const ConsensusVersion = 4

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the bank module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the bank module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the bank module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// types.RegisterLegacyAminoCodec(cdc) //TODO:implement
}

// DefaultGenesis returns default genesis state as raw bytes for the bank
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the bank module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	// var data types.GenesisState
	// if err := cdc.UnmarshalJSON(bz, &data); err != nil {
	// 	return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	// }

	// return data.Validate()
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the bank module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	// if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
	// 	panic(err)
	// }
}

// GetTxCmd returns the root tx command for the bank module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil //TODO:implement
}

// GetQueryCmd returns no root query command for the bank module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil //TODO:implement
}

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// types.RegisterInterfaces(registry)
}

// AppModule implements an application module for the bank module.
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

var _ appmodule.AppModule = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	// types.RegisterQueryServer(cfg.QueryServer(), am.keeper)

}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// Name returns the bank module's name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterInvariants registers the bank module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// AppModuleSimulation functions

// // GenerateGenesisState creates a randomized GenState of the circuit module.
// func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
// 	simulation.RandomizedGenState(simState)
// }

// // ProposalContents doesn't return any content functions for governance proposals.
// func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
// 	return nil
// }

// // RegisterStoreDecoder registers a decoder for supply module's types
// func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// // WeightedOperations returns the all the gov module operations with their respective weights.
// func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
// 	return simulation.WeightedOperations(
// 		simState.AppParams, simState.Cdc, am.keeper,
// 	)
// }
