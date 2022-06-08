package tx

import (
	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/depinject"

	modulev1 "cosmossdk.io/api/cosmos/tx/module/v1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var DefaultTxModuleConfig = TxModuleConfig{
	// TODO set good defaults
}

type TxModuleConfig struct {
	TxConfig    client.TxConfig
	AnteHandler sdk.AnteHandler
	PostHandler sdk.AnteHandler
}

type TxModuleOption func(*TxModuleConfig)

func SetCustomTxConfig(txConfig client.TxConfig) TxModuleOption {
	return func(cfg *TxModuleConfig) {
		cfg.TxConfig = txConfig
	}
}

func SetCustomAnteHandler(anteHandler sdk.AnteHandler) TxModuleOption {
	return func(cfg *TxModuleConfig) {
		cfg.AnteHandler = anteHandler
	}
}

func SetCustomPostHandler(PostHandler sdk.AnteHandler) TxModuleOption {
	return func(cfg *TxModuleConfig) {
		cfg.PostHandler = PostHandler
	}
}

//
// New App Wiring Setup
//

func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(provideModule),
	)
}

type txInputs struct {
	depinject.In

	options []TxModuleOption
}

type txOutputs struct {
	depinject.Out

	TxConfig    client.TxConfig
	AnteHandler sdk.AnteHandler
	PostHandler sdk.AnteHandler
}

func provideModule(in txInputs) txOutputs {
	cfg := DefaultTxModuleConfig
	for _, opt := range in.options {
		opt(&cfg)
	}

	return txOutputs{TxConfig: cfg.TxConfig, AnteHandler: cfg.AnteHandler, PostHandler: cfg.PostHandler}
}
