package tracekv

import (
	"io"
	"sync"

	"cosmossdk.io/store/types"
)

var (
	_ types.MultiStore      = (*MultiStore)(nil)
	_ types.CacheMultiStore = (*MultiStore)(nil)
)

// MultiStore wraps a types.MultiStore and traces all KVStore operations.
type MultiStore struct {
	parent types.MultiStore
	writer io.Writer

	mu      sync.Mutex
	context types.TraceContext
}

// NewMultiStore returns a new tracing MultiStore wrapper.
func NewMultiStore(parent types.MultiStore, writer io.Writer, tc types.TraceContext) *MultiStore {
	return &MultiStore{parent: parent, writer: writer, context: tc}
}

func (ms *MultiStore) GetStoreType() types.StoreType {
	return ms.parent.GetStoreType()
}

func (ms *MultiStore) CacheWrap() types.CacheWrap {
	return ms.CacheMultiStore().(types.CacheWrap)
}

func (ms *MultiStore) CacheMultiStore() types.CacheMultiStore {
	return NewMultiStore(ms.parent.CacheMultiStore(), ms.writer, ms.getTraceContext())
}

func (ms *MultiStore) CacheMultiStoreWithVersion(version int64) (types.CacheMultiStore, error) {
	cms, err := ms.parent.CacheMultiStoreWithVersion(version)
	if err != nil {
		return nil, err
	}
	return NewMultiStore(cms, ms.writer, ms.getTraceContext()), nil
}

func (ms *MultiStore) GetStore(key types.StoreKey) types.Store {
	return ms.parent.GetStore(key)
}

func (ms *MultiStore) GetKVStore(key types.StoreKey) types.KVStore {
	return NewStore(ms.parent.GetKVStore(key), ms.writer, ms.getTraceContext())
}

func (ms *MultiStore) GetObjKVStore(key types.StoreKey) types.ObjKVStore {
	return ms.parent.GetObjKVStore(key)
}

func (ms *MultiStore) LatestVersion() int64 {
	return ms.parent.LatestVersion()
}

// SetTraceContext merges the given context into the existing trace context.
// Existing keys are overwritten by new values.
func (ms *MultiStore) SetTraceContext(tc types.TraceContext) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.context = ms.context.Merge(tc)
}

// getTraceContext returns a copy of the current trace context.
func (ms *MultiStore) getTraceContext() types.TraceContext {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.context == nil {
		return nil
	}

	ctx := types.TraceContext{}
	for k, v := range ms.context {
		ctx[k] = v
	}

	return ctx
}

// Write implements CacheMultiStore. It delegates to the parent if it supports Write.
func (ms *MultiStore) Write() {
	if cms, ok := ms.parent.(types.CacheMultiStore); ok {
		cms.Write()
	}
}
