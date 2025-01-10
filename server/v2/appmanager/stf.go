package appmanager

import (
	"context"
	"iter"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

// StateTransitionFunction is an interface for processing transactions and blocks.
type StateTransitionFunction[T transaction.Tx] interface {
	// DeliverBlock executes a block of transactions.
	DeliverBlock(
		ctx context.Context,
		block *server.BlockRequest[T],
		state store.ReaderMap,
	) (blockResult *server.BlockResponse, newState store.WriterMap, err error)

	// ValidateTx validates a transaction.
	ValidateTx(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		tx T,
	) server.TxResult

	// Simulate executes a transaction in simulation mode.
	Simulate(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		tx T,
	) (server.TxResult, store.WriterMap)

	// Query executes a query on the application.
	Query(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		req transaction.Msg,
	) (transaction.Msg, error)

	// DeliverSims provides an interface for state transitions by sims.
	DeliverSims(
		ctx context.Context,
		block *server.BlockRequest[T],
		state store.ReaderMap,
		simsBuilder func(ctx context.Context) iter.Seq[T],
	) (blockResult *server.BlockResponse, newState store.WriterMap, err error)
}
