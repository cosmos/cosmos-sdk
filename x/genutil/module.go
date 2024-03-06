package genutil

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/genesis"
	"cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var (
	_ module.HasName        = AppModule{}
	_ module.HasABCIGenesis = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

// AppModule implements an application module for the genutil module.
type AppModule struct {
	accountKeeper    types.AccountKeeper
	stakingKeeper    types.StakingKeeper
	deliverTx        genesis.TxHandler
	txEncodingConfig client.TxEncodingConfig
	genTxValidator   types.MessageValidator
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	accountKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper,
	deliverTx genesis.TxHandler,
	txEncodingConfig client.TxEncodingConfig,
	genTxValidator types.MessageValidator,
) module.AppModule {
	return AppModule{
		accountKeeper:    accountKeeper,
		stakingKeeper:    stakingKeeper,
		deliverTx:        deliverTx,
		txEncodingConfig: txEncodingConfig,
		genTxValidator:   genTxValidator,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the genutil module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the genutil module.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the genutil module.
func (b AppModule) ValidateGenesis(cdc codec.JSONCodec, txEncodingConfig client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return types.ValidateGenesis(&data, txEncodingConfig.TxJSONDecoder(), b.genTxValidator)
}

// InitGenesis performs genesis initialization for the genutil module.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	validators, err := InitGenesis(ctx, am.stakingKeeper, am.deliverTx, genesisState, am.txEncodingConfig)
	if err != nil {
		panic(err)
	}
	return validators
}

// ExportGenesis returns the exported genesis state as raw bytes for the genutil module.
func (am AppModule) ExportGenesis(_ context.Context, cdc codec.JSONCodec) json.RawMessage {
	return am.DefaultGenesis(cdc)
}

// GenTxValidator returns the genutil module's genesis transaction validator.
func (am AppModule) GenTxValidator() types.MessageValidator {
	return am.genTxValidator
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// RegisterInterfaces implements module.AppModule.
func (AppModule) RegisterInterfaces(registry.LegacyRegistry) {}
