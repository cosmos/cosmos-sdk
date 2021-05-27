package app

import (
	"context"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

type Module interface {
	RegisterTypes(codectypes.InterfaceRegistry)

	InitGenesis(context.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis(context.Context, codec.JSONCodec) json.RawMessage

	RegisterMsgServices(grpc.ServiceRegistrar)
	RegisterQueryServices(grpc.ServiceRegistrar)
}

type BeginBlocker interface {
	Module

	BeginBlock(context.Context, abci.RequestBeginBlock)
}

type EndBlocker interface {
	Module

	EndBlock(context.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
}
