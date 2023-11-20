package mempool

import (
	"context"
)

// TODO here until we rebase to get the txcodec items
type TxValidator interface {
	ValidateTx(ctx context.Context, tx any, simulate bool) (context.Context, error)
}

var _ Mempool = (*NoOpMempool)(nil)

type NoOpMempool struct {
	txValidator TxValidator
}

func NewNoopMempool(txv TxValidator) *NoOpMempool {
	return &NoOpMempool{txValidator: txv}
}

func (s *NoOpMempool) Start() error {
	// NoOpMempool does not require any initialization
	return nil
}

func (s *NoOpMempool) Stop() error {
	// NoOpMempool does not require any cleanup
	return nil
}

func (npm NoOpMempool) Insert(ctx context.Context, tx any) error {
	_, err := npm.txValidator.ValidateTx(ctx, tx, false)
	return err
}

func (NoOpMempool) GetTxs(ctx context.Context, size uint32) (any, error) {
	return nil, nil
}

func (NoOpMempool) CountTx() uint32  { return 0 }
func (NoOpMempool) Remove(any) error { return nil }
