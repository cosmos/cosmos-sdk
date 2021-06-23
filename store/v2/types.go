package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	v1 "github.com/cosmos/cosmos-sdk/store/types"
)

type Store interface {
	GetStoreType() StoreType
	// CacheWrapper
}

// something that can persist to disk
type Committer = v1.Committer

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

type PruningOptions = v1.PruningOptions
type TraceContext = v1.TraceContext
type StoreKey = v1.StoreKey
type CommitID = v1.CommitID
type WriteListener = v1.WriteListener

// type KVStorePrefixIterator = v1.KVStorePrefixIterator

//---------subsp-------------------------------
// KVStore

// BasicKVStore is a simple interface to get/set data
type BasicKVStore interface {
	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key or value.
	Set(key, value []byte)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)
}

// KVStore additionally provides iteration and deletion
type KVStore interface {
	Store
	BasicKVStore

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	Iterator(start, end []byte) Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	ReverseIterator(start, end []byte) Iterator
}

// Iterator is an alias db's Iterator for convenience.
type Iterator = dbm.Iterator

// CacheKVStore branches a KVStore and provides read cache functionality.
// After calling .Write() on the CacheKVStore, all previously created
// CacheKVStores on the object expire.
type CacheKVStore interface {
	KVStore

	// Writes operations to underlying KVStore
	Write()
}

// CommitKVStore is an interface for MultiStore.
type CommitKVStore interface {
	Committer
	KVStore
}

type VersionedKVStore interface {
	CommitKVStore
	AtVersion(int64) (KVStore, error)
	VersionExists(int64) bool
}

//----------------------------------------
// CacheWrap

// CacheWrap is the most appropriate interface for store ephemeral branching and cache.
// For example, IAVLStore.CacheWrap() returns a CacheKVStore. CacheWrap should not return
// a Committer, since Commit ephemeral store make no sense. It can return KVStore,
// HeapStore, SpaceStore, etc.
type CacheWrap interface {
	CacheWrapper
	// Write syncs with the underlying store.
	Write()
}

type CacheWrapper interface {
	// CacheWrap branches a store.
	CacheWrap() CacheWrap
}

//----------------------------------------
// Store types

// kind of store
type StoreType int

const (
	StoreTypeMulti StoreType = iota
	StoreTypeDB
	StoreTypeIAVL
	StoreTypeDecoupled
	StoreTypeTransient
	StoreTypeMemory
	StoreTypeSMT
)

func (st StoreType) String() string {
	switch st {
	case StoreTypeMulti:
		return "StoreTypeMulti"

	case StoreTypeDB:
		return "StoreTypeDB"

	case StoreTypeIAVL:
		return "StoreTypeIAVL"

	case StoreTypeDecoupled:
		return "StoreTypeDecoupled"

	case StoreTypeTransient:
		return "StoreTypeTransient"

	case StoreTypeMemory:
		return "StoreTypeMemory"

	case StoreTypeSMT:
		return "StoreTypeSMT"
	}

	return "unknown store type"
}
