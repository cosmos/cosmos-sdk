package genutil

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var (
	_ module.AppModuleGenesis = AppModule{}
	_ module.AppModuleBasic   = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the genutil module.
type AppModuleBasic struct{}

// Name returns the genutil module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the genutil module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry) {}

// DefaultGenesis returns default genesis state as raw bytes for the genutil
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the genutil module.
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, txEncodingConfig client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return types.ValidateGenesis(&data, txEncodingConfig.TxJSONDecoder())
}

// RegisterRESTRoutes registers the REST routes for the genutil module.
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the genutil module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {
}

// GetTxCmd returns no root tx command for the genutil module.
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// GetQueryCmd returns no root query command for the genutil module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }

// ____________________________________________________________________________

// AppModule implements an application module for the genutil module.
type AppModule struct {
	AppModuleBasic

	accountKeeper    types.AccountKeeper
	stakingKeeper    types.StakingKeeper
	deliverTx        deliverTxfn
	txEncodingConfig client.TxEncodingConfig
}

// NewAppModule creates a new AppModule object
func NewAppModule(accountKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper, deliverTx deliverTxfn,
	txEncodingConfig client.TxEncodingConfig,
) module.AppModule {

	return module.NewGenesisOnlyAppModule(AppModule{
		AppModuleBasic:   AppModuleBasic{},
		accountKeeper:    accountKeeper,
		stakingKeeper:    stakingKeeper,
		deliverTx:        deliverTx,
		txEncodingConfig: txEncodingConfig,
	})
}

// InitGenesis performs genesis initialization for the genutil module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	validators, err := InitGenesis(ctx, am.stakingKeeper, am.deliverTx, genesisState, am.txEncodingConfig)
	if err != nil {
		panic(err)
	}
	return validators
}

// ExportGenesis returns the exported genesis state as raw bytes for the genutil
// module.
func (am AppModule) ExportGenesis(_ sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return am.DefaultGenesis(cdc)
}
