package lockingkv

import (
	"io"
	"sort"
	"sync"

	"golang.org/x/exp/slices"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/tracekv"
	storetypes "cosmossdk.io/store/types"
)

var (
	_ storetypes.CacheKVStore        = &LockableKV{}
	_ storetypes.LockingCacheWrapper = &LockableKV{}
	_ storetypes.CacheKVStore        = &LockedKV{}
	_ storetypes.LockingStore        = &LockedKV{}
)

func NewStore(parent storetypes.KVStore) *LockableKV {
	return &LockableKV{
		parent: parent,
		locks:  sync.Map{},
	}
}

// LockableKV is a store that is able to provide locks. Each locking key that is used for a lock must represent a
// disjoint partition of store keys that are able to be mutated. For example, locking per account public key would
// provide a lock over all mutations related to that account.
type LockableKV struct {
	parent    storetypes.KVStore
	locks     sync.Map // map from string key to *sync.Mutex.
	mutations sync.Map // map from string key to []byte.
}

func (s *LockableKV) Write() {
	s.locks.Range(func(key, value any) bool {
		lock := value.(*sync.Mutex)
		// We should be able to acquire the lock and only would not be able to if for some reason a child
		// store was not unlocked.
		if !lock.TryLock() {
			panic("LockedKV is missing Unlock() invocation.")
		}

		// We specifically don't unlock here which prevents users from acquiring the locks again and
		// mutating the values allowing the Write() invocation only to happen once effectively.

		return true
	})

	values := make(map[string][]byte)
	s.mutations.Range(func(key, value any) bool {
		values[key.(string)] = value.([]byte)
		return true
	})

	// We need to make the mutations to the parent in a deterministic order to ensure a deterministic hash.
	for _, sortedKey := range getSortedKeys[sort.StringSlice](values) {
		value := values[sortedKey]

		if value == nil {
			s.parent.Delete([]byte(sortedKey))
		} else {
			s.parent.Set([]byte(sortedKey), value)
		}
	}
}

func (s *LockableKV) GetStoreType() storetypes.StoreType {
	return s.parent.GetStoreType()
}

// CacheWrap allows for branching the store. Care must be taken to ensure that synchronization outside of this
// store is performed to ensure that reads and writes are linearized.
func (s *LockableKV) CacheWrap() storetypes.CacheWrap {
	return cachekv.NewStore(s)
}

// CacheWrapWithTrace allows for branching the store with tracing. Care must be taken to ensure that synchronization
// outside of this store is performed to ensure that reads and writes are linearized.
func (s *LockableKV) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(s, w, tc))
}

// CacheWrapWithLocks returns a store that allows mutating a set of store keys that are covered by the
// set of lock keys. Each lock key should represent a disjoint partitioned space of store keys for which
// the caller is acquiring locks for.
func (s *LockableKV) CacheWrapWithLocks(lockKeys [][]byte) storetypes.CacheWrap {
	stringLockedKeys := make([]string, len(lockKeys))
	for i, key := range lockKeys {
		stringLockedKeys[i] = string(key)
	}
	// Ensure that we always operate in a deterministic ordering when acquiring locks to prevent deadlock.
	slices.Sort(stringLockedKeys)
	for _, stringKey := range stringLockedKeys {
		v, _ := s.locks.LoadOrStore(stringKey, &sync.Mutex{})
		lock := v.(*sync.Mutex)
		lock.Lock()
	}

	return &LockedKV{
		parent:    s,
		lockKeys:  stringLockedKeys,
		mutations: make(map[string][]byte),
	}
}

func (s *LockableKV) Get(key []byte) []byte {
	v, loaded := s.mutations.Load(string(key))
	if loaded {
		return v.([]byte)
	}

	return s.parent.Get(key)
}

func (s *LockableKV) Has(key []byte) bool {
	v, loaded := s.mutations.Load(string(key))
	if loaded {
		return v.([]byte) != nil
	}

	return s.parent.Has(key)
}

func (s *LockableKV) Set(key, value []byte) {
	s.mutations.Store(string(key), value)
}

func (s *LockableKV) Delete(key []byte) {
	s.Set(key, nil)
}

func (s *LockableKV) Iterator(start, end []byte) storetypes.Iterator {
	panic("This store does not support iteration.")
}

func (s *LockableKV) ReverseIterator(start, end []byte) storetypes.Iterator {
	panic("This store does not support iteration.")
}

func (s *LockableKV) writeMutations(mutations map[string][]byte) {
	// We don't need to sort here since the sync.Map stores keys and values in an arbitrary order.
	// LockableKV.Write is responsible for sorting all the keys to ensure a deterministic write order.
	for key, mutation := range mutations {
		s.mutations.Store(key, mutation)
	}
}

func (s *LockableKV) unlock(lockKeys []string) {
	for _, key := range lockKeys {
		v, ok := s.locks.Load(key)
		if !ok {
			panic("Key not found")
		}
		lock := v.(*sync.Mutex)

		lock.Unlock()
	}
}

// LockedKV is a store that only allows setting of keys that have been locked via CacheWrapWithLocks.
// All other keys are allowed to be read but the user must ensure that no one else is able to mutate those
// values without the appropriate synchronization occurring outside of this store.
//
// This store does not support iteration.
type LockedKV struct {
	parent *LockableKV

	lockKeys  []string
	mutations map[string][]byte
}

func (s *LockedKV) Write() {
	s.parent.writeMutations(s.mutations)
}

func (s *LockedKV) Unlock() {
	s.parent.unlock(s.lockKeys)
}

func (s *LockedKV) GetStoreType() storetypes.StoreType {
	return s.parent.GetStoreType()
}

func (s *LockedKV) CacheWrap() storetypes.CacheWrap {
	return cachekv.NewStore(s)
}

func (s *LockedKV) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(s, w, tc))
}

func (s *LockedKV) Get(key []byte) []byte {
	if key == nil {
		panic("nil key")
	}

	if value, ok := s.mutations[string(key)]; ok {
		return value
	}

	return s.parent.Get(key)
}

func (s *LockedKV) Has(key []byte) bool {
	if key == nil {
		panic("nil key")
	}

	if value, ok := s.mutations[string(key)]; ok {
		return value != nil
	}

	return s.parent.Has(key)
}

func (s *LockedKV) Set(key, value []byte) {
	if key == nil {
		panic("nil key")
	}

	s.mutations[string(key)] = value
}

func (s *LockedKV) Delete(key []byte) {
	s.Set(key, nil)
}

func (s *LockedKV) Iterator(start, end []byte) storetypes.Iterator {
	panic("This store does not support iteration.")
}

func (s *LockedKV) ReverseIterator(start, end []byte) storetypes.Iterator {
	panic("This store does not support iteration.")
}

// getSortedKeys returns the keys of the map in sorted order.
func getSortedKeys[R interface {
	~[]K
	sort.Interface
}, K comparable, V any](m map[K]V,
) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Sort(R(keys))
	return keys
}
