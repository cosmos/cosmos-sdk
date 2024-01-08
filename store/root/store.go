package root

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/cockroachdb/errors"

	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/kv/branch"
	"cosmossdk.io/store/v2/kv/trace"
	"cosmossdk.io/store/v2/metrics"
	"cosmossdk.io/store/v2/pruning"
)

var _ store.RootStore = (*Store)(nil)

// Store defines the SDK's default RootStore implementation. It contains a single
// State Storage (SS) backend and a single State Commitment (SC) backend. Note,
// this means all store keys are ignored and commitments exist in a single commitment
// tree.
type Store struct {
	logger         log.Logger
	initialVersion uint64

	// stateStore reflects the state storage backend
	stateStore store.VersionedDatabase

	// stateCommitment reflects the state commitment (SC) backend
	stateCommitment store.Committer

	// kvStores reflects a mapping of store keys, typically dedicated to modules,
	// to a dedicated BranchedKVStore. Each store is used to accumulate writes
	// and branch off of.
	kvStores map[string]store.BranchedKVStore

	// commitHeader reflects the header used when committing state (note, this isn't required and only used for query purposes)
	commitHeader *coreheader.Info

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

	// telemetry reflects a telemetry agent responsible for emitting metrics (if any)
	telemetry metrics.StoreMetrics
}

func New(
	logger log.Logger,
	ss store.VersionedDatabase,
	sc store.Committer,
	storeKeys []string,
	ssOpts, scOpts pruning.Options,
	m metrics.StoreMetrics,
) (store.RootStore, error) {
	kvStores := make(map[string]store.BranchedKVStore, len(storeKeys))
	for _, storeKey := range storeKeys {
		bkv, err := branch.New(storeKey, ss)
		if err != nil {
			return nil, err
		}

		kvStores[storeKey] = bkv
	}

	pruningManager := pruning.NewManager(logger, ss, sc)
	pruningManager.SetStorageOptions(ssOpts)
	pruningManager.SetCommitmentOptions(scOpts)
	pruningManager.Start()

	return &Store{
		logger:          logger.With("module", "root_store"),
		initialVersion:  1,
		stateStore:      ss,
		stateCommitment: sc,
		kvStores:        kvStores,
		pruningManager:  pruningManager,
		telemetry:       m,
	}, nil
}

// Close closes the store and resets all internal fields. Note, Close() is NOT
// idempotent and should only be called once.
func (s *Store) Close() (err error) {
	err = errors.Join(err, s.stateStore.Close())
	err = errors.Join(err, s.stateCommitment.Close())

	s.stateStore = nil
	s.stateCommitment = nil
	s.lastCommitInfo = nil
	s.commitHeader = nil

	s.pruningManager.Stop()

	return err
}

func (s *Store) SetMetrics(m metrics.Metrics) {
	s.telemetry = m
}

func (s *Store) SetInitialVersion(v uint64) error {
	s.initialVersion = v

	return s.stateCommitment.SetInitialVersion(v)
}

// GetSCStore returns the store's state commitment (SC) backend.
func (s *Store) GetSCStore() store.Committer {
	return s.stateCommitment
}

// LastCommitID returns a CommitID based off of the latest internal CommitInfo.
// If an internal CommitInfo is not set, a new one will be returned with only the
// latest version set, which is based off of the SS view.
func (s *Store) LastCommitID() (store.CommitID, error) {
	if s.lastCommitInfo != nil {
		return s.lastCommitInfo.CommitID(), nil
	}

	// XXX/TODO: We cannot use SS to get the latest version when lastCommitInfo
	// is nil if SS is flushed asynchronously. This is because the latest version
	// in SS might not be the latest version in the SC stores.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314
	latestVersion, err := s.stateStore.GetLatestVersion()
	if err != nil {
		return store.CommitID{}, err
	}

	// sanity check: ensure integrity of latest version against SC
	scVersion, err := s.stateCommitment.GetLatestVersion()
	if err != nil {
		return store.CommitID{}, err
	}

	if scVersion != latestVersion {
		return store.CommitID{}, fmt.Errorf("SC and SS version mismatch; got: %d, expected: %d", scVersion, latestVersion)
	}

	return store.CommitID{Version: latestVersion}, nil
}

// GetLatestVersion returns the latest version based on the latest internal
// CommitInfo. An error is returned if the latest CommitInfo or version cannot
// be retrieved.
func (s *Store) GetLatestVersion() (uint64, error) {
	lastCommitID, err := s.LastCommitID()
	if err != nil {
		return 0, err
	}

	return lastCommitID.Version, nil
}

func (s *Store) Query(storeKey string, version uint64, key []byte, prove bool) (store.QueryResult, error) {
	if s.telemetry != nil {
		now := time.Now()
		s.telemetry.MeasureSince(now, "root_store", "query")
	}

	val, err := s.stateStore.Get(storeKey, version, key)
	if err != nil {
		return store.QueryResult{}, err
	}

	result := store.QueryResult{
		Key:     key,
		Value:   val,
		Version: version,
	}

	if prove {
		proof, err := s.stateCommitment.GetProof(storeKey, version, key)
		if err != nil {
			return store.QueryResult{}, err
		}

		result.Proof = store.NewIAVLCommitmentOp(key, proof)
	}

	return result, nil
}

// GetKVStore returns a KVStore for the given store key. Any writes to this store
// without branching will be committed to SC and SS upon Commit(). Branching will create
// a branched KVStore that allow writes to be discarded and propagated to the
// root KVStore using Write().
func (s *Store) GetKVStore(storeKey string) store.KVStore {
	bkv, ok := s.kvStores[storeKey]
	if !ok {
		panic(fmt.Sprintf("unknown store key: %s", storeKey))
	}

	if s.TracingEnabled() {
		return trace.New(bkv, s.traceWriter, s.traceContext)
	}

	return bkv
}

func (s *Store) GetBranchedKVStore(storeKey string) store.BranchedKVStore {
	// Branching will soon be removed.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/18981
	panic("TODO: WILL BE REMOVED!")
}

func (s *Store) LoadLatestVersion() error {
	if s.telemetry != nil {
		now := time.Now()
		s.telemetry.MeasureSince(now, "root_store", "load_latest_version")
	}

	lv, err := s.GetLatestVersion()
	if err != nil {
		return err
	}

	return s.loadVersion(lv)
}

func (s *Store) LoadVersion(version uint64) error {
	if s.telemetry != nil {
		now := time.Now()
		s.telemetry.MeasureSince(now, "root_store", "load_version")
	}

	return s.loadVersion(version)
}

func (s *Store) loadVersion(v uint64) error {
	s.logger.Debug("loading version", "version", v)

	// Reset each KVStore s.t. the latest version is v. Any writes will overwrite
	// existing versions.
	for storeKey, kvStore := range s.kvStores {
		if err := kvStore.Reset(v); err != nil {
			return fmt.Errorf("failed to reset %s KVStore: %w", storeKey, err)
		}
	}

	if err := s.stateCommitment.LoadVersion(v); err != nil {
		return fmt.Errorf("failed to load SS version %d: %w", v, err)
	}

	s.workingHash = nil
	s.commitHeader = nil

	// set lastCommitInfo explicitly s.t. Commit commits the correct version, i.e. v+1
	s.lastCommitInfo = &store.CommitInfo{Version: v}

	return nil
}

func (s *Store) SetTracingContext(tc store.TraceContext) {
	s.traceContext = tc
}

func (s *Store) SetTracer(w io.Writer) {
	s.traceWriter = w
}

func (s *Store) TracingEnabled() bool {
	return s.traceWriter != nil
}

func (s *Store) SetCommitHeader(h *coreheader.Info) {
	s.commitHeader = h
}

func (s *Store) Branch() store.BranchedRootStore {
	// Branching will soon be removed.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/18981
	panic("TODO: WILL BE REMOVED!")
}

// WorkingHash returns the working hash of the root store. Note, WorkingHash()
// should only be called once per block once all writes are complete and prior
// to Commit() being called.
//
// If working hash is nil, then we need to compute and set it on the root store
// by constructing a CommitInfo object, which in turn creates and writes a batch
// of the current changeset to the SC tree.
func (s *Store) WorkingHash() ([]byte, error) {
	if s.telemetry != nil {
		now := time.Now()
		s.telemetry.MeasureSince(now, "root_store", "working_hash")
	}

	if s.workingHash == nil {
		if err := s.writeSC(); err != nil {
			return nil, err
		}

		s.workingHash = s.lastCommitInfo.Hash()
	}

	return slices.Clone(s.workingHash), nil
}

func (s *Store) Write() {
	for _, kvStore := range s.kvStores {
		kvStore.Write()
	}
}

// Commit commits all state changes to the underlying SS and SC backends. Note,
// at the time of Commit(), we expect WorkingHash() to have already been called,
// which internally sets the working hash, retrieved by writing a batch of the
// changeset to the SC tree, and CommitInfo on the root store. The changeset is
// retrieved from the rootKVStore and represents the entire set of writes to be
// committed. The same changeset is used to flush writes to the SS backend.
//
// Note, Commit() commits SC and SC synchronously.
func (s *Store) Commit() ([]byte, error) {
	if s.telemetry != nil {
		now := time.Now()
		s.telemetry.MeasureSince(now, "root_store", "commit")
	}

	if s.workingHash == nil {
		return nil, fmt.Errorf("working hash is nil; must call WorkingHash() before Commit()")
	}

	version := s.lastCommitInfo.Version

	if s.commitHeader != nil && uint64(s.commitHeader.Height) != version {
		s.logger.Debug("commit header and version mismatch", "header_height", s.commitHeader.Height, "version", version)
	}

	changeset := store.NewChangeset()
	for _, kvStore := range s.kvStores {
		changeset.Merge(kvStore.GetChangeset())
	}

	// commit SS
	if err := s.stateStore.ApplyChangeset(version, changeset); err != nil {
		return nil, fmt.Errorf("failed to commit SS: %w", err)
	}

	// commit SC
	if err := s.commitSC(); err != nil {
		return nil, fmt.Errorf("failed to commit SC stores: %w", err)
	}

	if s.commitHeader != nil {
		s.lastCommitInfo.Timestamp = s.commitHeader.Time
	}

	for storeKey, kvStore := range s.kvStores {
		if err := kvStore.Reset(version); err != nil {
			return nil, fmt.Errorf("failed to reset %s KVStore: %w", storeKey, err)
		}
	}

	s.workingHash = nil

	// prune SS and SC
	s.pruningManager.Prune(version)

	return s.lastCommitInfo.Hash(), nil
}

// writeSC gets the current changeset from the rootKVStore and writes that as a
// batch to the underlying SC tree, which allows us to retrieve the working hash
// of the SC tree. Finally, we construct a *CommitInfo and return the hash.
// Note, this should only be called once per block!
func (s *Store) writeSC() error {
	changeset := store.NewChangeset()
	for _, kvStore := range s.kvStores {
		changeset.Merge(kvStore.GetChangeset())
	}

	if err := s.stateCommitment.WriteBatch(changeset); err != nil {
		return fmt.Errorf("failed to write batch to SC store: %w", err)
	}

	var previousHeight, version uint64
	if s.lastCommitInfo.GetVersion() == 0 && s.initialVersion > 1 {
		// This case means that no commit has been made in the store, we
		// start from initialVersion.
		version = s.initialVersion
	} else {
		// This case can means two things:
		//
		// 1. There was already a previous commit in the store, in which case we
		// 		increment the version from there.
		// 2. There was no previous commit, and initial version was not set, in which
		// 		case we start at version 1.
		previousHeight = s.lastCommitInfo.GetVersion()
		version = previousHeight + 1
	}

	s.lastCommitInfo = &store.CommitInfo{
		Version:    version,
		StoreInfos: s.stateCommitment.WorkingStoreInfos(version),
	}

	return nil
}

// commitSC commits the SC store. At this point, a batch of the current changeset
// should have already been written to the SC via WorkingHash(). This method
// solely commits that batch. An error is returned if commit fails or if the
// resulting commit hash is not equivalent to the working hash.
func (s *Store) commitSC() error {
	commitStoreInfos, err := s.stateCommitment.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit SC store: %w", err)
	}

	commitHash := (&store.CommitInfo{
		Version:    s.lastCommitInfo.Version,
		StoreInfos: commitStoreInfos,
	}).Hash()

	workingHash, err := s.WorkingHash()
	if err != nil {
		return fmt.Errorf("failed to get working hash: %w", err)
	}

	if !bytes.Equal(commitHash, workingHash) {
		return fmt.Errorf("unexpected commit hash; got: %X, expected: %X", commitHash, workingHash)
	}

	return nil
}
