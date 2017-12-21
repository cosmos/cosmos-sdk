package types

import (
	sdk "github.com/cosmos/cosmos-sdk"
)

// Handler handles both ABCI DeliverTx and CheckTx requests.
// Iff ABCI.CheckTx, ctx.IsCheckTx() returns true.
type Handler func(ctx Context, ms sdk.MultiStore, tx Tx) Result
