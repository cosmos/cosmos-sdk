package iavlx

import (
	"fmt"
	io "io"
	"sync"

	storetypes "cosmossdk.io/store/types"
)

type MultiTree struct {
	latestVersion int64
	trees         []storetypes.CacheKVStore   // always ordered by tree name
	treesByKey    map[storetypes.StoreKey]int // index of the trees by name
}

func (t *MultiTree) Write() {
	var wg sync.WaitGroup
	for _, tree := range t.trees {
		// TODO check if trees are dirty before spinning off a goroutine
		wg.Add(1)
		go func(t storetypes.CacheKVStore) {
			defer wg.Done()
			t.Write()
		}(tree)
	}
	wg.Wait()
}

func (t *MultiTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

func (t *MultiTree) CacheWrap() storetypes.CacheWrap {
	return t.CacheMultiStore()
}

func (t *MultiTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO implement tracing
	return t.CacheWrap()
}

func (t *MultiTree) CacheMultiStore() storetypes.CacheMultiStore {
	wrapped := &MultiTree{
		treesByKey: t.treesByKey,
		trees:      make([]storetypes.CacheKVStore, len(t.trees)),
	}
	for i, tree := range t.trees {
		wrapped.trees[i] = tree.CacheWrap().(storetypes.CacheKVStore)
	}
	return wrapped
}

func (t *MultiTree) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	return nil, fmt.Errorf("CacheMultiStoreWithVersion can only be called on CommitMultiStore")
}

func (t *MultiTree) GetStore(key storetypes.StoreKey) storetypes.Store {
	return t.trees[t.treesByKey[key]]
}

func (t *MultiTree) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	index, ok := t.treesByKey[key]
	if !ok {
		panic(fmt.Sprintf("store not found for key: %s (key type: %T)", key.Name(), key))
	}
	if index >= len(t.trees) {
		panic(fmt.Sprintf("store index %d out of bounds for key %s (trees length: %d)", index, key.Name(), len(t.trees)))
	}
	return t.trees[index]
}

func (t *MultiTree) TracingEnabled() bool {
	return false
}

func (t *MultiTree) SetTracer(w io.Writer) storetypes.MultiStore {
	// TODO implement me
	panic("implement me")
}

func (t *MultiTree) SetTracingContext(context storetypes.TraceContext) storetypes.MultiStore {
	// TODO implement me
	panic("implement me")
}

func (t *MultiTree) LatestVersion() int64 {
	return t.latestVersion
}

var _ storetypes.CacheMultiStore = &MultiTree{}
