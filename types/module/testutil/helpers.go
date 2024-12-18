package testutil

import (
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// TestEncodingConfig defines an encoding configuration that is used for testing
// purposes. Note, MakeTestEncodingConfig takes a series of AppModule types
// which should only contain the relevant module being tested and any potential
// dependencies.
type TestEncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

func MakeTestEncodingConfig(codecOpt testutil.CodecOptions, modules ...module.AppModule) TestEncodingConfig {
	aminoCodec := codec.NewLegacyAmino()
	interfaceRegistry := codecOpt.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	signingCtx := cdc.InterfaceRegistry().SigningContext()

	encCfg := TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             cdc,
		TxConfig:          tx.NewTxConfig(cdc, signingCtx.AddressCodec(), signingCtx.ValidatorAddressCodec(), tx.DefaultSignModes),
		Amino:             aminoCodec,
	}

	mb := module.NewManager(modules...)
	std.RegisterLegacyAminoCodec(encCfg.Amino)
	std.RegisterInterfaces(encCfg.InterfaceRegistry)
	mb.RegisterLegacyAminoCodec(encCfg.Amino)
	mb.RegisterInterfaces(encCfg.InterfaceRegistry)

	return encCfg
}

func MakeTestTxConfig(codecOpt testutil.CodecOptions) client.TxConfig {
	interfaceRegistry := codecOpt.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	signingCtx := interfaceRegistry.SigningContext()
	return tx.NewTxConfig(cdc, signingCtx.AddressCodec(), signingCtx.ValidatorAddressCodec(), tx.DefaultSignModes)
}

type TestBuilderTxConfig struct {
	client.TxConfig
	TxBuilder *TestTxBuilder
}

func MakeBuilderTestTxConfig(codecOpt testutil.CodecOptions) TestBuilderTxConfig {
	return TestBuilderTxConfig{
		TxConfig: MakeTestTxConfig(codecOpt),
	}
}

func (cfg TestBuilderTxConfig) NewTxBuilder() client.TxBuilder {
	if cfg.TxBuilder == nil {
		cfg.TxBuilder = &TestTxBuilder{
			TxBuilder: cfg.TxConfig.NewTxBuilder(),
		}
	}
	return cfg.TxBuilder
}

type TestTxBuilder struct {
	client.TxBuilder
	ExtOptions []*types.Any
}

func (b *TestTxBuilder) SetExtensionOptions(extOpts ...*types.Any) {
	b.ExtOptions = extOpts
}

// CreateMultiStore is a helper for setting up multiple stores for provided modules.
func CreateMultiStore(keys map[string]*storetypes.KVStoreKey, logger log.Logger) storetypes.CommitMultiStore {
	db := coretesting.NewMemDB()
	cms := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())

	for key := range keys {
		cms.MountStoreWithDB(keys[key], storetypes.StoreTypeIAVL, db)
	}

	_ = cms.LoadLatestVersion()
	return cms
}
