//go:build rocksdb_build

package rocksdb

import (
	"sync/atomic"

	"github.com/cosmos/cosmos-sdk/db"
	dbutil "github.com/cosmos/cosmos-sdk/db/internal"
	"github.com/cosmos/gorocksdb"
)

type rocksDBBatch struct {
	batch *gorocksdb.WriteBatch
	mgr   *dbManager
}

var _ db.Writer = (*rocksDBBatch)(nil)

func (mgr *dbManager) newRocksDBBatch() *rocksDBBatch {
	return &rocksDBBatch{
		batch: gorocksdb.NewWriteBatch(),
		mgr:   mgr,
	}
}

// Set implements Writer.
func (b *rocksDBBatch) Set(key, value []byte) error {
	if err := dbutil.ValidateKv(key, value); err != nil {
		return err
	}
	if b.batch == nil {
		return db.ErrTransactionClosed
	}
	b.batch.Put(key, value)
	return nil
}

// Delete implements Writer.
func (b *rocksDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	if b.batch == nil {
		return db.ErrTransactionClosed
	}
	b.batch.Delete(key)
	return nil
}

// Write implements Writer.
func (b *rocksDBBatch) Commit() (err error) {
	if b.batch == nil {
		return db.ErrTransactionClosed
	}
	defer func() { err = dbutil.CombineErrors(err, b.Discard(), "Discard also failed") }()
	err = b.mgr.current.Write(b.mgr.opts.wo, b.batch)
	return
}

// Close implements Writer.
func (b *rocksDBBatch) Discard() error {
	if b.batch != nil {
		defer atomic.AddInt32(&b.mgr.openWriters, -1)
		b.batch.Destroy()
		b.batch = nil
	}
	return nil
}
