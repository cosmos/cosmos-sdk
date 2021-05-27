package app

import (
	"context"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/types/tx"
)

type TxMiddlewareHandler interface {
	ConfigType() proto.Message
	Init(config proto.Message) TxMiddleware
}

type TxMiddleware func(ctx context.Context, tx tx.Tx, next TxMiddleware) error
