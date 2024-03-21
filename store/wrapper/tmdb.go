package wrapper

import (
	"encoding/binary"

	tdbm "github.com/cometbft/cometbft-db"
	iavldb "github.com/cosmos/iavl/db"
)

type DBWrapper struct {
	db tdbm.DB
}

func NewIAVLDB(db tdbm.DB) iavldb.DB {
	return &DBWrapper{db: db}
}

func (db *DBWrapper) Get(key []byte) ([]byte, error) {
	return db.db.Get(key)
}

func (db *DBWrapper) Has(key []byte) (bool, error) {
	return db.db.Has(key)
}

func (db *DBWrapper) Set(key []byte, value []byte) error {
	return db.db.Set(key, value)
}

func (db *DBWrapper) SetSync(key []byte, value []byte) error {
	return db.db.SetSync(key, value)
}

func (db *DBWrapper) Delete(key []byte) error {
	return db.db.Delete(key)
}

func (db *DBWrapper) DeleteSync(key []byte) error {
	return db.db.DeleteSync(key)
}

func (db *DBWrapper) Iterator(start, end []byte) (iavldb.Iterator, error) {
	it, err := db.db.Iterator(start, end)
	return it, err
}

func (db *DBWrapper) ReverseIterator(start, end []byte) (iavldb.Iterator, error) {
	it, err := db.db.ReverseIterator(start, end)
	return it, err
}

func (db *DBWrapper) NewBatch() iavldb.Batch {
	return NewCosmosBatch(db.db.NewBatch())
}

func (db *DBWrapper) NewBatchWithSize(size int) iavldb.Batch {
	return NewCosmosBatch(db.db.NewBatch())
}

func (db *DBWrapper) Print() error {
	return db.db.Print()
}

func (db *DBWrapper) Stats() map[string]string {
	return db.db.Stats()
}

func (db *DBWrapper) Close() error {
	return db.db.Close()
}

type BatchWrapper struct {
	batch tdbm.Batch
	size  int
}

func NewCosmosBatch(batch tdbm.Batch) iavldb.Batch {
	return &BatchWrapper{batch: batch}
}

func (b *BatchWrapper) Set(key, value []byte) error {
	b.size += 1 + binary.MaxVarintLen32 + len(key) + binary.MaxVarintLen32 + len(value)
	return b.batch.Set(key, value)
}

func (b *BatchWrapper) Delete(key []byte) error {
	b.size += 1 + binary.MaxVarintLen32 + len(key)
	return b.batch.Delete(key)
}

func (b *BatchWrapper) Write() error {
	b.size = 0
	return b.batch.Write()
}

func (b *BatchWrapper) WriteSync() error {
	b.size = 0
	return b.batch.WriteSync()
}

func (b *BatchWrapper) Close() error {
	return b.batch.Close()
}

func (b *BatchWrapper) GetByteSize() (int, error) {
	return b.size, nil
}
