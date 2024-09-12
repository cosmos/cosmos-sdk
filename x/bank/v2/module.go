package bankv2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/x/bank/v2/keeper"
	"cosmossdk.io/x/bank/v2/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// ConsensusVersion defines the current x/bank/v2 module consensus version.
const ConsensusVersion = 1

var (
	_ appmodulev2.AppModule             = AppModule{}
	_ appmodulev2.HasGenesis            = AppModule{}
	_ appmodulev2.HasRegisterInterfaces = AppModule{}
	_ appmodulev2.HasQueryHandlers      = AppModule{}
	_ appmodulev2.HasMsgHandlers        = AppModule{}
)

// AppModule implements an application module for the bank module.
type AppModule struct {
	cdc    codec.Codec
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper *keeper.Keeper) AppModule {
	return AppModule{
		cdc:    cdc,
		keeper: keeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the bank module's name.
// Deprecated: kept for legacy reasons.
func (AppModule) Name() string { return types.ModuleName }

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModule) RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	types.RegisterInterfaces(registrar)
}

// DefaultGenesis returns default genesis state as raw bytes for the bank module.
func (am AppModule) DefaultGenesis() json.RawMessage {
	return am.cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the bank module.
func (am AppModule) ValidateGenesis(bz json.RawMessage) error {
	var data types.GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return data.Validate()
}

// InitGenesis performs genesis initialization for the bank module.
func (am AppModule) InitGenesis(ctx context.Context, data json.RawMessage) error {
	var genesisState types.GenesisState
	if err := am.cdc.UnmarshalJSON(data, &genesisState); err != nil {
		return err
	}

	return am.keeper.InitGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the bank/v2 module.
func (am AppModule) ExportGenesis(ctx context.Context) (json.RawMessage, error) {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		return nil, err
	}

	return am.cdc.MarshalJSON(gs)
}

// RegisterMsgHandlers registers the message handlers for the bank module.
func (am AppModule) RegisterMsgHandlers(router appmodulev2.MsgRouter) {
	handlers := keeper.NewHandlers(am.keeper)

	var errs error
	if err := appmodulev2.RegisterHandler(
		router, gogoproto.MessageName(&types.MsgUpdateParams{}), handlers.MsgUpdateParams,
	); err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		panic(errs)
	}
}

// RegisterQueryHandlers registers the query handlers for the bank module.
func (am AppModule) RegisterQueryHandlers(router appmodulev2.QueryRouter) {
	handlers := keeper.NewHandlers(am.keeper)

	var errs error
	if err := appmodulev2.RegisterHandler(
		router, gogoproto.MessageName(&types.QueryParamsRequest{}), handlers.QueryParams,
	); err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		panic(errs)
	}
}
