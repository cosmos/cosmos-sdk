package app

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Handler struct {
	// Genesis
	InitGenesis     func(context.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate
	DefaultGenesis  func(codec.JSONCodec) json.RawMessage
	ValidateGenesis func(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error
	ExportGenesis   func(sdk.Context, codec.JSONCodec) json.RawMessage

	// ABCI
	BeginBlocker  func(context.Context, abci.RequestBeginBlock)
	EndBlocker    func(context.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
	MsgServices   []ServiceImpl
	QueryServices []ServiceImpl

	// CLI
	QueryCommand *cobra.Command
	TxCommand    *cobra.Command
}

type ServiceImpl struct {
	Desc *grpc.ServiceDesc
	Impl interface{}
}
