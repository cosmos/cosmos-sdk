package gov

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	govclient "cosmossdk.io/x/gov/client"
	"cosmossdk.io/x/gov/client/cli"
	"cosmossdk.io/x/gov/keeper"
	"cosmossdk.io/x/gov/simulation"
	govtypes "cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

const ConsensusVersion = 6

var (
	_ module.HasName               = AppModule{}
	_ module.HasAminoCodec         = AppModule{}
	_ module.HasGRPCGateway        = AppModule{}
	_ module.HasRegisterInterfaces = AppModule{}
	_ module.AppModuleSimulation   = AppModule{}
	_ module.HasGenesis            = AppModule{}
	_ module.HasInvariants         = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasEndBlocker = AppModule{}
	_ appmodule.HasServices   = AppModule{}
	_ appmodule.HasMigrations = AppModule{}
)

// AppModule implements an application module for the gov module.
type AppModule struct {
	cdc                    codec.Codec
	legacyProposalHandlers []govclient.ProposalHandler

	keeper        *keeper.Keeper
	accountKeeper govtypes.AccountKeeper
	bankKeeper    govtypes.BankKeeper
	poolKeeper    govtypes.PoolKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec, keeper *keeper.Keeper,
	ak govtypes.AccountKeeper, bk govtypes.BankKeeper,
	pk govtypes.PoolKeeper, legacyProposalHandlers ...govclient.ProposalHandler,
) AppModule {
	return AppModule{
		cdc:                    cdc,
		legacyProposalHandlers: legacyProposalHandlers,
		keeper:                 keeper,
		accountKeeper:          ak,
		bankKeeper:             bk,
		poolKeeper:             pk,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the gov module's name.
func (AppModule) Name() string {
	return govtypes.ModuleName
}

// RegisterLegacyAminoCodec registers the gov module's types for the given codec.
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	v1beta1.RegisterLegacyAminoCodec(cdc)
	v1.RegisterLegacyAminoCodec(cdc)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the gov module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := v1.RegisterQueryHandlerClient(context.Background(), mux, v1.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
	if err := v1beta1.RegisterQueryHandlerClient(context.Background(), mux, v1beta1.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the gov module.
func (am AppModule) GetTxCmd() *cobra.Command {
	legacyProposalCLIHandlers := getProposalCLIHandlers(am.legacyProposalHandlers)

	return cli.NewTxCmd(legacyProposalCLIHandlers)
}

func getProposalCLIHandlers(handlers []govclient.ProposalHandler) []*cobra.Command {
	proposalCLIHandlers := make([]*cobra.Command, 0, len(handlers))
	for _, proposalHandler := range handlers {
		proposalCLIHandlers = append(proposalCLIHandlers, proposalHandler.CLIHandler())
	}
	return proposalCLIHandlers
}

// RegisterInterfaces implements InterfaceModule.RegisterInterfaces
func (AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	v1.RegisterInterfaces(registry)
	v1beta1.RegisterInterfaces(registry)
}

// RegisterInvariants registers module invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper, am.bankKeeper)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	msgServer := keeper.NewMsgServerImpl(am.keeper)
	v1beta1.RegisterMsgServer(registrar, keeper.NewLegacyMsgServerImpl(am.accountKeeper.GetModuleAddress(govtypes.ModuleName).String(), msgServer))
	v1.RegisterMsgServer(registrar, msgServer)

	v1beta1.RegisterQueryServer(registrar, keeper.NewLegacyQueryServer(am.keeper))
	v1.RegisterQueryServer(registrar, keeper.NewQueryServer(am.keeper))

	return nil
}

// RegisterMigrations registers module migrations
func (am AppModule) RegisterMigrations(mr appmodule.MigrationRegistrar) error {
	m := keeper.NewMigrator(am.keeper)
	if err := mr.Register(govtypes.ModuleName, 1, m.Migrate1to2); err != nil {
		return fmt.Errorf("failed to migrate x/gov from version 1 to 2: %w", err)
	}

	if err := mr.Register(govtypes.ModuleName, 2, m.Migrate2to3); err != nil {
		return fmt.Errorf("failed to migrate x/gov from version 2 to 3: %w", err)
	}

	if err := mr.Register(govtypes.ModuleName, 3, m.Migrate3to4); err != nil {
		return fmt.Errorf("failed to migrate x/gov from version 3 to 4: %w", err)
	}

	if err := mr.Register(govtypes.ModuleName, 4, m.Migrate4to5); err != nil {
		return fmt.Errorf("failed to migrate x/gov from version 4 to 5: %w", err)
	}

	if err := mr.Register(govtypes.ModuleName, 5, m.Migrate5to6); err != nil {
		return fmt.Errorf("failed to migrate x/gov from version 5 to 6: %w", err)
	}

	return nil
}

// DefaultGenesis returns default genesis state as raw bytes for the gov module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(v1.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the gov module.
func (am AppModule) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data v1.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", govtypes.ModuleName, err)
	}

	return v1.ValidateGenesis(am.accountKeeper.AddressCodec(), &data)
}

// InitGenesis performs genesis initialization for the gov module.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState v1.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.accountKeeper, am.bankKeeper, am.keeper, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the gov module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := ExportGenesis(ctx, am.keeper)
	if err != nil {
		panic(err)
	}
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// EndBlock returns the end blocker for the gov module.
func (am AppModule) EndBlock(ctx context.Context) error {
	c := sdk.UnwrapSDKContext(ctx)
	return EndBlocker(c, am.keeper)
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the gov module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the gov content functions used to
// simulate governance proposals.
func (AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent { //nolint:staticcheck // used for legacy testing
	return simulation.ProposalContents()
}

// ProposalMsgs returns all the gov msgs used to simulate governance proposals.
func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// RegisterStoreDecoder registers a decoder for gov module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[govtypes.StoreKey] = simtypes.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.Schema)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.TxConfig,
		am.accountKeeper, am.bankKeeper, am.keeper,
		simState.ProposalMsgs, simState.LegacyProposalContents,
	)
}
