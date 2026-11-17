package safety

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/safety/keeper"
	"github.com/cosmos/cosmos-sdk/x/safety/simulation"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

// ConsensusVersion defines the current x/safety consensus version.
const ConsensusVersion = 1

var (
	_ module.AppModuleBasic      = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.AppModuleSimulation = AppModule{}
	_ module.HasProposalMsgs     = AppModule{}
	_ simsx.HasProposalMsgsX     = AppModule{}

	_ appmodule.AppModule   = AppModule{}
	_ appmodule.HasServices = AppModule{}
)

// AppModule implements the AppModule interface for x/safety.
type AppModule struct {
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule.
func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (AppModule) IsOnePerModuleType() {}

// Name returns the x/safety module name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the x/safety types on the LegacyAmino codec.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the x/safety interface implementations.
func (AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for x/safety.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterServices registers the x/safety msg and query services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	keeper.RegisterMsgServer(registrar, am.keeper)
	types.RegisterQueryServer(registrar, keeper.NewQueryServer(am.keeper))
	return nil
}

// defaultAuthority returns the keeper's default authority. A zero-value
// AppModule, used as a module basic, has no keeper and falls back to the gov
// module address.
func (am AppModule) defaultAuthority() string {
	if a := am.keeper.DefaultAuthority(); a != "" {
		return a
	}
	return authtypes.NewModuleAddress(govtypes.ModuleName).String()
}

// addressCodec returns the keeper's address codec. A zero-value AppModule
// falls back to the bech32 account prefix in the global sdk config.
func (am AppModule) addressCodec() address.Codec {
	if ac := am.keeper.AddressCodec(); ac != nil {
		return ac
	}
	return addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
}

// DefaultGenesis returns the default genesis state: the keeper's default
// authority, no operator, an empty scope, and an empty disabled list.
func (am AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState(am.defaultAuthority()))
}

// ValidateGenesis validates the x/safety genesis state.
func (am AppModule) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return gs.Validate(am.addressCodec())
}

// InitGenesis validates and writes the x/safety genesis state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	var gs types.GenesisState
	cdc.MustUnmarshalJSON(bz, &gs)
	if err := am.keeper.InitGenesis(ctx, &gs); err != nil {
		panic(fmt.Errorf("failed to initialize %s genesis state: %w", types.ModuleName, err))
	}
}

// ExportGenesis exports the x/safety state as raw JSON.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to export %s genesis state: %w", types.ModuleName, err))
	}
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements HasConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// GenerateGenesisState creates a randomized x/safety genesis state.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState, am.defaultAuthority())
}

// RegisterStoreDecoder registers a decoder for x/safety store keys.
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simtypes.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.Schema)
}

// ProposalMsgs returns the msgs simulated gov proposals contain: a random
// MsgUpdateParams.
func (AppModule) ProposalMsgs(module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// ProposalMsgsX registers the msgs simulated gov proposals contain.
func (AppModule) ProposalMsgsX(weights simsx.WeightSource, reg simsx.Registry) {
	reg.Add(weights.Get("msg_update_params", uint32(simulation.DefaultWeightMsgUpdateParams)), simulation.MsgUpdateParamsFactory())
}

// WeightedOperations returns no operations; tripping in random simulation
// would stall every other module.
func (AppModule) WeightedOperations(module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
