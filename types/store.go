package types

import (
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/transient"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// nolint - reexport
type (
	PruningStrategy  = types.PruningStrategy
	CommitStore      = types.CommitStore
	Queryable        = types.Queryable
	KVStore          = types.KVStore
	CacheKVStore     = types.CacheKVStore
	CommitKVStore    = types.CommitKVStore
	MultiStore       = types.MultiStore
	CacheMultiStore  = types.CacheMultiStore
	CommitMultiStore = types.CommitMultiStore
	CommitID         = types.CommitID
	Iterator         = types.Iterator

	KVPair = types.KVPair

	KVStoreKey        = types.KVStoreKey
	IAVLStoreKey      = iavl.StoreKey
	TransientStoreKey = transient.StoreKey

	Gas       = types.Gas
	GasTank   = types.GasTank
	GasConfig = types.GasConfig
	GasMeter  = types.GasMeter

	Tracer       = types.Tracer
	TraceContext = types.TraceContext

	ErrorOutOfGas = types.ErrorOutOfGas
)

// nolint - reexport
const (
	PruneNothing    = types.PruneNothing
	PruneEverything = types.PruneEverything
	PruneSyncable   = types.PruneSyncable
)

// nolint - reexport
func KVStorePrefixIterator(store KVStore, prefix []byte) Iterator {
	return types.KVStorePrefixIterator(store, prefix)
}
func KVStoreReversePrefixIterator(store KVStore, prefix []byte) Iterator {
	return types.KVStoreReversePrefixIterator(store, prefix)
}
func PrefixEndBytes(prefix []byte) []byte {
	return types.PrefixEndBytes(prefix)
}
func InclusiveEndBytes(inclusiveBytes []byte) []byte {
	return types.InclusiveEndBytes(inclusiveBytes)
}
func NewKVStoreKey(name string) *IAVLStoreKey {
	return iavl.NewKey(name)
}
func NewTransientStoreKey(name string) *TransientStoreKey {
	return transient.NewKey(name)
}
func NewGasMeter(limit Gas) GasMeter {
	return types.NewGasMeter(limit)
}
func NewInfiniteGasMeter() GasMeter {
	return types.NewInfiniteGasMeter()
}
func NewGasTank(limit Gas, config GasConfig) *GasTank {
	return types.NewGasTank(limit, config)
}
func NewInfiniteGasTank(config GasConfig) *GasTank {
	return types.NewInfiniteGasTank(config)
}

// nolint: internal
var (
	cachedKVGasConfig        = types.KVGasConfig()
	cachedTransientGasConfig = types.TransientGasConfig()
)
