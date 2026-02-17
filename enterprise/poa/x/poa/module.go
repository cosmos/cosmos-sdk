// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package poa

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/client/cli"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.HasABCIEndBlock  = AppModule{}
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasServices      = AppModule{}
	_ appmodule.AppModule     = AppModule{}
	_ module.HasABCIGenesis   = AppModule{}
	_ module.HasGenesisBasics = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module for the POA module.
type AppModuleBasic struct {
	pubkeyFactory map[string]func(codec.Codec, []byte) *codectypes.Any
}

// Name returns the module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the POA module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers the module's interface types with the interface registry.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec registers the module's types with the Amino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// GetTxCmd returns the root transaction command for the POA module.
func (ab AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCommand(ab.pubkeyFactory)
}

// DefaultGenesis returns the default genesis state for the POA module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the POA module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return gs.ValidateBasic()
}

// AppModule implements the AppModule interface for the POA module.
type AppModule struct {
	AppModuleBasic

	cdc    codec.BinaryCodec
	keeper *keeper.Keeper
}

// ModuleOption allows modifying the module's default fields.
type ModuleOption func(app *AppModule)

// WithPubkeyFactory allows for injecting custom pubkey derivations for the PoA module.
// This is used specifically for the create-validator tx command.
//
// Pubkey factories will take the pubkey type name (i.e. secp256k1, ed25519) and return a function that constructs
// the underlying key for that type given the pk bytes, and returns the Any representation using the supplied codec.
//
// The PoA module will default with a factory that builds keys for secp256k1 and ed25519.
// Applications with custom key types, such as cosmos/evm's ethsecp256k1,
// should use this option to make this module compatible with their application.
func WithPubkeyFactory(f map[string]func(codec.Codec, []byte) *codectypes.Any) ModuleOption {
	return func(app *AppModule) {
		app.pubkeyFactory = f
	}
}

// WithSecp256k1Support adds secp256k1 pubkey support to the PoA module.
// This is needed when you want to create validators with secp256k1 keys.
// IMPORTANT: You must also enable secp256k1 in the consensus params by setting
// consensus.params.validator.pub_key_types to include "secp256k1" in your genesis.
func WithSecp256k1Support() ModuleOption {
	return func(appModule *AppModule) {
		appModule.pubkeyFactory[string(hd.Secp256k1Type)] = func(cdc codec.Codec, bz []byte) *codectypes.Any {
			pubKey := &secp256k1.PubKey{
				Key: bz,
			}
			anyPK, err := codectypes.NewAnyWithValue(pubKey)
			if err != nil {
				panic(err)
			}

			return anyPK
		}
	}
}

// defaultPubkeyFactory is the default factory for the PoA module.
var defaultPubkeyFactory = map[string]func(codec.Codec, []byte) *codectypes.Any{
	string(hd.Ed25519Type): func(cdc codec.Codec, bz []byte) *codectypes.Any {
		pubKey := &ed25519.PubKey{
			Key: bz,
		}
		anyPK, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			panic(err)
		}
		return anyPK
	},
}

// NewAppModule constructs a new PoA module.
// If your application has custom pubkey types, please use the WithPubkeyFactory to supply your own Pubkey factory.
// This is needed to make the create-validator command compatible with your app's custom pubkeys.
func NewAppModule(cdc codec.BinaryCodec, poaKeeper *keeper.Keeper, opts ...ModuleOption) AppModule {
	am := AppModule{
		AppModuleBasic: AppModuleBasic{
			pubkeyFactory: defaultPubkeyFactory,
		},
		cdc:    cdc,
		keeper: poaKeeper,
	}
	for _, opt := range opts {
		opt(&am)
	}
	return am
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (AppModule) IsOnePerModuleType() {}

// RegisterServices registers module services with the configurator.
func (m AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(m.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), m.keeper)
}

// ExportGenesis exports the module's genesis state as JSON.
// It delegates to the keeper's ExportGenesis method to retrieve the current state.
func (m AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs, err := m.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}

	return cdc.MustMarshalJSON(gs)
}

// InitGenesis initializes the module's state from a JSON-encoded genesis state.
// It unmarshals the genesis state and delegates to the keeper's InitGenesis method
// to initialize the store. Returns validator updates that should be applied to consensus.
func (m AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) []abci.ValidatorUpdate {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesis)

	updates, err := m.keeper.InitGenesis(ctx, m.cdc, &genesis)
	if err != nil {
		panic(err)
	}

	return updates
}

// EndBlock returns validator updates to CometBFT at the end of each block.
func (m AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return m.keeper.EndBlocker(sdkCtx)
}
