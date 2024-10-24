package wrapper

import (
	tdbm "github.com/cometbft/cometbft-db"
	cdbm "github.com/cosmos/cosmos-db"
)

type CosmosDBWrapper struct {
	db cdbm.DB
}

func NewCosmosDB(db cdbm.DB) cdbm.DB {
	return &CosmosDBWrapper{db: db}
}

func (db *CosmosDBWrapper) Get(key []byte) ([]byte, error) {
	return db.db.Get(key)
}

func (db *CosmosDBWrapper) Has(key []byte) (bool, error) {
	return db.db.Has(key)
}

func (db *CosmosDBWrapper) Set(key []byte, value []byte) error {
	return db.db.Set(key, value)
}

func (db *CosmosDBWrapper) SetSync(key []byte, value []byte) error {
	return db.db.SetSync(key, value)
}

func (db *CosmosDBWrapper) Delete(key []byte) error {
	return db.db.Delete(key)
}

func (db *CosmosDBWrapper) DeleteSync(key []byte) error {
	return db.db.DeleteSync(key)
}

func (db *CosmosDBWrapper) Iterator(start, end []byte) (cdbm.Iterator, error) {
	it, err := db.db.Iterator(start, end)
	return it.(cdbm.Iterator), err
}

func (db *CosmosDBWrapper) ReverseIterator(start, end []byte) (cdbm.Iterator, error) {
	it, err := db.db.ReverseIterator(start, end)
	return it.(cdbm.Iterator), err
}

func (db *CosmosDBWrapper) NewBatch() cdbm.Batch {
	return NewCosmosBatch(db.db.NewBatch())
}

// NewBatchWithSize(int) Batch

func (db *CosmosDBWrapper) NewBatchWithSize(size int) cdbm.Batch {
	return NewCosmosBatch(db.db.NewBatch())
}

func (db *CosmosDBWrapper) Print() error {
	return db.db.Print()
}

func (db *CosmosDBWrapper) Stats() map[string]string {
	return db.db.Stats()
}

func (db *CosmosDBWrapper) Close() error {
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
