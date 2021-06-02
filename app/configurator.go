package app

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
)

type Configurator interface {
	module.Configurator

	RegisterBeginBlocker(func(context.Context, abci.RequestBeginBlock))
	RegisterEndBlocker(func(context.Context, abci.RequestEndBlock) []abci.ValidatorUpdate)
	RegisterGenesisHandler(handler GenesisHandler)
}

type GenesisBasicHandler interface {
	DefaultGenesis(codec.JSONCodec) json.RawMessage
	ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error
}

type GenesisHandler interface {
	GenesisBasicHandler

	InitGenesis(context.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis(context.Context, codec.JSONCodec) json.RawMessage
}
