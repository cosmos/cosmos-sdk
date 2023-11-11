package rootmulti

import (
	"io"

	"cosmossdk.io/store/v2"
)

var _ store.UpgradeableRootStore = (*Store)(nil)

// Store defines a multi-tree implementation variant of a RootStore .It contains
// a single State Storage (SS) backend and a State Commitment (SC) backend per module,
// i.e. store key. This implementation is meant to be congruent with the store
// v1 RootMultiStore and support existing application that DO NOT wish to migrate
// to the SDK's default single tree RootStore variant.
type Store struct{}

func (s *Store) GetSCStore(storeKey string) store.Committer {
	panic("not implemented!")
}

func (s *Store) MountSCStore(storeKey string, sc store.Committer) error {
	panic("not implemented!")
}

func (s *Store) GetKVStore(storeKey string) store.KVStore {
	panic("not implemented!")
}

func (s *Store) GetBranchedKVStore(storeKey string) store.BranchedKVStore {
	panic("not implemented!")
}

func (s *Store) Query(storeKey string, version uint64, key []byte, prove bool) (store.QueryResult, error) {
	panic("not implemented!")
}

func (s *Store) Branch() store.BranchedRootStore {
	panic("not implemented!")
}

func (s *Store) SetTracingContext(tc store.TraceContext) {
	panic("not implemented!")
}

func (s *Store) SetTracer(w io.Writer) {
	panic("not implemented!")
}

func (s *Store) TracingEnabled() bool {
	panic("not implemented!")
}

func (s *Store) LoadVersion(version uint64) error {
	panic("not implemented!")
}

func (s *Store) LoadLatestVersion() error {
	panic("not implemented!")
}

func (s *Store) LoadVersionAndUpgrade(version uint64, upgrades *store.StoreUpgrades) error {
	panic("not implemented!")
}

func (s *Store) GetLatestVersion() (uint64, error) {
	panic("not implemented!")
}

func (s *Store) SetCommitHeader(h store.CommitHeader) {
	panic("not implemented!")
}

func (s *Store) WorkingHash() ([]byte, error) {
	panic("not implemented!")
}

func (s *Store) Commit() ([]byte, error) {
	panic("not implemented!")
}

func (s *Store) Close() error {
	panic("not implemented!")
}
