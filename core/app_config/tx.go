package app_config

import (
	"context"

	"github.com/cosmos/cosmos-sdk/core"
	"github.com/cosmos/cosmos-sdk/core/module/app"

	"github.com/gogo/protobuf/proto"

	abci "github.com/tendermint/tendermint/abci/types"
)

type TxHandler interface {
	CheckTx(ctx context.Context, req abci.RequestCheckTx) (abci.ResponseCheckTx, error)
	DeliverTx(ctx context.Context, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error)
}

type TxHandlerParams struct {
	MsgRouter MsgRouter
}

type MsgRouter interface {
	RouteMsg(core.Msg) (proto.Message, error)
}

type HasTxHandler interface {
	app.Handler

	TxHandler(TxHandlerParams) TxHandler
}
