package evidence

import (
	"context"
	"encoding/json"
	"fmt"
	eviclient "github.com/cosmos/cosmos-sdk/contrib/x/evidence/client"
	"github.com/cosmos/cosmos-sdk/contrib/x/evidence/client/cli"
	keeper2 "github.com/cosmos/cosmos-sdk/contrib/x/evidence/keeper"
	"github.com/cosmos/cosmos-sdk/contrib/x/evidence/simulation"
	types2 "github.com/cosmos/cosmos-sdk/contrib/x/evidence/types"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	modulev1 "cosmossdk.io/api/cosmos/evidence/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/comet"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

var (
	_ module.AppModuleBasic      = AppModule{}
	_ module.AppModuleSimulation = AppModule{}
	_ module.HasGenesis          = AppModule{}

	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasServices     = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface for the evidence module.
type AppModuleBasic struct {
	evidenceHandlers []eviclient.EvidenceHandler // eviclient evidence submission handlers
}

// NewAppModuleBasic creates a AppModuleBasic without the codec.
func NewAppModuleBasic(evidenceHandlers ...eviclient.EvidenceHandler) AppModuleBasic {
	return AppModuleBasic{
		evidenceHandlers: evidenceHandlers,
	}
}

// Name returns the evidence module's name.
func (AppModuleBasic) Name() string {
	return types2.ModuleName
}

// RegisterLegacyAminoCodec registers the evidence module's types to the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types2.RegisterLegacyAminoCodec(cdc)
}

// DefaultGenesis returns the evidence module's default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types2.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the evidence module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types2.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types2.ModuleName, err)
	}

	return gs.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the evidence module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types2.RegisterQueryHandlerClient(context.Background(), mux, types2.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the evidence module's root tx command.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	evidenceCLIHandlers := make([]*cobra.Command, len(a.evidenceHandlers))

	for i, evidenceHandler := range a.evidenceHandlers {
		evidenceCLIHandlers[i] = evidenceHandler.CLIHandler()
	}

	return cli.GetTxCmd(evidenceCLIHandlers)
}

// RegisterInterfaces registers the evidence module's interface types
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types2.RegisterInterfaces(registry)
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the evidence module.
type AppModule struct {
	AppModuleBasic

	keeper keeper2.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(keeper keeper2.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types2.RegisterMsgServer(registrar, keeper2.NewMsgServerImpl(am.keeper))
	types2.RegisterQueryServer(registrar, keeper2.NewQuerier(&am.keeper))
	return nil
}

// InitGenesis performs the evidence module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	var gs types2.GenesisState
	err := cdc.UnmarshalJSON(bz, &gs)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal %s genesis state: %s", types2.ModuleName, err))
	}

	InitGenesis(ctx, am.keeper, &gs)
}

// ExportGenesis returns the evidence module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(ExportGenesis(ctx, am.keeper))
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes all ABCI BeginBlock logic respective to the evidence module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return am.keeper.BeginBlocker(ctx)
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the evidence module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for evidence module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[types2.StoreKey] = simtypes.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.Schema)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}

//
// App Wiring Setup
//

func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	StoreService store.KVStoreService
	Cdc          codec.Codec

	StakingKeeper  types2.StakingKeeper
	SlashingKeeper types2.SlashingKeeper
	AddressCodec   address.Codec

	BlockInfoService comet.BlockInfoService
}

type ModuleOutputs struct {
	depinject.Out

	EvidenceKeeper keeper2.Keeper
	Module         appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper2.NewKeeper(in.Cdc, in.StoreService, in.StakingKeeper, in.SlashingKeeper, in.AddressCodec, in.BlockInfoService)
	m := NewAppModule(*k)

	return ModuleOutputs{EvidenceKeeper: *k, Module: m}
}
