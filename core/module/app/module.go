package app

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Module interface {
	RegisterTypes(codectypes.InterfaceRegistry)

	InitGenesis(sdk.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis(sdk.Context, codec.JSONCodec) json.RawMessage

	RegisterMsgServices(grpc.ServiceRegistrar)
	RegisterQueryServices(grpc.ServiceRegistrar)
}

type BeginBlocker interface {
	Module

	BeginBlock(sdk.Context, abci.RequestBeginBlock)
}

type EndBlocker interface {
	Module

	EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
}

type HasTxMiddleware interface {
	Module

	RegisterTxMiddleware(registrar TxMiddlewareRegistrar)
}

type TxMiddlewareRegistrar interface {
	RegisterHandler(handler TxMiddlewareHandler)
}
