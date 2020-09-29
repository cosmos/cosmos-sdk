package types

import (
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

type (
	PruningOptions = types.PruningOptions
)

type (
	Store                     = types.Store
	Committer                 = types.Committer
	CommitStore               = types.CommitStore
	Queryable                 = types.Queryable
	MultiStore                = types.MultiStore
	CacheMultiStore           = types.CacheMultiStore
	CommitMultiStore          = types.CommitMultiStore
	MultiStorePersistentCache = types.MultiStorePersistentCache
	KVStore                   = types.KVStore
	Iterator                  = types.Iterator
)

// StoreDecoderRegistry defines each of the modules store decoders. Used for ImportExport
// simulation.
type StoreDecoderRegistry map[string]func(kvA, kvB kv.Pair) string

// Iterator over all the keys with a certain prefix in ascending order
func KVStorePrefixIterator(kvs KVStore, prefix []byte) Iterator {
	return types.KVStorePrefixIterator(kvs, prefix)
}

// Iterator over all the keys with a certain prefix in descending order.
func KVStoreReversePrefixIterator(kvs KVStore, prefix []byte) Iterator {
	return types.KVStoreReversePrefixIterator(kvs, prefix)
}

// KVStorePrefixIteratorPaginated returns iterator over items in the selected page.
// Items iterated and skipped in ascending order.
func KVStorePrefixIteratorPaginated(kvs KVStore, prefix []byte, page, limit uint) Iterator {
	return types.KVStorePrefixIteratorPaginated(kvs, prefix, page, limit)
}

// KVStoreReversePrefixIteratorPaginated returns iterator over items in the selected page.
// Items iterated and skipped in descending order.
func KVStoreReversePrefixIteratorPaginated(kvs KVStore, prefix []byte, page, limit uint) Iterator {
	return types.KVStoreReversePrefixIteratorPaginated(kvs, prefix, page, limit)
}

// DiffKVStores compares two KVstores and returns all the key/value pairs
// that differ from one another. It also skips value comparison for a set of provided prefixes
func DiffKVStores(a KVStore, b KVStore, prefixesToSkip [][]byte) (kvAs, kvBs []kv.Pair) {
	return types.DiffKVStores(a, b, prefixesToSkip)
}

type (
	CacheKVStore  = types.CacheKVStore
	CommitKVStore = types.CommitKVStore
	CacheWrap     = types.CacheWrap
	CacheWrapper  = types.CacheWrapper
	CommitID      = types.CommitID
)

type StoreType = types.StoreType

const (
	StoreTypeMulti     = types.StoreTypeMulti
	StoreTypeDB        = types.StoreTypeDB
	StoreTypeIAVL      = types.StoreTypeIAVL
	StoreTypeTransient = types.StoreTypeTransient
	StoreTypeMemory    = types.StoreTypeMemory
)

type (
	StoreKey          = types.StoreKey
	CapabilityKey     = types.CapabilityKey
	KVStoreKey        = types.KVStoreKey
	TransientStoreKey = types.TransientStoreKey
	MemoryStoreKey    = types.MemoryStoreKey
)

// NewKVStoreKey returns a new pointer to a KVStoreKey.
// Use a pointer so keys don't collide.
func NewKVStoreKey(name string) *KVStoreKey {
	return types.NewKVStoreKey(name)
}

// NewKVStoreKeys returns a map of new  pointers to KVStoreKey's.
// Uses pointers so keys don't collide.
func NewKVStoreKeys(names ...string) map[string]*KVStoreKey {
	keys := make(map[string]*KVStoreKey)
	for _, name := range names {
		keys[name] = NewKVStoreKey(name)
	}

	return keys
}

// Constructs new TransientStoreKey
// Must return a pointer according to the ocap principle
func NewTransientStoreKey(name string) *TransientStoreKey {
	return types.NewTransientStoreKey(name)
}

// NewTransientStoreKeys constructs a new map of TransientStoreKey's
// Must return pointers according to the ocap principle
func NewTransientStoreKeys(names ...string) map[string]*TransientStoreKey {
	keys := make(map[string]*TransientStoreKey)
	for _, name := range names {
		keys[name] = NewTransientStoreKey(name)
	}

	return keys
}

// NewMemoryStoreKeys constructs a new map matching store key names to their
// respective MemoryStoreKey references.
func NewMemoryStoreKeys(names ...string) map[string]*MemoryStoreKey {
	keys := make(map[string]*MemoryStoreKey)
	for _, name := range names {
		keys[name] = types.NewMemoryStoreKey(name)
	}

	return keys
}

// PrefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func PrefixEndBytes(prefix []byte) []byte {
	return types.PrefixEndBytes(prefix)
}

// InclusiveEndBytes returns the []byte that would end a
// range query such that the input would be included
func InclusiveEndBytes(inclusiveBytes []byte) (exclusiveBytes []byte) {
	return types.InclusiveEndBytes(inclusiveBytes)
}

//----------------------------------------

// key-value result for iterator queries
type KVPair = types.KVPair

//----------------------------------------

// TraceContext contains TraceKVStore context data. It will be written with
// every trace operation.
type TraceContext = types.TraceContext

// --------------------------------------

type (
	Gas       = types.Gas
	GasMeter  = types.GasMeter
	GasConfig = types.GasConfig
)

func NewGasMeter(limit Gas) GasMeter {
	return types.NewGasMeter(limit)
}

type (
	ErrorOutOfGas    = types.ErrorOutOfGas
	ErrorGasOverflow = types.ErrorGasOverflow
)

func NewInfiniteGasMeter() GasMeter {
	return types.NewInfiniteGasMeter()
}
