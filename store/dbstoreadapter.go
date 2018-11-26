package store

import (
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// Wrapper type for dbm.Db with implementation of KVStore
type dbStoreAdapter struct {
	dbm.DB
}

// Implements Store.
func (dbStoreAdapter) GetStoreType() StoreType {
	return sdk.StoreTypeDB
}

// Implements KVStore.
func (dsa dbStoreAdapter) CacheWrap() CacheWrap {
	return NewCacheKVStore(dsa)
}

// CacheWrapWithTrace implements the KVStore interface.
func (dsa dbStoreAdapter) CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap {
	return NewCacheKVStore(NewTraceKVStore(dsa, w, tc))
}

// Implements KVStore
func (dsa dbStoreAdapter) Prefix(prefix []byte) KVStore {
	return prefixStore{dsa, prefix}
}

// Implements KVStore
func (dsa dbStoreAdapter) Gas(meter GasMeter, config GasConfig) KVStore {
	return NewGasKVStore(meter, config, dsa)
}

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ KVStore = dbStoreAdapter{}

//----------------------------------------
// commitDBStoreWrapper should only be used for simulation/debugging,
// as it doesn't compute any commit hash, and it cannot load older state.

// Wrapper type for dbm.Db with implementation of KVStore
type commitDBStoreAdapter struct {
	dbStoreAdapter
}

func (cdsa commitDBStoreAdapter) Commit() CommitID {
	return CommitID{
		Version: -1,
		Hash:    []byte("FAKE_HASH"),
	}
}

func (cdsa commitDBStoreAdapter) LastCommitID() CommitID {
	return CommitID{
		Version: -1,
		Hash:    []byte("FAKE_HASH"),
	}
}

func (cdsa commitDBStoreAdapter) SetPruning(_ PruningStrategy) {}
