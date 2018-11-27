package types

import (
	"bytes"
	"fmt"
	"io"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// NOTE: These are implemented in cosmos-sdk/store.

// PruningStrategy specfies how old states will be deleted over time
type PruningStrategy uint8

const (
	// PruneSyncable means only those states not needed for state syncing will be deleted (keeps last 100 + every 10000th)
	PruneSyncable PruningStrategy = iota

	// PruneEverything means all saved states will be deleted, storing only the current state
	PruneEverything PruningStrategy = iota

	// PruneNothing means all historic states will be saved, nothing will be deleted
	PruneNothing PruningStrategy = iota
)

type Store interface { //nolint
	GetStoreType() StoreType
	CacheWrapper
}

// something that can persist to disk
type Committer interface {
	Commit() CommitID
	LastCommitID() CommitID
	SetPruning(PruningStrategy)
}

// Stores of MultiStore must implement CommitStore.
type CommitStore interface {
	Committer
	Store
}

// Queryable allows a Store to expose internal state to the abci.Query
// interface. Multistore can route requests to the proper Store.
//
// This is an optional, but useful extension to any CommitStore
type Queryable interface {
	Query(abci.RequestQuery) abci.ResponseQuery
}

//----------------------------------------
// MultiStore

type MultiStore interface { //nolint
	Store

	// Cache wrap MultiStore.
	// NOTE: Caller should probably not call .Write() on each, but
	// call CacheMultiStore.Write().
	CacheMultiStore() CacheMultiStore

	// Convenience for fetching substores.
	// If the store does not exist, panics.
	GetStore(StoreKey) Store
	GetKVStore(StoreKey) KVStore

	// TracingEnabled returns if tracing is enabled for the MultiStore.
	TracingEnabled() bool

	// WithTracer sets the tracer for the MultiStore that the underlying
	// stores will utilize to trace operations. A MultiStore is returned.
	WithTracer(w io.Writer) MultiStore

	// WithTracingContext sets the tracing context for a MultiStore. It is
	// implied that the caller should update the context when necessary between
	// tracing operations. A MultiStore is returned.
	WithTracingContext(TraceContext) MultiStore

	// ResetTraceContext resets the current tracing context.
	ResetTraceContext() MultiStore
}

// From MultiStore.CacheMultiStore()....
type CacheMultiStore interface {
	MultiStore
	Write() // Writes operations to underlying KVStore
}

// A non-cache MultiStore.
type CommitMultiStore interface {
	Committer
	MultiStore

	// Mount a store of type using the given db.
	// If db == nil, the new store will use the CommitMultiStore db.
	MountStoreWithDB(key StoreKey, typ StoreType, db dbm.DB)

	// Panics on a nil key.
	GetCommitStore(key StoreKey) CommitStore

	// Panics on a nil key.
	GetCommitKVStore(key StoreKey) CommitKVStore

	// Load the latest persisted version.  Called once after all
	// calls to Mount*Store() are complete.
	LoadLatestVersion() error

	// Load a specific persisted version.  When you load an old
	// version, or when the last commit attempt didn't complete,
	// the next commit after loading must be idempotent (return the
	// same commit id).  Otherwise the behavior is undefined.
	LoadVersion(ver int64) error
}

//---------subsp-------------------------------
// KVStore

// KVStore is a simple interface to get/set data
type KVStore interface {
	Store

	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key or value.
	Set(key, value []byte)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	Iterator(start, end []byte) Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	ReverseIterator(start, end []byte) Iterator

	// TODO Not yet implemented.
	// CreateSubKVStore(key *storeKey) (KVStore, error)

	// TODO Not yet implemented.
	// GetSubKVStore(key *storeKey) KVStore

	// Prefix applied keys with the argument
	// CONTRACT: when Prefix is called on a KVStore more than once,
	// the concatanation of the prefixes is applied
	Prefix(prefix []byte) KVStore

	// Gas consuming store
	// CONTRACT: when Gas is called on a KVStore more than once,
	// the concatanation of the meters/configs is applied
	Gas(GasMeter, GasConfig) KVStore
}

// Alias iterator to db's Iterator for convenience.
type Iterator = dbm.Iterator

// Iterator over all the keys with a certain prefix in ascending order
func KVStorePrefixIterator(kvs KVStore, prefix []byte) Iterator {
	return kvs.Iterator(prefix, PrefixEndBytes(prefix))
}

// Iterator over all the keys with a certain prefix in descending order.
func KVStoreReversePrefixIterator(kvs KVStore, prefix []byte) Iterator {
	return kvs.ReverseIterator(prefix, PrefixEndBytes(prefix))
}

// Compare two KVstores, return either the first key/value pair
// at which they differ and whether or not they are equal, skipping
// value comparison for a set of provided prefixes
func DiffKVStores(a KVStore, b KVStore, prefixesToSkip [][]byte) (kvA cmn.KVPair, kvB cmn.KVPair, count int64, equal bool) {
	iterA := a.Iterator(nil, nil)
	iterB := b.Iterator(nil, nil)
	count = int64(0)
	for {
		if !iterA.Valid() && !iterB.Valid() {
			break
		}
		var kvA, kvB cmn.KVPair
		if iterA.Valid() {
			kvA = cmn.KVPair{Key: iterA.Key(), Value: iterA.Value()}
			iterA.Next()
		}
		if iterB.Valid() {
			kvB = cmn.KVPair{Key: iterB.Key(), Value: iterB.Value()}
			iterB.Next()
		}
		if !bytes.Equal(kvA.Key, kvB.Key) {
			return kvA, kvB, count, false
		}
		compareValue := true
		for _, prefix := range prefixesToSkip {
			// Skip value comparison if we matched a prefix
			if bytes.Equal(kvA.Key[:len(prefix)], prefix) {
				compareValue = false
			}
		}
		if compareValue && !bytes.Equal(kvA.Value, kvB.Value) {
			return kvA, kvB, count, false
		}
		count++
	}
	return cmn.KVPair{}, cmn.KVPair{}, count, true
}

// CacheKVStore cache-wraps a KVStore.  After calling .Write() on
// the CacheKVStore, all previously created CacheKVStores on the
// object expire.
type CacheKVStore interface {
	KVStore

	// Writes operations to underlying KVStore
	Write()
}

// Stores of MultiStore must implement CommitStore.
type CommitKVStore interface {
	Committer
	KVStore
}

//----------------------------------------
// CacheWrap

// CacheWrap makes the most appropriate cache-wrap. For example,
// IAVLStore.CacheWrap() returns a CacheKVStore. CacheWrap should not return
// a Committer, since Commit cache-wraps make no sense. It can return KVStore,
// HeapStore, SpaceStore, etc.
type CacheWrap interface {
	// Write syncs with the underlying store.
	Write()

	// CacheWrap recursively wraps again.
	CacheWrap() CacheWrap

	// CacheWrapWithTrace recursively wraps again with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap
}

type CacheWrapper interface { //nolint
	// CacheWrap cache wraps.
	CacheWrap() CacheWrap

	// CacheWrapWithTrace cache wraps with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap
}

//----------------------------------------
// CommitID

// CommitID contains the tree version number and its merkle root.
type CommitID struct {
	Version int64
	Hash    []byte
}

func (cid CommitID) IsZero() bool { //nolint
	return cid.Version == 0 && len(cid.Hash) == 0
}

func (cid CommitID) String() string {
	return fmt.Sprintf("CommitID{%v:%X}", cid.Hash, cid.Version)
}

//----------------------------------------
// Store types

// kind of store
type StoreType int

const (
	//nolint
	StoreTypeMulti StoreType = iota
	StoreTypeDB
	StoreTypeIAVL
	StoreTypeTransient
)

//----------------------------------------
// Keys for accessing substores

// StoreKey is a key used to index stores in a MultiStore.
type StoreKey interface {
	Name() string
	String() string
}

// KVStoreKey is used for accessing substores.
// Only the pointer value should ever be used - it functions as a capabilities key.
type KVStoreKey struct {
	name string
}

// NewKVStoreKey returns a new pointer to a KVStoreKey.
// Use a pointer so keys don't collide.
func NewKVStoreKey(name string) *KVStoreKey {
	return &KVStoreKey{
		name: name,
	}
}

func (key *KVStoreKey) Name() string {
	return key.name
}

func (key *KVStoreKey) String() string {
	return fmt.Sprintf("KVStoreKey{%p, %s}", key, key.name)
}

// PrefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func PrefixEndBytes(prefix []byte) []byte {
	if prefix == nil {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		} else {
			end = end[:len(end)-1]
			if len(end) == 0 {
				end = nil
				break
			}
		}
	}
	return end
}

// InclusiveEndBytes returns the []byte that would end a
// range query such that the input would be included
func InclusiveEndBytes(inclusiveBytes []byte) (exclusiveBytes []byte) {
	exclusiveBytes = append(inclusiveBytes, byte(0x00))
	return exclusiveBytes
}

// TransientStoreKey is used for indexing transient stores in a MultiStore
type TransientStoreKey struct {
	name string
}

// Constructs new TransientStoreKey
// Must return a pointer according to the ocap principle
func NewTransientStoreKey(name string) *TransientStoreKey {
	return &TransientStoreKey{
		name: name,
	}
}

// Implements StoreKey
func (key *TransientStoreKey) Name() string {
	return key.name
}

// Implements StoreKey
func (key *TransientStoreKey) String() string {
	return fmt.Sprintf("TransientStoreKey{%p, %s}", key, key.name)
}

//----------------------------------------

// key-value result for iterator queries
type KVPair cmn.KVPair

//----------------------------------------

// TraceContext contains TraceKVStore context data. It will be written with
// every trace operation.
type TraceContext map[string]interface{}
