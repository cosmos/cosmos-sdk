package rootmulti

import (
	"io"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/pruning"
)

var _ store.UpgradeableRootStore = (*Store)(nil)

// Store defines a multi-tree implementation variant of a RootStore .It contains
// a single State Storage (SS) backend and a State Commitment (SC) backend per module,
// i.e. store key. This implementation is meant to be congruent with the store
// v1 RootMultiStore and support existing application that DO NOT wish to migrate
// to the SDK's default single tree RootStore variant.
type Store struct {
	logger         log.Logger
	initialVersion uint64

	// stateStore reflects the state storage backend
	stateStore store.VersionedDatabase

	// commitHeader reflects the header used when committing state (note, this isn't required and only used for query purposes)
	commitHeader store.CommitHeader

	// lastCommitInfo reflects the last version/hash that has been committed
	lastCommitInfo *store.CommitInfo
	// workingHash defines the current (yet to be committed) hash
	workingHash []byte

	// traceWriter defines a writer for store tracing operation
	traceWriter io.Writer
	// traceContext defines the tracing context, if any, for trace operations
	traceContext store.TraceContext

	// pruningManager manages pruning of the SS and SC backends
	pruningManager *pruning.Manager
}

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
