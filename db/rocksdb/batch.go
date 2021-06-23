package rocksdb

import (
	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/tecbot/gorocksdb"
)

type rocksDBBatch struct {
	db    dbCxn
	batch *gorocksdb.WriteBatch
	opts  *dbOptions
}

var _ dbm.DBWriter = (*rocksDBBatch)(nil)

func (mgr *dbManager) newRocksDBBatch() *rocksDBBatch {
	return &rocksDBBatch{
		db:    mgr.current,
		batch: gorocksdb.NewWriteBatch(),
		opts:  &mgr.opts,
	}
}

// Set implements DBWriter.
func (b *rocksDBBatch) Set(key, value []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	if value == nil {
		return dbm.ErrValueNil
	}
	if b.batch == nil {
		return dbm.ErrBatchClosed
	}
	b.batch.Put(key, value)
	return nil
}

// Delete implements DBWriter.
func (b *rocksDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	if b.batch == nil {
		return dbm.ErrBatchClosed
	}
	b.batch.Delete(key)
	return nil
}

// Write implements DBWriter.
func (b *rocksDBBatch) Commit() error {
	if b.batch == nil {
		return dbm.ErrBatchClosed
	}
	// TODO: expose TransactionDB::Write
	// err := b.db.db.Write(b.opts.wo, b.batch)
	// if err != nil {
	// 	return err
	// }
	// Make sure batch cannot be used afterwards.
	b.Discard()
	return nil
}

// Close implements DBWriter.
func (b *rocksDBBatch) Discard() {
	if b.batch != nil {
		b.batch.Destroy()
		b.batch = nil
	}
}
