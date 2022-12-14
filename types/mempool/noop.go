package mempool

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Mempool = (*NoOpMempool)(nil)

// NoOpMempool defines a no-op mempool. Transactions are completely discarded and
// ignored when BaseApp interacts with the mempool.
type NoOpMempool struct{}

func (m NoOpMempool) Insert(context.Context, sdk.Tx) error      { return nil }
func (m NoOpMempool) Select(context.Context, [][]byte) Iterator { return nil }
func (m NoOpMempool) CountTx() int                              { return 0 }
func (m NoOpMempool) Remove(sdk.Tx) error                       { return nil }
