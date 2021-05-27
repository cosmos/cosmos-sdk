package app

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/tx"
)

type TxMiddlewareFactory func(interface{}) TxMiddleware

type TxMiddleware func(ctx context.Context, tx tx.Tx, next TxMiddleware) error
