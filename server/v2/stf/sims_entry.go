package stf

import (
	"context"
	appmanager "cosmossdk.io/core/app"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

func (s STF[T]) DoSimsTXs(simsBuilder func(ctx context.Context) (T, bool)) doInBlockDeliveryFn[T] {
	return func(
		exCtx *executionContext,
		_ []T,
		newState store.WriterMap,
		hi header.Info,
	) ([]appmanager.TxResult, error) {
		var results []appmanager.TxResult
		exCtx.Context = context.WithValue(exCtx.Context, "sims.header.time", hi.Time) // using string key to decouple
		// use exec context so that the msg factories get access to db state in keepers
		simsCtx := exCtx
		for tx, exit := simsBuilder(simsCtx); !exit; tx, exit = simsBuilder(simsCtx) {
			if err := isCtxCancelled(exCtx); err != nil {
				return nil, err
			}
			results = append(results, s.deliverTx(exCtx, newState, tx, transaction.ExecModeFinalize, hi))
		}
		return results, nil
	}
}
