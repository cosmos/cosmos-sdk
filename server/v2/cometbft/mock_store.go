package cometbft

import (
	"cosmossdk.io/log"
	ammstore "cosmossdk.io/server/v2/appmanager/store"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/commitment"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/storage/pebbledb"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/store/v2/proof"
	cmt "cosmossdk.io/server/v2/cometbft/types"
	ics23 "github.com/cosmos/ics23/go"
)

type MockStore struct {
	Storage ammstore.Storage[*pebbledb.Database]
	Commiter commitment.CommitStore
}

func NewMockStore() *MockStore {
	storageDB, _ := pebbledb.New("dir")
	noopLog := log.NewNopLogger()
	ss, _ := ammstore.New(storageDB)
	sc, _ := commitment.NewCommitStore(map[string]commitment.Tree{}, dbm.NewMemDB(), nil, noopLog)
	return &MockStore{Storage: ss, Commiter: *sc}
}

func (s *MockStore) LatestVersion() (uint64, error) {
	v, _, err :=  s.Storage.StateLatest()
	return v, err
}

func (s *MockStore) StateLatest() (uint64, corestore.ReaderMap, error) {
	return s.Storage.StateLatest()
}
func (s *MockStore) StateCommit(changes []corestore.StateChanges) (corestore.Hash, error) {
	_, state, _ := s.Storage.StateLatest()
	writer := branch.DefaultNewWriterMap(state)
	err := writer.ApplyStateChanges(changes)
	return []byte{}, err
}

type Result struct{
	key      []byte
	value    []byte
	version  uint64
	proofOps []proof.CommitmentOp
}

var _ cmt.QueryResult = (*Result)(nil)

func (s Result) Key() []byte {
	return s.key
}

func (s Result) Value() []byte {
	return s.value
}

func (s Result) Version() uint64 {
	return s.version
}

func (s Result) Proof() *ics23.CommitmentProof {
	return nil
}

func (s Result) ProofType() string {
	return ""
}

func (s *MockStore) Query(storeKey string, version uint64, key []byte, prove bool) (cmt.QueryResult, error) {
	state, err := s.Storage.StateAt(version)
	reader, err := state.GetReader([]byte(storeKey))
	value, err := reader.Get(key)
	res := Result{
		key: key,
		value: value,
		version: version,
	}
	return res, err
}

func (s *MockStore) LastCommitID() (proof.CommitID, error) {
	v, _, err := s.Storage.StateLatest()
	return proof.CommitID{
		Version: v,
		Hash: []byte{},
	}, err
}




