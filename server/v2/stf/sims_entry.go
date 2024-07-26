package stf

import (
	"context"
	appmanager "cosmossdk.io/core/app"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/stf/internal"
)

func (s STF[T]) DoSimsTXs(simsBuilder func(ctx context.Context) (T, bool)) DoInBlockDeliveryFn[T] {
	return func(
		ctx context.Context,
		_ []T,
		newState store.WriterMap,
		hi header.Info,
	) ([]appmanager.TxResult, error) {
		var results []appmanager.TxResult
		// use exec context so that the msg factories get access to db state in keepers
		exCtx := s.makeContext(ctx, appmanager.ConsensusIdentity, newState, internal.ExecModeFinalize)
		exCtx.setHeaderInfo(hi)

		tx, hasNext := simsBuilder(exCtx)
		_ = hasNext
		//for ; hasNext; tx, hasNext = simsBuilder(ctx) {
		if err := isCtxCancelled(ctx); err != nil {
			return nil, err
		}
		println("+++ delivering msg")
		results = append(results, s.deliverTx(ctx, newState, tx, transaction.ExecModeFinalize, hi))
		//}
		return results, nil
	}
}
