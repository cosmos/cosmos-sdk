package app

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
)

type Handler struct {
	BasicGenesisHandler
	InitGenesis   func(context.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis func(context.Context, codec.JSONCodec) json.RawMessage
	BeginBlocker  func(context.Context, abci.RequestBeginBlock)
	EndBlocker    func(context.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
	MsgServices   []ServiceImpl
	QueryServices []ServiceImpl
}

type ServiceImpl struct {
	Desc *grpc.ServiceDesc
	Impl interface{}
}

type BasicGenesisHandler struct {
	DefaultGenesis  func(codec.JSONCodec) json.RawMessage
	ValidateGenesis func(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error
}
