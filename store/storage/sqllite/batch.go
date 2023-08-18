package sqllite

import (
	"database/sql"
	"fmt"

	"cosmossdk.io/store/v2"
)

var _ store.Batch = (*Batch)(nil)

type batchAction int

const (
	batchActionSet batchAction = 0
	batchActionDel batchAction = 1
)

type batchOp struct {
	action     batchAction
	storeKey   string
	key, value []byte
}

type Batch struct {
	tx      *sql.Tx
	ops     []batchOp
	size    int
	version uint64
}

func NewBatch(storage *sql.DB, version uint64) (*Batch, error) {
	tx, err := storage.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL transaction: %w", err)
	}

	return &Batch{
		tx:      tx,
		ops:     make([]batchOp, 0),
		version: version,
	}, nil
}

func (b *Batch) Size() int {
	return b.size
}

func (b *Batch) Reset() {
	b.ops = nil
	b.ops = make([]batchOp, 0)
	b.size = 0
}

func (b *Batch) Set(storeKey string, key, value []byte) error {
	b.size += len(key) + len(value)
	b.ops = append(b.ops, batchOp{action: batchActionSet, storeKey: storeKey, key: key, value: value})
	return nil
}

func (b *Batch) Delete(storeKey string, key []byte) error {
	b.size += len(key)
	b.ops = append(b.ops, batchOp{action: batchActionDel, storeKey: storeKey, key: key})
	return nil
}

func (b *Batch) Write() error {
	// for _, op := range b.ops {
	// 	switch op.action {
	// 	case batchActionSet:
	// 		 b.tx.

	// 	case batchActionDel:
	// 	}
	// }

	if err := b.tx.Commit(); err != nil {
		return fmt.Errorf("failed to write SQL transaction: %w", err)
	}

	return nil
}
