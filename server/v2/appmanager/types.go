package appmanager

import (
	"context"

	appmanager "cosmossdk.io/core/app"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

type StateTransitionFunction[T transaction.Tx] interface {
	DeliverBlock(
		ctx context.Context,
		block *appmanager.BlockRequest[T],
		state store.ReaderMap,
	) (blockResult *appmanager.BlockResponse, newState store.WriterMap, err error)

	validateTx(
		ctx context.Context,
		state store.WriterMap,
		gasLimit uint64,
		tx T,
	) (gasUsed uint64, events []event.Event, err error)

	Simulate(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		tx T,
	) (appmanager.TxResult, store.WriterMap)
}
