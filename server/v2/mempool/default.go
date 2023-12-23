package mempool

import (
	"context"
	"sync"

	"cosmossdk.io/server/v2/core/transaction"
)

var _ Mempool[transaction.Tx] = (*Default[transaction.Tx])(nil)

type Default[T transaction.Tx] struct {
	mu   *sync.Mutex
	list []Tx[T]
}

func (d Default[T]) Push(_ context.Context, txs []Tx[T]) error {
	d.mu.Lock()
	d.list = append(d.list, txs...)
	d.mu.Unlock()
	return nil
}

func (d Default[T]) Pull(_ context.Context, num int) ([]Tx[T], error) {
	d.mu.Lock()
	num = min(num, len(d.list))
	pulledTx := d.list[:num]
	d.list = d.list[num:]
	d.mu.Unlock()
	return pulledTx, nil
}

func (d Default[T]) Count(_ context.Context) (int, error) {
	d.mu.Lock()
	num := len(d.list)
	d.mu.Unlock()
	return num, nil
}
