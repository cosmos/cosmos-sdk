package app

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/types/tx"
)

type HasTxMiddleware interface {
	Module

	RegisterTxMiddleware(registrar TxMiddlewareRegistrar)
}

type TxMiddlewareRegistrar interface {
	RegisterTxMiddlewareFactory(configType proto.Message, factory TxMiddlewareFactory)
}

type TxMiddlewareFactory func(config interface{}) TxMiddleware

type TxMiddleware interface {
	OnCheckTx(ctx context.Context, tx tx.Tx, req abci.RequestCheckTx, next TxHandler) (abci.ResponseCheckTx, error)
	OnDeliverTx(ctx context.Context, tx tx.Tx, req abci.RequestDeliverTx, next TxHandler) (abci.ResponseDeliverTx, error)
}

type TxHandler interface {
	CheckTx(ctx context.Context, tx tx.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error)
	DeliverTx(ctx context.Context, tx tx.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error)
}
