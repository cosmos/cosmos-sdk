package mock

import (
	"crypto/sha256"
	"fmt"

	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/proof"
)

type MockStore struct {
	Committer storev2.Committer
}

func NewMockCommiter(logger log.Logger, actors ...string) storev2.Committer {
	treeMap := make(map[string]commitment.Tree)
	for _, actor := range actors {
		tree := iavl.NewIavlTree(dbm.NewMemDB(), logger, iavl.DefaultConfig())
		treeMap[actor] = tree
	}
	sc, _ := commitment.NewCommitStore(treeMap, treeMap, dbm.NewMemDB(), logger)
	return sc
}

func NewMockStore(sc storev2.Committer) *MockStore {
	return &MockStore{Committer: sc}
}

func (s *MockStore) GetLatestVersion() (uint64, error) {
	lastCommitID, err := s.LastCommitID()
	if err != nil {
		return 0, err
	}

	return uint64(lastCommitID.Version), nil
}

func (s *MockStore) StateLatest() (uint64, corestore.ReaderMap, error) {
	v, err := s.GetLatestVersion()
	if err != nil {
		return 0, nil, err
	}

	return v, NewMockReaderMap(v, s), nil
}

func (s *MockStore) Commit(changeset *corestore.Changeset) (corestore.Hash, error) {
	err := s.Committer.WriteChangeset(changeset)
	if err != nil {
		return []byte{}, err
	}

	_, err = s.Committer.Commit(changeset.Version)
	return []byte{}, err
}

func (s *MockStore) StateAt(version uint64) (corestore.ReaderMap, error) {
	info, err := s.Committer.GetCommitInfo(version)
	if err != nil || info == nil {
		return nil, fmt.Errorf("failed to get commit info for version %d: %w", version, err)
	}
	return NewMockReaderMap(version, s), nil
}

func (s *MockStore) GetStateCommitment() storev2.Committer {
	return s.Committer
}

func (s *MockStore) Query(storeKey []byte, version uint64, key []byte, prove bool) (storev2.QueryResult, error) {
	state, err := s.StateAt(version)
	if err != nil {
		return storev2.QueryResult{}, err
	}

	reader, err := state.GetReader(storeKey)
	if err != nil {
		return storev2.QueryResult{}, err
	}

	value, err := reader.Get(key)
	if err != nil {
		return storev2.QueryResult{}, err
	}

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
		Version: int64(v),
		Hash:    bz[:],
	}, err
}

func (s *MockStore) SetInitialVersion(v uint64) error {
	return s.Committer.SetInitialVersion(v)
}
