package app

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
)

type GenesisHandler interface {
	InitGenesis(context.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis(context.Context, codec.JSONCodec) json.RawMessage
}

type QueryHandler interface {
	RegisterQueryServices(grpc.ServiceRegistrar)
}

type Handler interface {
	QueryHandler

	RegisterMsgServices(grpc.ServiceRegistrar)
}

type BeginBlocker interface {
	BeginBlock(context.Context, abci.RequestBeginBlock)
}

type EndBlocker interface {
	EndBlock(context.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
}
