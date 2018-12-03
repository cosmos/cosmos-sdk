package store

import (
	"github.com/cosmos/cosmos-sdk/store/types"
	stypes "github.com/cosmos/cosmos-sdk/store/types"
)

// Import cosmos-sdk/types/store.go for convenience.
// nolint
type (
	PruningStrategy  = types.PruningStrategy
	Store            = types.Store
	Committer        = types.Committer
	CommitStore      = types.CommitStore
	MultiStore       = types.MultiStore
	CacheMultiStore  = types.CacheMultiStore
	CommitMultiStore = types.CommitMultiStore
	KVStore          = types.KVStore
	KVPair           = types.KVPair
	Iterator         = types.Iterator
	CacheKVStore     = types.CacheKVStore
	CommitKVStore    = types.CommitKVStore
	CacheWrapper     = types.CacheWrapper
	CacheWrap        = types.CacheWrap
	CommitID         = types.CommitID
	StoreKey         = types.StoreKey
	StoreType        = types.StoreType
	Queryable        = types.Queryable
	TraceContext     = types.TraceContext
	Gas              = stypes.Gas
	GasMeter         = types.GasMeter
	GasConfig        = stypes.GasConfig
)
