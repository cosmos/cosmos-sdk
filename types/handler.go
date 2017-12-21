package types

import (
	"github.com/cosmos/cosmos-sdk/store"
)

// Handler handles both ABCI DeliverTx and CheckTx requests.
// Iff ABCI.CheckTx, ctx.IsCheckTx() returns true.
type Handler func(ctx Context, store store.MultiStore, tx Tx) Result
