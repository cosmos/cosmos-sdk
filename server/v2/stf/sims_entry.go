package stf

import (
	"context"
	"iter"

	"cosmossdk.io/core/header"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

// doSimsTXs constructs a function to simulate transactions in a block execution context using the provided simsBuilder.
func (s STF[T]) doSimsTXs(simsBuilder func(ctx context.Context) iter.Seq[T]) doInBlockDeliveryFn[T] {
	return func(
		exCtx context.Context,
		_ []T,
		newState store.WriterMap,
		headerInfo header.Info,
	) ([]server.TxResult, error) {
		const key = "sims.header.time"
		simsCtx := context.WithValue(exCtx, key, headerInfo.Time) //nolint: staticcheck // using string key to decouple
		var results []server.TxResult
		var i int32
		for tx := range simsBuilder(simsCtx) {
			if err := isCtxCancelled(simsCtx); err != nil {
				return nil, err
			}
			results = append(results, s.deliverTx(simsCtx, newState, tx, transaction.ExecModeFinalize, headerInfo, i+1))
			i++
		}
		return results, nil
	}
}
