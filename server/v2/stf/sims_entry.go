package stf

import (
	"context"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"iter"
)

func (s STF[T]) doSimsTXs(simsBuilder func(ctx context.Context) iter.Seq[T]) doInBlockDeliveryFn[T] {
	return func(
		exCtx context.Context,
		_ []T,
		newState store.WriterMap,
		hi header.Info,
	) ([]server.TxResult, error) {
		simsCtx := context.WithValue(exCtx, "sims.header.time", hi.Time) // using string key to decouple
		var results []server.TxResult
		var i int32
		for tx := range simsBuilder(simsCtx) {
			if err := isCtxCancelled(simsCtx); err != nil {
				return nil, err
			}
			results = append(results, s.deliverTx(simsCtx, newState, tx, transaction.ExecModeFinalize, hi, i+1))
			i++
		}
		return results, nil
	}
}
