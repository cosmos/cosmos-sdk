package iavlv2

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/telemetry"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/snapshots"
)

var (
	_           store.VersionedWriter        = (*Store)(nil)
	_           snapshots.StorageSnapshotter = (*Store)(nil)
	_           store.Pruner                 = (*Store)(nil)
	_           store.UpgradableDatabase     = (*Store)(nil)
	once                                     = sync.Once{}
	metricOpKey                              = []string{"store", "iavlv2", "op"}
	getLabel                                 = telemetry.Label{Name: "method", Value: "get"}
	hasLabel                                 = telemetry.Label{Name: "method", Value: "has"}
)

type Store struct {
	telemetry telemetry.Service
	trees     map[string]Reader
}

type Reader interface {
	Has(version uint64, key []byte) (bool, error)
	Get(version uint64, key []byte) ([]byte, error)
	Iterator(version uint64, start, end []byte, ascending bool) (corestore.Iterator, error)
	Version() uint64
}

func NewStore(trees map[string]Reader, ts telemetry.Service) *Store {
	return &Store{
		telemetry: ts,
		trees:     trees,
	}
}

func unsafeString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }

func (s *Store) PruneStoreKeys(storeKeys []string, version uint64) error {
	return nil
}

func (s *Store) Has(storeKey []byte, version uint64, key []byte) (bool, error) {
	defer s.telemetry.MeasureSince(time.Now(), metricOpKey, hasLabel)
	if tree, ok := s.trees[unsafeString(storeKey)]; !ok {
		return false, fmt.Errorf("store key %s not found", storeKey)
	} else {
		return tree.Has(version, key)
	}
}

func (s *Store) Get(storeKey []byte, version uint64, key []byte) ([]byte, error) {
	defer s.telemetry.MeasureSince(time.Now(), metricOpKey, getLabel)
	if tree, ok := s.trees[unsafeString(storeKey)]; !ok {
		return nil, fmt.Errorf("store key %s not found", storeKey)
	} else {
		return tree.Get(version, key)
	}
}

func (s *Store) GetLatestVersion() (uint64, error) {
	// FIXME: hack
	for _, tree := range s.trees {
		return tree.Version(), nil
	}
	return 0, errors.New("no trees mounted")
}

func (s *Store) VersionExists(v uint64) (bool, error) {
	// FIXME: hack
	return true, nil
}

func (s *Store) Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	if tree, ok := s.trees[unsafeString(storeKey)]; !ok {
		return nil, fmt.Errorf("store key %s not found", storeKey)
	} else {
		if version != tree.Version() {
			return nil, fmt.Errorf("loading past version not yet supported")
		}
		return tree.Iterator(version, start, end, true)
	}
}

func (s *Store) ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	if tree, ok := s.trees[unsafeString(storeKey)]; !ok {
		return nil, fmt.Errorf("store key %s not found", storeKey)
	} else {
		if version != tree.Version() {
			return nil, fmt.Errorf("loading past version not yet supported")
		}
		return tree.Iterator(version, start, end, false)
	}
}

func (s *Store) SetLatestVersion(version uint64) error {
	// nothing to do
	return nil
}

func (s *Store) ApplyChangeset(cs *corestore.Changeset) error {
	// nothing to do
	return nil
}

func (s *Store) Close() error {
	// nothing to do
	return nil
}

func (s *Store) Restore(version uint64, chStorage <-chan *corestore.StateChanges) error {
	// nothing to do
	return nil
}

func (s *Store) Prune(version uint64) error {
	// nothing to do
	return nil
}

func (s *Store) NewBatch(version uint64) (store.Batch, error) {
	// nothing to do
	return &nopBatch{}, nil
}

var _ store.Batch = (*nopBatch)(nil)

type nopBatch struct{}

func (b *nopBatch) Write() error { return nil }

func (b *nopBatch) Set(storeKey, key, value []byte) error { return nil }

func (b *nopBatch) Delete(storeKey, key []byte) error { return nil }

func (b *nopBatch) Size() int { return 0 }

func (b *nopBatch) Reset() error { return nil }
