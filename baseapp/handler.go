package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// core function variable which application runs for transactions
type Handler func(ctx sdk.Context, msg Msg) sdk.Result

// core function variable which application runs to handle fees
type FeeHandler func(ctx sdk.Context, tx Tx, fee Coins)

// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx sdk.Context, tx Tx) (newCtx sdk.Context, result sdk.Result, abort bool)
