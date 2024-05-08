package appmanager

import (
	"context"

	appmanager "cosmossdk.io/core/app"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

type StateTransitionFunction[T transaction.Tx] interface {
	DeliverBlock(
		ctx context.Context,
		block *appmanager.BlockRequest[T],
		state store.ReaderMap,
	) (blockResult *appmanager.BlockResponse, newState store.WriterMap, err error)

	ValidateTx(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		tx T,
	) appmanager.TxResult

	Simulate(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		tx T,
	) (appmanager.TxResult, store.WriterMap)

	Query(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		req transaction.Type,
	) (transaction.Type, error)

	RunWithCtx(
		ctx context.Context,
		state store.ReaderMap,
		closure func(ctx context.Context) error,
	) (store.WriterMap, error)
}
