package mock

import (
	"fmt"

	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"

	// ammstore "cosmossdk.io/server/v2/appmanager/store"
	"cosmossdk.io/server/v2/stf/branch"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/pebbledb"
)

type MockStore struct {
	Storage  storev2.VersionedDatabase
	Commiter storev2.Committer
}

func NewMockStorage(logger log.Logger) storev2.VersionedDatabase {
	storageDB, _ := pebbledb.New("dir")
	ss := storage.NewStorageStore(storageDB, logger)
	return ss
}

func NewMockCommiter(logger log.Logger) storev2.Committer {
	sc, _ := commitment.NewCommitStore(map[string]commitment.Tree{}, dbm.NewMemDB(), logger)
	return sc
}

func NewMockStore(ss storev2.VersionedDatabase, sc storev2.Committer) *MockStore {
	return &MockStore{Storage: ss, Commiter: sc}
}

func (s *MockStore) GetLatestVersion() (uint64, error) {
	v, err := s.Storage.GetLatestVersion()
	return v, err
}

func (s *MockStore) StateLatest() (uint64, corestore.ReaderMap, error) {
	v, err := s.GetLatestVersion()
	if err != nil {
		return 0, nil, err
	}

	return v, NewMockReaderMap(v, s), nil
}

func (s *MockStore) Commit(changeset *corestore.Changeset) (corestore.Hash, error) {
	_, state, _ := s.StateLatest()
	writer := branch.DefaultNewWriterMap(state)
	err := writer.ApplyStateChanges(changeset.Changes)
	return []byte{}, err
}

func (s *MockStore) StateAt(version uint64) (corestore.ReaderMap, error) {
	info, err := s.Commiter.GetCommitInfo(version)
	if err != nil || info == nil {
		return nil, fmt.Errorf("failed to get commit info for version %d: %w", version, err)
	}
	return NewMockReaderMap(version, s), nil
}

func (s *MockStore) GetStateStorage() storev2.VersionedDatabase {
	return s.Storage
}

func (s *MockStore) GetStateCommitment() storev2.Committer {
	return s.Commiter
}

type Result struct {
	key      []byte
	value    []byte
	version  uint64
	proofOps []proof.CommitmentOp
}

func (s *MockStore) Query(storeKey []byte, version uint64, key []byte, prove bool) (storev2.QueryResult, error) {
	state, err := s.StateAt(version)
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
	v, _, err := s.StateLatest()
	return proof.CommitID{
		Version: v,
		Hash:    []byte{},
	}, err
}

func (s *MockStore) SetInitialVersion(v uint64) error {
	return s.Commiter.SetInitialVersion(v)
}

func (s *MockStore) WorkingHash(changeset *corestore.Changeset) (corestore.Hash, error) {
	return s.Commiter.SetInitialVersion(v)
}


