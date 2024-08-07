package mock

import (
	"crypto/sha256"
	"fmt"

	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"

	
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/sqlite"
	"cosmossdk.io/store/v2/commitment/iavl"
)

type MockStore struct {
	Storage  storev2.VersionedDatabase
	Commiter storev2.Committer
}

func NewMockStorage(logger log.Logger, dir string) storev2.VersionedDatabase {
	storageDB, err := sqlite.New(dir)
	fmt.Println("storageDB", storageDB, err)
	ss := storage.NewStorageStore(storageDB, logger)
	return ss
}

func NewMockCommiter(logger log.Logger, actors ...string) storev2.Committer {
	treeMap := make(map[string]commitment.Tree)
	for _, actor := range actors {
		tree := iavl.NewIavlTree(dbm.NewMemDB(), logger, iavl.DefaultConfig())
		treeMap[actor] = tree
	}
	sc, _ := commitment.NewCommitStore(treeMap, dbm.NewMemDB(), logger)
	return sc
}

func NewMockStore(ss storev2.VersionedDatabase, sc storev2.Committer) *MockStore {
	return &MockStore{Storage: ss, Commiter: sc}
}

func (s *MockStore) GetLatestVersion() (uint64, error) {
	lastCommitID, err := s.LastCommitID()
	if err != nil {
		return 0, err
	}

	return lastCommitID.Version, nil
}

func (s *MockStore) StateLatest() (uint64, corestore.ReaderMap, error) {
	v, err := s.GetLatestVersion()
	if err != nil {
		return 0, nil, err
	}

	return v, NewMockReaderMap(v, s), nil
}

func (s *MockStore) Commit(changeset *corestore.Changeset) (corestore.Hash, error) {
	v, _, _ := s.StateLatest()
	err := s.Storage.ApplyChangeset(v, changeset)
	if err != nil {
		return []byte{}, err
	}

	err = s.Commiter.WriteChangeset(changeset)
	if err != nil {
		return []byte{}, err
	}

	commitInfo, err := s.Commiter.Commit(v+1)
	fmt.Println("commitInfo", commitInfo, err)
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
	v, err := s.GetStateCommitment().GetLatestVersion()
	bz := sha256.Sum256([]byte{})
	return proof.CommitID{
		Version: v,
		Hash:    bz[:],
	}, err
}

func (s *MockStore) SetInitialVersion(v uint64) error {
	return s.Commiter.SetInitialVersion(v)
}

func (s *MockStore) WorkingHash(changeset *corestore.Changeset) (corestore.Hash, error) {
	v, _, _ := s.StateLatest()
	err := s.Storage.ApplyChangeset(v, changeset)
	if err != nil {
		return []byte{}, err
	}

	err = s.Commiter.WriteChangeset(changeset)
	if err != nil {
		return []byte{}, err
	}
	return []byte{}, nil
}


