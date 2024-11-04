package sqlite

import (
	"fmt"
	"sync"

	"github.com/bvinc/go-sqlite-lite/sqlite3"

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
	storeKey   []byte
	key, value []byte
}

type Batch struct {
	db      *sqlite3.Conn
	lock    *sync.Mutex
	ops     []batchOp
	size    int
	version int64
}

func NewBatch(db *sqlite3.Conn, writeLock *sync.Mutex, version uint64) (*Batch, error) {
	if version&(1<<63) != 0 {
		return nil, fmt.Errorf("%d too large; uint64 with the highest bit set are not supported", version)
	}
	err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL transaction: %w", err)
	}

	return &Batch{
		db:      db,
		lock:    writeLock,
		ops:     make([]batchOp, 0),
		version: int64(version),
	}, nil
}

func (b *Batch) Size() int {
	return b.size
}

func (b *Batch) Reset() error {
	b.ops = nil
	b.ops = make([]batchOp, 0)
	b.size = 0

	err := b.db.Begin()
	if err != nil {
		return err
	}

	return nil
}

func (b *Batch) Set(storeKey, key, value []byte) error {
	b.size += len(key) + len(value)
	b.ops = append(b.ops, batchOp{action: batchActionSet, storeKey: storeKey, key: key, value: value})
	return nil
}

func (b *Batch) Delete(storeKey, key []byte) error {
	b.size += len(key)
	b.ops = append(b.ops, batchOp{action: batchActionDel, storeKey: storeKey, key: key})
	return nil
}

func (b *Batch) Write() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	err := b.db.Exec(reservedUpsertStmt, reservedStoreKey, keyLatestHeight, b.version, 0, b.version)
	if err != nil {
		return fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	for _, op := range b.ops {
		switch op.action {
		case batchActionSet:
			err := b.db.Exec(upsertStmt, op.storeKey, op.key, op.value, b.version, op.value)
			if err != nil {
				return fmt.Errorf("failed to exec SQL statement: %w", err)
			}

		case batchActionDel:
			err := b.db.Exec(delStmt, b.version, op.storeKey, op.key, b.version)
			if err != nil {
				return fmt.Errorf("failed to exec SQL statement: %w", err)
			}
		}
	}

	if err := b.db.Commit(); err != nil {
		return fmt.Errorf("failed to write SQL transaction: %w", err)
	}

	return b.db.Close()
}
