package cachemulti

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/store/dbadapter"
	"cosmossdk.io/store/types"
)

func TestStoreGetKVStore(t *testing.T) {
	require := require.New(t)

	s := Store{stores: map[types.StoreKey]types.CacheWrap{}}
	key := types.NewKVStoreKey("abc")
	errMsg := fmt.Sprintf("kv store with key %v has not been registered in stores", key)

	require.PanicsWithValue(errMsg,
		func() { s.GetStore(key) })

	require.PanicsWithValue(errMsg,
		func() { s.GetKVStore(key) })
}

// TestNewFromKVStoreIsolatesTraceContext locks in the fix for #25841: the
// trace-context map handed to NewFromKVStore must not be retained by
// reference, otherwise downstream Clone()/Merge iteration races with any
// caller still mutating their copy of the map (e.g. SetTracingContext or a
// sibling cms branched off the same parent).
func TestNewFromKVStoreIsolatesTraceContext(t *testing.T) {
	tc := types.TraceContext{"chain_id": "test-chain"}
	cms := NewFromKVStore(
		dbadapter.Store{DB: dbm.NewMemDB()},
		map[types.StoreKey]types.CacheWrapper{},
		nil,
		new(bytes.Buffer),
		tc,
	)

	// Mutating the input map after construction must not bleed into the
	// cms's private trace context.
	tc["chain_id"] = "OTHER"
	tc["leak"] = "value"

	require.Equal(t, "test-chain", cms.traceContext["chain_id"])
	_, leaked := cms.traceContext["leak"]
	require.False(t, leaked, "external mutation must not reach cms.traceContext")
}

// TestSetTracingContextDoesNotMutateInPlace pins the SetTracingContext fix
// for #25841 — the receiver returns a Store backed by a fresh map so prior
// branched cms keep pointing at the old map without observing the merge.
func TestSetTracingContextDoesNotMutateInPlace(t *testing.T) {
	tc := types.TraceContext{"chain_id": "test-chain"}
	cms := NewFromKVStore(
		dbadapter.Store{DB: dbm.NewMemDB()},
		map[types.StoreKey]types.CacheWrapper{},
		nil,
		new(bytes.Buffer),
		tc,
	)

	// Hold a reference to the map cms is using right now so we can prove
	// the merge does not write through it.
	originalMap := cms.traceContext

	cms2 := cms.SetTracingContext(types.TraceContext{"height": "42"}).(Store)

	// originalMap stays untouched.
	_, mutated := originalMap["height"]
	require.False(t, mutated, "SetTracingContext must not mutate the prior trace-context map")

	// cms2 sees the merged result.
	require.Equal(t, "test-chain", cms2.traceContext["chain_id"])
	require.Equal(t, "42", cms2.traceContext["height"])
}

// TestCacheMultiStoreConcurrentWithSetTracingContext is the direct regression
// test for #25841. Before the fix, running this under `go test -race` raised
// "fatal error: concurrent map iteration and map write" because branched
// cms instances and the parent shared one trace-context map; one goroutine
// iterating via Clone()/Merge collided with another writing via
// SetTracingContext.
func TestCacheMultiStoreConcurrentWithSetTracingContext(t *testing.T) {
	cms := NewFromKVStore(
		dbadapter.Store{DB: dbm.NewMemDB()},
		map[types.StoreKey]types.CacheWrapper{},
		nil,
		new(bytes.Buffer),
		types.TraceContext{"chain_id": "test-chain"},
	)

	const goroutines = 16
	const iterations = 200
	var wg sync.WaitGroup
	wg.Add(2 * goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = cms.CacheMultiStore()
			}
		}()
		go func(i int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = cms.SetTracingContext(types.TraceContext{
					fmt.Sprintf("k-%d-%d", i, j): "v",
				})
			}
		}(i)
	}
	wg.Wait()
}
