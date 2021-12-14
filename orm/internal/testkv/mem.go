package testkv

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

// NewSplitMemIndexCommitmentStore returns an IndexCommitmentStore instance
// which uses two separate memory stores to simulate behavior when there
// are really two separate backing stores.
func NewSplitMemIndexCommitmentStore() kvstore.IndexCommitmentStore {
	sharedMem := NewSharedMemIndexCommitmentStore().(*sharedMemICS)
	return &splitMemICS{sharedMemICS: sharedMem}
}

type splitMemICS struct {
	*sharedMemICS
}

var (
	commitmentPrefix = []byte{0}
	indexPrefix      = []byte{1}
)

func (s splitMemICS) ReadCommitmentStore() kvstore.ReadStore {
	return dbm.NewPrefixDB(s.db, commitmentPrefix)
}

func (s splitMemICS) ReadIndexStore() kvstore.ReadStore {
	return dbm.NewPrefixDB(s.db, indexPrefix)
}

func (s splitMemICS) CommitmentStore() kvstore.Store {
	return dbm.NewPrefixDB(s.writeStore, commitmentPrefix)
}

func (s splitMemICS) IndexStore() kvstore.Store {
	return dbm.NewPrefixDB(s.writeStore, indexPrefix)
}

// NewSharedMemIndexCommitmentStore returns an IndexCommitmentStore instance
// which uses a single backing memory store to simulate legacy scenarios
// where only a single KV-store is available to modules.
func NewSharedMemIndexCommitmentStore() kvstore.IndexCommitmentStore {
	store := dbm.NewMemDB()
	return &sharedMemICS{
		store,
		&writeStore{
			DB:    store,
			batch: store.NewBatch(),
		},
	}
}

type sharedMemICS struct {
	db         dbm.DB
	writeStore *writeStore
}

func (s sharedMemICS) ReadCommitmentStore() kvstore.ReadStore {
	return s.db
}

func (s sharedMemICS) ReadIndexStore() kvstore.ReadStore {
	return s.db
}

func (s sharedMemICS) CommitmentStore() kvstore.Store {
	return s.writeStore
}

func (s sharedMemICS) IndexStore() kvstore.Store {
	return s.writeStore
}

func (s sharedMemICS) Commit() error {
	err := s.writeStore.batch.Write()
	if err != nil {
		return err
	}
	err = s.writeStore.batch.Close()
	s.writeStore.batch = s.db.NewBatch()
	return err
}

func (s sharedMemICS) Rollback() error {
	s.writeStore.batch = s.db.NewBatch()
	return nil
}

type writeStore struct {
	dbm.DB
	batch dbm.Batch
}

func (w writeStore) Set(key, value []byte) error {
	return w.batch.Set(key, value)
}

func (w writeStore) Delete(key []byte) error {
	return w.batch.Delete(key)
}
