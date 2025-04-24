// Deprecated: This package is deprecated and will be removed in the next major release. The `x/group` module will be moved to a separate repo `github.com/cosmos/cosmos-sdk-legacy`.
package module

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	modulev1 "cosmossdk.io/api/cosmos/group/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	store "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/group"            //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/group/client/cli" //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/group/keeper"     //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/group/simulation" //nolint:staticcheck // deprecated and to be removed
)

// ConsensusVersion defines the current x/group module consensus version.
const ConsensusVersion = 2

var (
	_ module.AppModuleBasic      = AppModule{}
	_ module.AppModuleSimulation = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasEndBlocker = AppModule{}
)

type AppModule struct {
	AppModuleBasic
	keeper     keeper.Keeper
	bankKeeper group.BankKeeper
	accKeeper  group.AccountKeeper
	registry   cdctypes.InterfaceRegistry
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, ak group.AccountKeeper, bk group.BankKeeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc, ac: ak.AddressCodec()},
		keeper:         keeper,
		bankKeeper:     bk,
		accKeeper:      ak,
		registry:       registry,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

type AppModuleBasic struct {
	cdc codec.Codec
	ac  address.Codec
}

// Name returns the group module's name.
func (AppModuleBasic) Name() string {
	return group.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the group
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(group.NewGenesisState())
}

// ValidateGenesis performs genesis state validation for the group module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data group.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", group.ModuleName, err)
	}
	return data.Validate()
}

// GetTxCmd returns the transaction commands for the group module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.TxCmd(a.Name(), a.ac)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the group module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *gwruntime.ServeMux) {
	if err := group.RegisterQueryHandlerClient(context.Background(), mux, group.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers the group module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	group.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec registers the group module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	group.RegisterLegacyAminoCodec(cdc)
}

// InitGenesis performs genesis initialization for the group module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	am.keeper.InitGenesis(ctx, cdc, data)
}

// ExportGenesis returns the exported genesis state as raw bytes for the group
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx, cdc)
	return cdc.MustMarshalJSON(gs)
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	group.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	group.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := keeper.NewMigrator(am.keeper)
	if err := cfg.RegisterMigration(group.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", group.ModuleName, err))
	}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// EndBlock implements the group module's EndBlock.
func (am AppModule) EndBlock(ctx context.Context) error {
	c := sdk.UnwrapSDKContext(ctx)
	return EndBlocker(c, am.keeper)
}

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the group module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for group module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[group.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
// migrate to WeightedOperationsX. This method is ignored when WeightedOperationsX exists and will be removed in the future
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		am.registry,
		simState.AppParams, simState.Cdc, simState.TxConfig,
		am.accKeeper, am.bankKeeper, am.keeper, am.cdc,
	)
}

// WeightedOperationsX registers weighted group module operations for simulation.
func (am AppModule) WeightedOperationsX(weights simsx.WeightSource, reg simsx.Registry) {
	s := simulation.NewSharedState()
	// note: using old keys for backwards compatibility
	reg.Add(weights.Get("msg_create_group", 100), simulation.MsgCreateGroupFactory())
	reg.Add(weights.Get("msg_update_group_admin", 5), simulation.MsgUpdateGroupAdminFactory(am.keeper, s))
	reg.Add(weights.Get("msg_update_group_metadata", 5), simulation.MsgUpdateGroupMetadataFactory(am.keeper, s))
	reg.Add(weights.Get("msg_update_group_members", 5), simulation.MsgUpdateGroupMembersFactory(am.keeper, s))
	reg.Add(weights.Get("msg_create_group_account", 50), simulation.MsgCreateGroupPolicyFactory(am.keeper, s))
	reg.Add(weights.Get("msg_create_group_with_policy", 50), simulation.MsgCreateGroupWithPolicyFactory())
	reg.Add(weights.Get("msg_update_group_account_admin", 5), simulation.MsgUpdateGroupPolicyAdminFactory(am.keeper, s))
	reg.Add(weights.Get("msg_update_group_account_decision_policy", 5), simulation.MsgUpdateGroupPolicyDecisionPolicyFactory(am.keeper, s))
	reg.Add(weights.Get("msg_update_group_account_metadata", 5), simulation.MsgUpdateGroupPolicyMetadataFactory(am.keeper, s))
	reg.Add(weights.Get("msg_submit_proposal", 2*90), simulation.MsgSubmitProposalFactory(am.keeper, s))
	reg.Add(weights.Get("msg_withdraw_proposal", 20), simulation.MsgWithdrawProposalFactory(am.keeper, s))
	reg.Add(weights.Get("msg_vote", 90), simulation.MsgVoteFactory(am.keeper, s))
	reg.Add(weights.Get("msg_exec", 90), simulation.MsgExecFactory(am.keeper, s))
	reg.Add(weights.Get("msg_leave_group", 5), simulation.MsgLeaveGroupFactory(am.keeper, s))
}

//
// App Wiring Setup
//

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type GroupInputs struct {
	depinject.In

	Config           *modulev1.Module
	Key              *store.KVStoreKey
	Cdc              codec.Codec
	AccountKeeper    group.AccountKeeper
	BankKeeper       group.BankKeeper
	Registry         cdctypes.InterfaceRegistry
	MsgServiceRouter baseapp.MessageRouter
}

type GroupOutputs struct {
	depinject.Out

	GroupKeeper keeper.Keeper
	Module      appmodule.AppModule
}

func ProvideModule(in GroupInputs) GroupOutputs {
	/*
		Example of setting group params:
		in.Config.MaxMetadataLen = 1000
		in.Config.MaxExecutionPeriod = "1209600s"
	*/

	k := keeper.NewKeeper(in.Key, in.Cdc, in.MsgServiceRouter, in.AccountKeeper, group.Config{MaxExecutionPeriod: in.Config.MaxExecutionPeriod.AsDuration(), MaxMetadataLen: in.Config.MaxMetadataLen})
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.Registry)
	return GroupOutputs{GroupKeeper: k, Module: m}
}
