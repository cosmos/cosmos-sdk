package cometbft

/*

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	ammstore "cosmossdk.io/server/v2/appmanager/store"
	"cosmossdk.io/server/v2/stf/branch"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/storage/pebbledb"
)

type MockStore struct {
	Storage  ammstore.Storage[*pebbledb.Database]
	Commiter commitment.CommitStore
}

func NewMockStore() *MockStore {
	storageDB, _ := pebbledb.New("dir")
	noopLog := log.NewNopLogger()
	// ss := storage.NewStorageStore(storageDB, nil, log.NewNopLogger()) // for store/v2
	ss, _ := ammstore.New(storageDB)
	sc, _ := commitment.NewCommitStore(map[string]commitment.Tree{}, dbm.NewMemDB(), nil, noopLog)
	return &MockStore{Storage: ss, Commiter: *sc}
}

func (s *MockStore) GetLatestVersion() (uint64, error) {
	v, _, err := s.Storage.StateLatest()
	return v, err
}

func (s *MockStore) StateLatest() (uint64, corestore.ReaderMap, error) {
	return s.Storage.StateLatest()
}

func (s *MockStore) Commit(changeset *corestore.Changeset) (corestore.Hash, error) {
	_, state, _ := s.Storage.StateLatest()
	writer := branch.DefaultNewWriterMap(state)
	err := writer.ApplyStateChanges(changeset.Changes)
	return []byte{}, err
}

func (s *MockStore) GetStateStorage() storev2.VersionedDatabase {
	// TODO
	return nil
}

func (s *MockStore) GetStateCommitment() storev2.Committer {
	return &s.Commiter
}

type Result struct {
	key      []byte
	value    []byte
	version  uint64
	proofOps []proof.CommitmentOp
}

func (s *MockStore) Query(storeKey []byte, version uint64, key []byte, prove bool) (storev2.QueryResult, error) {
	state, err := s.Storage.StateAt(version)
	reader, err := state.GetReader(storeKey)
	value, err := reader.Get(key)
	res := storev2.QueryResult{
		Key:     key,
		Value:   value,
		Version: version,
	}
	return res, err
}

func (s *MockStore) LastCommitID() (proof.CommitID, error) {
	v, _, err := s.Storage.StateLatest()
	return proof.CommitID{
		Version: v,
		Hash:    []byte{},
	}, err
}
*/
