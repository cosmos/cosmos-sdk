package types

import "github.com/cosmos/cosmos-sdk/baseapp"

// core function variable which application runs for transactions
type Handler func(ctx baseapp.Context, msg Msg) baseapp.Result

// core function variable which application runs to handle fees
type FeeHandler func(ctx baseapp.Context, tx Tx, fee Coins)

// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx baseapp.Context, tx Tx) (newCtx baseapp.Context, result baseapp.Result, abort bool)
