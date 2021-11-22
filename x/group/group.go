package group

import (
	"math/rand"

	"github.com/gorilla/mux"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "group"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

type AppModule struct {
	keeper Keeper
	// Registry      types.InterfaceRegistry
	BankKeeper    BankKeeper
	AccountKeeper AccountKeeper
}

// var _ module.AppModuleBasic = Module{}
// var _ module.AppModuleSimulation = Module{}
// var _ servermodule.Module = Module{}

func AccountCondition(id uint64) Condition {
	return NewCondition("group", "account", orm.EncodeSequence(id))
}

func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

func (a AppModule) Name() string {
	return ModuleName
}

// RegisterInterfaces registers module concrete types into protobuf Any.
func (a AppModule) RegisterInterfaces(registry types.InterfaceRegistry) {
	RegisterTypes(registry)
}

// func (a AppModule) RegisterServices(configurator servermodule.Configurator) {
// 	RegisterMsgServer(configurator.MsgServer(), NewMsgServerImpl(a.keeper))
// 	RegisterQueryServer(configurator.QueryServer(), a.keeper)
// 	a.RegisterServices(configurator, a.AccountKeeper, a.BankKeeper)
// }

// func (a Module) DefaultGenesis(marshaler codec.JSONCodec) json.RawMessage {
// 	return marshaler.MustMarshalJSON(NewGenesisState())
// }

// func (a Module) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
// 	var data GenesisState
// 	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
// 		return fmt.Errorf("failed to unmarshal %s genesis state: %w", group.ModuleName, err)
// 	}
// 	return data.Validate()
// }

// func (a Module) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
// 	RegisterQueryHandlerClient(context.Background(), mux, NewQueryClient(clientCtx))
// }

// func (a Module) GetTxCmd() *cobra.Command {
// 	return client.TxCmd(a.Name())
// }

// func (a Module) GetQueryCmd() *cobra.Command {
// 	return client.QueryCmd(a.Name())
// }

/**** DEPRECATED ****/
func (a AppModule) RegisterRESTRoutes(sdkclient.Context, *mux.Router) {}
func (a AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	RegisterLegacyAminoCodec(cdc)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenesisState of the group module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the group content functions used to
// simulate proposals.
func (AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized group param changes for the simulator.
func (AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for group module's types
func (AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
}

// WeightedOperations returns all the group module operations with their respective weights.
// NOTE: This is no longer needed for the modules which uses ADR-33, group module `WeightedOperations`
// registered in the `x/group/server` package.
func (AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
