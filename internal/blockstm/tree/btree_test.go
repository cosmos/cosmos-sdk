package tree

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/test-go/testify/require"
)

type batchTestItem struct {
	key   []byte
	value *int
}

func (item batchTestItem) GetKey() []byte {
	return item.key
}

func TestBTreeBatchGetOrDefault(t *testing.T) {
	ctx := context.Background()
	tree := NewBTree(KeyItemLess[batchTestItem], 4)
	existing := tree.GetOrDefault(ctx, batchTestItem{key: []byte("a")}, func(item *batchTestItem) {
		item.value = new(int)
	})

	var initialized atomic.Int64
	results := tree.BatchGetOrDefault(ctx, []batchTestItem{
		{key: []byte("a")},
		{key: []byte("b")},
		{key: []byte("b")},
		{key: []byte("c")},
	}, func(item *batchTestItem) {
		initialized.Add(1)
		item.value = new(int)
	})

	require.Equal(t, int64(2), initialized.Load())
	require.True(t, existing.value == results[0].value)
	require.True(t, results[1].value == results[2].value)

	var keys []string
	tree.Scan(ctx, func(item batchTestItem) bool {
		keys = append(keys, string(item.key))
		return true
	})
	require.Equal(t, []string{"a", "b", "c"}, keys)

	snapshot := tree.Load()
	tree.BatchGetOrDefault(ctx, []batchTestItem{{key: []byte("a")}, {key: []byte("c")}}, func(*batchTestItem) {
		t.Fatal("defaults called for an existing key")
	})
	require.True(t, snapshot == tree.Load(), "an all-existing batch should not publish a new snapshot")
}

func TestBTreeBatchGetOrDefaultConcurrentOverlap(t *testing.T) {
	const (
		workers       = 8
		keysPerWorker = 100
		keyCount      = 240
	)

	ctx := context.Background()
	tree := NewBTree(KeyItemLess[batchTestItem], 4)
	var (
		wg       sync.WaitGroup
		seen     sync.Map
		mismatch atomic.Bool
	)
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			items := make([]batchTestItem, keysPerWorker)
			for i := range items {
				items[i].key = []byte(fmt.Sprintf("%03d", (worker*31+i)%keyCount))
			}
			results := tree.BatchGetOrDefault(ctx, items, func(item *batchTestItem) {
				item.value = new(int)
			})
			for _, result := range results {
				key := string(result.key)
				if value, loaded := seen.LoadOrStore(key, result.value); loaded && value.(*int) != result.value {
					mismatch.Store(true)
				}
			}
		}(worker)
	}
	wg.Wait()

	require.False(t, mismatch.Load(), "overlapping batches returned different values for the same key")
	var count int
	tree.Scan(ctx, func(batchTestItem) bool {
		count++
		return true
	})
	require.Equal(t, keyCount, count)
}

func BenchmarkBTreeBatchGetOrDefault(b *testing.B) {
	ctx := context.Background()
	for _, size := range []int{1, 10, 100, 1000} {
		items := make([]batchTestItem, size)
		for i := range items {
			items[i].key = []byte(fmt.Sprintf("%08d", i))
		}
		fillDefaults := func(item *batchTestItem) {
			item.value = new(int)
		}

		b.Run(fmt.Sprintf("individual/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				tree := NewBTree(KeyItemLess[batchTestItem], 4)
				for _, item := range items {
					tree.GetOrDefault(ctx, item, fillDefaults)
				}
			}
		})
		b.Run(fmt.Sprintf("batch/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				tree := NewBTree(KeyItemLess[batchTestItem], 4)
				tree.BatchGetOrDefault(ctx, items, fillDefaults)
			}
		})
	}
}
