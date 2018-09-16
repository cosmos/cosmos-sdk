package types

import (
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/transient"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// nolint: reexport
type (
	CommitStore      = types.CommitStore
	KVStore          = types.KVStore
	CacheKVStore     = types.CacheKVStore
	MultiStore       = types.MultiStore
	CommitMultiStore = types.CommitMultiStore
	CommitID         = types.CommitID
	Iterator         = types.Iterator

	StoreKey          = types.StoreKey
	KVStoreKey        = iavl.KVStoreKey
	TransientStoreKey = transient.TransientStoreKey

	Gas       = types.Gas
	GasTank   = types.GasTank
	GasConfig = types.GasConfig
	GasMeter  = types.GasMeter

	ErrorOutOfGas = types.ErrorOutOfGas
)

// nolint: reexport
func KVStorePrefixIterator(store KVStore, prefix []byte) Iterator {
	return types.KVStorePrefixIterator(store, prefix)
}
func KVStoreReversePrefixIterator(store KVStore, prefix []byte) Iterator {
	return types.KVStoreReversePrefixIterator(store, prefix)
}
func PrefixEndBytes(prefix []byte) []byte {
	return types.PrefixEndBytes(prefix)
}
func NewKVStoreKey(name string) *KVStoreKey {
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
