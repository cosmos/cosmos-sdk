package wrapper

import (
	tdbm "github.com/cometbft/cometbft-db"
	cdbm "github.com/cosmos/cosmos-db"
)

type DBWrapper struct {
	db tdbm.DB
}

func NewCosmosDB(db tdbm.DB) cdbm.DB {
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

func (db *DBWrapper) Iterator(start, end []byte) (cdbm.Iterator, error) {
	it, err := db.db.Iterator(start, end)
	return it.(cdbm.Iterator), err
}

func (db *DBWrapper) ReverseIterator(start, end []byte) (cdbm.Iterator, error) {
	it, err := db.db.ReverseIterator(start, end)
	return it.(cdbm.Iterator), err
}

func (db *DBWrapper) NewBatch() cdbm.Batch {
	return NewCosmosBatch(db.db.NewBatch())
}

func (db *DBWrapper) NewBatchWithSize(size int) cdbm.Batch {
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
}

func NewCosmosBatch(batch tdbm.Batch) cdbm.Batch {
	return &BatchWrapper{batch: batch}
}

func (b *BatchWrapper) Set(key, value []byte) error {
	return b.batch.Set(key, value)
}

func (b *BatchWrapper) Delete(key []byte) error {
	return b.batch.Delete(key)
}

func (b *BatchWrapper) Write() error {
	return b.batch.Write()
}

func (b *BatchWrapper) WriteSync() error {
	return b.batch.WriteSync()
}

func (b *BatchWrapper) Close() error {
	return b.batch.Close()
}

func (b *BatchWrapper) GetByteSize() (int, error) {
	return 0, nil
}
