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
	return &splitMemICS{shared: sharedMem}
}

var (
	commitmentPrefix = []byte{0}
	indexPrefix      = []byte{1}
)

type splitMemICS struct {
	shared *sharedMemICS
}

type splitMemICSWriter struct {
	dbm.DB
	split splitMemICS
	batch dbm.Batch
}

func (s splitMemICSWriter) CommitmentStoreReader() kvstore.Reader {
	return s.split.CommitmentStoreReader()
}

func (s splitMemICSWriter) IndexStoreReader() kvstore.Reader {
	return s.split.IndexStoreReader()
}

type splitMemWriter struct {
	dbm.DB
	batch  dbm.Batch
	prefix []byte
}

func (s splitMemWriter) Set(key, value []byte) error {
	return s.batch.Set(append(s.prefix, key...), value)
}

func (s splitMemWriter) Delete(key []byte) error {
	return s.batch.Delete(append(s.prefix, key...))
}

func (s splitMemICSWriter) CommitmentStoreWriter() kvstore.Writer {
	return &splitMemWriter{
		DB:     dbm.NewPrefixDB(s.DB, commitmentPrefix),
		batch:  s.batch,
		prefix: commitmentPrefix,
	}
}

func (s splitMemICSWriter) IndexStoreWriter() kvstore.Writer {
	return &splitMemWriter{
		DB:     dbm.NewPrefixDB(s.DB, indexPrefix),
		batch:  s.batch,
		prefix: indexPrefix,
	}
}

func (s splitMemICSWriter) Write() error {
	return s.batch.Write()
}

func (s splitMemICSWriter) Close() {
	err := s.batch.Close()
	if err != nil {
		panic(err)
	}
}

func (s splitMemICS) NewWriter() kvstore.IndexCommitmentStoreWriter {
	return &splitMemICSWriter{
		DB:    s.shared.db,
		batch: s.shared.db.NewBatch(),
		split: s,
	}
}

func (s splitMemICS) CommitmentStoreReader() kvstore.Reader {
	return dbm.NewPrefixDB(s.shared.db, commitmentPrefix)
}

func (s splitMemICS) IndexStoreReader() kvstore.Reader {
	return dbm.NewPrefixDB(s.shared.db, indexPrefix)
}

// NewSharedMemIndexCommitmentStore returns an IndexCommitmentStore instance
// which uses a single backing memory store to simulate legacy scenarios
// where only a single KV-store is available to modules.
func NewSharedMemIndexCommitmentStore() kvstore.IndexCommitmentStore {
	store := dbm.NewMemDB()
	return &sharedMemICS{store}
}

type sharedMemICS struct {
	db dbm.DB
}

type sharedMemICSWriter struct {
	dbm.DB
	batch dbm.Batch
}

func (s sharedMemICSWriter) CommitmentStoreReader() kvstore.Reader {
	return s.DB
}

func (s sharedMemICSWriter) IndexStoreReader() kvstore.Reader {
	return s.DB
}

func (s sharedMemICSWriter) Set(key, value []byte) error {
	return s.batch.Set(key, value)
}

func (s sharedMemICSWriter) Delete(key []byte) error {
	return s.batch.Delete(key)
}

func (s sharedMemICSWriter) CommitmentStoreWriter() kvstore.Writer {
	return s
}

func (s sharedMemICSWriter) IndexStoreWriter() kvstore.Writer {
	return s
}

func (s sharedMemICSWriter) Write() error {
	return s.batch.Write()
}

func (s sharedMemICSWriter) Close() {
	err := s.batch.Close()
	if err != nil {
		panic(err)
	}
}

func (s sharedMemICS) NewWriter() kvstore.IndexCommitmentStoreWriter {
	return &sharedMemICSWriter{
		DB:    s.db,
		batch: s.db.NewBatch(),
	}
}

func (s sharedMemICS) CommitmentStoreReader() kvstore.Reader {
	return s.db
}

func (s sharedMemICS) IndexStoreReader() kvstore.Reader {
	return s.db
}
