package app

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
)

type Handler struct {
	ID            ModuleID
	InitGenesis   func(context.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate
	BeginBlocker  func(context.Context, abci.RequestBeginBlock)
	EndBlocker    func(context.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
	MsgServices   []ServiceImpl
	QueryServices []ServiceImpl
}

type ServiceImpl struct {
	Desc *grpc.ServiceDesc
	Impl interface{}
}
