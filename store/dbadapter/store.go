package dbadapter

import (
	"io"

	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/cache"
)

// Wrapper type for dbm.Db with implementation of KVStore
type dbStoreAdapter struct {
	dbm.DB
}

// Implements Store.
func (dbStoreAdapter) GetStoreType() types.StoreType {
	return types.StoreTypeDB
}

// Implements KVStore.
func (dsa dbStoreAdapter) CacheWrap() types.CacheWrap {
	return cache.NewStore(dsa)
}

// CacheWrapWithTrace implements the KVStore interface.
func (dsa dbStoreAdapter) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return NewCacheKVStore(NewTraceKVStore(dsa, w, tc))
}

// XXX: delete
/*
// Implements KVStore
func (dsa dbStoreAdapter) Prefix(prefix []byte) KVStore {
	return prefixStore{dsa, prefix}
}

// Implements KVStore
func (dsa dbStoreAdapter) Gas(meter GasMeter, config GasConfig) KVStore {
	return NewGasKVStore(meter, config, dsa)
}
*/
// dbm.DB implements KVStore so we can CacheKVStore it.
var _ types.KVStore = dbStoreAdapter{}

//----------------------------------------
// commitDBStoreWrapper should only be used for simulation/debugging,
// as it doesn't compute any commit hash, and it cannot load older state.

// Wrapper type for dbm.Db with implementation of KVStore
type commitDBStoreAdapter struct {
	dbStoreAdapter
}

func (cdsa commitDBStoreAdapter) Commit() types.CommitID {
	return types.CommitID{
		Version: -1,
		Hash:    []byte("FAKE_HASH"),
	}
}

func (cdsa commitDBStoreAdapter) LastCommitID() types.CommitID {
	return types.CommitID{
		Version: -1,
		Hash:    []byte("FAKE_HASH"),
	}
}

func (cdsa commitDBStoreAdapter) SetPruning(_ types.PruningStrategy) {}
