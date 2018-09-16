package types

import (
	"fmt"

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

// Stores of MultiStore must implement CommitStore.
type CommitStore interface {
	// CONTRACT: return zero CommitID to skip writing
	Commit() CommitID
	LastCommitID() CommitID
	SetPruning(PruningStrategy)
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
	// Convenience for fetching substores.
	GetKVStore(StoreKey) KVStore

	// TODO: recursive multistore not yet supported
	// GetMultiStore(StoreKey) MultiStore

	GetTracer() *Tracer

	GetGasTank() *GasTank

	// CacheWrap cache wraps
	// Having this method here because there is currently no
	// implementation of MultiStore that panics on CacheWrap().
	// Move this method to CacheWrapperMultiStore when needed
	CacheWrap() CacheMultiStore
}

// From MultiStore.CacheMultiStore()....
type CacheMultiStore interface {
	MultiStore

	Write() // Writes operations to underlying KVStore
}

// A non-cache MultiStore.
type CommitMultiStore interface {
	CommitStore
	MultiStore

	// Mount a store of type using the given db.
	// If db == nil, the new store will use the CommitMultiStore db.
	MountStoreWithDB(key StoreKey, db dbm.DB)

	// Panics on a nil key.
	GetCommitStore(key StoreKey) CommitStore

	// Panics on a nil key.
	GetCommitKVStore(key StoreKey) CommitKVStore

	// Panics on a nil key.
	// TODO: recursive multistore not yet supported
	// GetCommitMultiStore(key StoreKey) CommitMultiStore

	// Load the latest persisted version.  Called once after all
	// calls to Mount*Store() are complete.
	LoadLatestVersion() error

	// Load a specific persisted version.  When you load an old
	// version, or when the last commit attempt didn't complete,
	// the next commit after loading must be idempotent (return the
	// same commit id).  Otherwise the behavior is undefined.
	LoadMultiStoreVersion(ver int64) error
}

//---------subsp-------------------------------
// KVStore

// KVStore is a simple interface to get/set data
type KVStore interface {
	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key.
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
	// Start must be greater than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	ReverseIterator(start, end []byte) Iterator

	// TODO Not yet implemented.
	// CreateSubKVStore(key *storeKey) (KVStore, error)

	// TODO Not yet implemented.
	// GetSubKVStore(key *storeKey) KVStore

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

type CacheWrapperKVStore interface {
	KVStore

	// CacheWrap cache wraps
	CacheWrap() CacheKVStore
}

// CacheKVStore cache-wraps a KVStore.  After calling .Write() on
// the CacheKVStore, all previously created CacheKVStores on the
// object expire.
type CacheKVStore interface {
	CacheWrapperKVStore

	// Writes operations to underlying KVStore
	Write()
}

// Stores of MultiStore must implement CommitStore.
type CommitKVStore interface {
	CommitStore
	KVStore

	// Load a specific persisted version.  When you load an old
	// version, or when the last commit attempt didn't complete,
	// the next commit after loading must be idempotent (return the
	// same commit id).  Otherwise the behavior is undefined.
	LoadKVStoreVersion(db dbm.DB, id CommitID) error
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
// Keys for accessing substores

// StoreKey is a key used to index stores in a MultiStore.
type StoreKey interface {
	Name() string
	String() string
	NewStore() CommitStore
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

//----------------------------------------

// key-value result for iterator queries
type KVPair cmn.KVPair
