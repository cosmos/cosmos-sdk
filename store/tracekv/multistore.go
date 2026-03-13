package tracekv

import (
	"io"
	"maps"
	"sync"

	"cosmossdk.io/store/types"
)

type SetTracingContext interface {
	SetTracingContext(types.TraceContext)
}

type tracerMultiStoreBase struct {
	writer            io.Writer
	traceContext      types.TraceContext
	traceContextMutex sync.Mutex
}

func newTracerMultiStoreBase(writer io.Writer, traceContext types.TraceContext) *tracerMultiStoreBase {
	return &tracerMultiStoreBase{writer: writer, traceContext: traceContext}
}

func (rs *tracerMultiStoreBase) SetTracingContext(tc types.TraceContext) {
	rs.traceContextMutex.Lock()
	defer rs.traceContextMutex.Unlock()
	rs.traceContext = rs.traceContext.Merge(tc)
}

func (rs *tracerMultiStoreBase) getTracingContext() types.TraceContext {
	rs.traceContextMutex.Lock()
	defer rs.traceContextMutex.Unlock()

	if rs.traceContext == nil {
		return nil
	}

	ctx := types.TraceContext{}
	maps.Copy(ctx, rs.traceContext)

	return ctx
}

type cacheMultiStore struct {
	types.CacheMultiStore
	*tracerMultiStoreBase
}

func (m *cacheMultiStore) GetKVStore(key types.StoreKey) types.KVStore {
	// tracing was only really ever implemented for KVStore's
	return NewStore(m.CacheMultiStore.GetKVStore(key), m.writer, m.getTracingContext())
}

type CommitMultiStore struct {
	types.CommitMultiStore
	*tracerMultiStoreBase
}

func NewCommitMultiStore(cms types.CommitMultiStore, writer io.Writer, ctx types.TraceContext) *CommitMultiStore {
	return &CommitMultiStore{CommitMultiStore: cms, tracerMultiStoreBase: newTracerMultiStoreBase(writer, ctx)}
}

func (cms *CommitMultiStore) CacheMultiStore() types.CacheMultiStore {
	return &cacheMultiStore{
		CacheMultiStore:      cms.CommitMultiStore.CacheMultiStore(),
		tracerMultiStoreBase: cms.tracerMultiStoreBase,
	}
}

func (cms *CommitMultiStore) CacheMultiStoreWithVersion(version int64) (types.CacheMultiStore, error) {
	ms, err := cms.CommitMultiStore.CacheMultiStoreWithVersion(version)
	if err != nil {
		return nil, err
	}
	return &cacheMultiStore{
		tracerMultiStoreBase: cms.tracerMultiStoreBase,
		CacheMultiStore:      ms,
	}, nil
}

func (cms *CommitMultiStore) GetKVStore(key types.StoreKey) types.KVStore {
	return NewStore(cms.CommitMultiStore.GetKVStore(key), cms.writer, cms.getTracingContext())
}
