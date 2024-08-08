package appmanager

import (
	"context"

	appmanager "cosmossdk.io/core/app"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
)

// StateTransitionFunction is an interface for processing transactions and blocks.
type StateTransitionFunction[T transaction.Tx] interface {
	// DeliverBlock executes a block of transactions.
	DeliverBlock(
		ctx context.Context,
		block *appmanager.BlockRequest[T],
		state store.ReaderMap,
	) (blockResult *appmanager.BlockResponse, newState store.WriterMap, err error)

	// ValidateTx validates a transaction.
	ValidateTx(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		tx T,
	) appmanager.TxResult

	// Simulate executes a transaction in simulation mode.
	Simulate(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		tx T,
	) (appmanager.TxResult, store.WriterMap)

	// Query executes a query on the application.
	Query(
		ctx context.Context,
		state store.ReaderMap,
		gasLimit uint64,
		req transaction.Msg,
	) (transaction.Msg, error)

	// RunWithCtx executes the provided closure within a context.
	// TODO: remove
	RunWithCtx(
		ctx context.Context,
		state store.ReaderMap,
		closure func(ctx context.Context) error,
	) (store.WriterMap, error)

	DeliverSims(
		ctx context.Context,
		block *appmanager.BlockRequest[T],
		state store.ReaderMap,
		simsBuilder func(ctx context.Context) (T, bool),
	) (blockResult *appmanager.BlockResponse, newState store.WriterMap, err error)
}
