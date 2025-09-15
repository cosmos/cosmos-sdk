package prefix

import (
	"math/rand"
	"testing"

	"cosmossdk.io/store/memstore"
	"github.com/stretchr/testify/require"
)

// TestPrefixMemStoreBatch_BasicOperations tests basic operations (Get, Set, Delete)
func TestPrefixMemStoreBatch_BasicOperations(t *testing.T) {
	// Create basic MemStoreStore
	tree := memstore.NewMemStoreManager()

	// Set up initial data
	setupBatch := tree.Branch()
	setupBatch.Set(bz("prefixkey1"), "value1")
	setupBatch.Set(bz("prefix"), "root")
	setupBatch.Set(bz("prefixkey2"), "value2")
	setupBatch.Set(bz("other"), "other")
	setupBatch.Commit()
	tree.Commit(1)

	// Create prefix batch
	prefix := bz("prefix")
	batch := tree.Branch()
	prefixBatch := NewMemStore(batch, prefix)

	// When the key is prefix, the value should be root
	require.Equal(t, "root", prefixBatch.Get(bz("")).(string))

	// prefixkey1 -> key1, value should be correctly fetched
	require.Equal(t, "value1", prefixBatch.Get(bz("key1")).(string))

	// prefixkey2 -> key2, value should be correctly fetched
	require.Equal(t, "value2", prefixBatch.Get(bz("key2")).(string))

	// Keys outside of prefix range should not be fetched
	require.Nil(t, prefixBatch.Get(bz("other")))

	// Test nil key
	require.Panics(t, func() {
		prefixBatch.Get(nil)
	})

	// Test setting values
	prefixBatch.Set(bz("newkey"), "newvalue")
	require.Equal(t, "newvalue", prefixBatch.Get(bz("newkey")).(string))

	// Panic when setting nil key
	require.Panics(t, func() {
		prefixBatch.Set(nil, "value")
	})

	// After commit, verify with a new batch
	prefixBatch.Commit()
	tree.Commit(1)

	// Create a new batch
	NewBranch := tree.Branch()
	newPrefixBatch := NewMemStore(NewBranch, prefix)
	require.Equal(t, "newvalue", newPrefixBatch.Get(bz("newkey")).(string))

	// Verify with prefixed key in base store
	baseBatch := tree.Branch()
	prefixedKey := append(prefix, bz("newkey")...)
	require.Equal(t, "newvalue", baseBatch.Get(prefixedKey))

	// Delete test
	deleteBatch := tree.Branch()
	deletePrefixBatch := NewMemStore(deleteBatch, prefix)
	deletePrefixBatch.Delete(bz("newkey"))
	deletePrefixBatch.Commit()
	tree.Commit(1)

	// Check
	checkBatch := tree.Branch()
	checkPrefixBatch := NewMemStore(checkBatch, prefix)
	require.Nil(t, checkPrefixBatch.Get(bz("newkey")))

	// Verify in base store as well
	baseBatch = tree.Branch()
	require.Nil(t, baseBatch.Get(prefixedKey))
}

// TestPrefixMemStoreBatch_Iterator tests the iterator functionality
func TestPrefixMemStoreBatch_Iterator(t *testing.T) {
	// Setup base store
	tree := memstore.NewMemStoreManager()
	prefix := bz("prefix")

	// Set data
	setupBatch := tree.Branch()
	setupBatch.Set(bz("prefix1"), "value1")
	setupBatch.Set(bz("prefix2"), "value2")
	setupBatch.Set(bz("prefix3"), "value3")
	setupBatch.Set(bz("other"), "other")
	setupBatch.Commit()
	tree.Commit(1)

	// Create prefix batch
	batch := tree.Branch()
	prefixBatch := NewMemStore(batch, prefix)

	// Prefix store should iterate only items starting with the prefix
	iter := prefixBatch.Iterator(nil, nil)

	// Check the start and end domain
	start, end := iter.Domain()
	require.Nil(t, start)
	require.Nil(t, end)

	// Iterate through items starting with prefix
	count := 0
	for ; iter.Valid(); iter.Next() {
		count++
		key := iter.Key()
		value := iter.Value().(string)

		// Check keys without prefix
		switch string(key) {
		case "1":
			require.Equal(t, "value1", value)
		case "2":
			require.Equal(t, "value2", value)
		case "3":
			require.Equal(t, "value3", value)
		default:
			t.Errorf("Unexpected key: %s", key)
		}
	}

	// Ensure only 3 items starting with the prefix are checked
	require.Equal(t, 3, count)

	// Clean up resources
	iter.Close()

	// Range-based iterator test
	iter = prefixBatch.Iterator(bz("1"), bz("3"))

	// Check domain
	start, end = iter.Domain()
	require.Equal(t, bz("1"), start)
	require.Equal(t, bz("3"), end)

	// Verify range items
	count = 0
	for ; iter.Valid(); iter.Next() {
		count++
		key := iter.Key()
		value := iter.Value().(string)

		switch string(key) {
		case "1":
			require.Equal(t, "value1", value)
		case "2":
			require.Equal(t, "value2", value)
		default:
			t.Errorf("Unexpected key: %s", key)
		}
	}

	// Only 2 items should be in the range
	require.Equal(t, 2, count)

	// Clean up resources
	iter.Close()
}

// TestPrefixMemStoreBatch_ReverseIterator tests reverse iterator functionality
func TestPrefixMemStoreBatch_ReverseIterator(t *testing.T) {
	// Setup base store
	tree := memstore.NewMemStoreManager()
	prefix := bz("prefix")

	// Set data
	setupBatch := tree.Branch()
	setupBatch.Set(bz("prefix1"), "value1")
	setupBatch.Set(bz("prefix2"), "value2")
	setupBatch.Set(bz("prefix3"), "value3")
	setupBatch.Set(bz("other"), "other")
	setupBatch.Commit()
	tree.Commit(1)

	// Create prefix batch
	batch := tree.Branch()
	prefixBatch := NewMemStore(batch, prefix)

	// Iterate in reverse over the entire range
	iter := prefixBatch.ReverseIterator(nil, nil)

	// Check the domain
	start, end := iter.Domain()
	require.Nil(t, start)
	require.Nil(t, end)

	// Reverse iteration
	expectedKeys := []string{"3", "2", "1"}
	count := 0

	for ; iter.Valid(); iter.Next() {
		require.Equal(t, expectedKeys[count], string(iter.Key()))
		count++
	}

	require.Equal(t, 3, count)

	// Clean up resources
	iter.Close()

	// Range-based reverse iterator
	iter = prefixBatch.ReverseIterator(bz("1"), bz("3"))

	// Check domain
	start, end = iter.Domain()
	require.Equal(t, bz("1"), start)
	require.Equal(t, bz("3"), end)

	// Verify range items in reverse
	expectedKeys = []string{"2", "1"}
	count = 0

	for ; iter.Valid(); iter.Next() {
		require.Equal(t, expectedKeys[count], string(iter.Key()))
		count++
	}

	require.Equal(t, 2, count)

	// Clean up resources
	iter.Close()
}

// TestPrefixMemStoreBatch_NestedBatch tests nested batch operations
func TestPrefixMemStoreBatch_NestedBatch(t *testing.T) {
	// Create base store
	tree := memstore.NewMemStoreManager()
	prefix := bz("prefix")

	// Create parent batch
	baseBatch := tree.Branch()
	parentBatch := NewMemStore(baseBatch, prefix)

	// Set data in parent batch
	parentBatch.Set(bz("parent"), "parentvalue")

	// Create and test nested batch (scoped for resource management)
	{
		nestedBatch := parentBatch.Branch()

		// Verify parent value in nested batch
		require.Equal(t, "parentvalue", nestedBatch.Get(bz("parent")).(string))

		// Set new value in nested batch
		nestedBatch.Set(bz("nested"), "nestedvalue")

		// Nested batch can see the value
		require.Equal(t, "nestedvalue", nestedBatch.Get(bz("nested")).(string))

		// Parent batch should not see it yet
		require.Nil(t, parentBatch.Get(bz("nested")))

		// Commit changes from nested batch to parent
		nestedBatch.Commit()
	}

	// After commit, parent batch should see nested changes
	require.Equal(t, "nestedvalue", parentBatch.Get(bz("nested")).(string))

	// Commit parent batch
	parentBatch.Commit()
	tree.Commit(1)

	// Verify commit in new batch
	checkBaseBatch := tree.Branch()
	checkBatch := NewMemStore(checkBaseBatch, prefix)
	require.Equal(t, "parentvalue", checkBatch.Get(bz("parent")).(string))
	require.Equal(t, "nestedvalue", checkBatch.Get(bz("nested")).(string))

	// Verify in base store (with prefix)
	verifyBatch := tree.Branch()
	require.Equal(t, "parentvalue", verifyBatch.Get(append(prefix, bz("parent")...)))
	require.Equal(t, "nestedvalue", verifyBatch.Get(append(prefix, bz("nested")...)))
}

// TestPrefixMemStoreBatch_EdgeCases tests edge cases
func TestPrefixMemStoreBatch_EdgeCases(t *testing.T) {
	// Create base store
	tree := memstore.NewMemStoreManager()

	// Test empty prefix
	emptyPrefix := []byte{}
	batch := tree.Branch()
	emptyPrefixBatch := NewMemStore(batch, emptyPrefix)
	emptyPrefixBatch.Set(bz("key"), "value")
	emptyPrefixBatch.Commit()
	tree.Commit(1)

	// Verify in base store
	baseBatch := tree.Branch()
	require.Equal(t, "value", baseBatch.Get(bz("key")))

	// Test edge case prefix (last byte is 0xFF)
	edgePrefix := []byte{0x01, 0xFF}
	edgeBatch := tree.Branch()
	edgePrefixBatch := NewMemStore(edgeBatch, edgePrefix)
	edgePrefixBatch.Set(bz("key"), "edgevalue")
	edgePrefixBatch.Commit()
	tree.Commit(1)

	// Verify in base store
	baseBatch = tree.Branch()
	require.Equal(t, "edgevalue", baseBatch.Get(append(edgePrefix, bz("key")...)))

	// Verify different prefix can't access
	wrongPrefix := []byte{0x01, 0xFE}
	wrongBaseBatch := tree.Branch()
	wrongPrefixBatch := NewMemStore(wrongBaseBatch, wrongPrefix)
	require.Nil(t, wrongPrefixBatch.Get(bz("key")))
}

// TestPrefixMemStoreBatch_SnapshotBranch tests snapshot batch functionality
func TestPrefixMemStoreBatch_SnapshotBranch(t *testing.T) {
	// Create base store
	tree := memstore.NewMemStoreManager()
	tree.SetSnapshotPoolLimit(10)

	// Set initial data (height 1)
	batch1 := tree.Branch()
	prefix := bz("prefix")
	prefixedKey := append(prefix, bz("key")...)
	batch1.Set(prefixedKey, "value1")
	batch1.Commit()
	tree.Commit(1)

	// Set data with height 2
	batch2 := tree.Branch()
	batch2.Set(prefixedKey, "value2")
	batch2.Commit()
	tree.Commit(2)

	// Get snapshot batch for height 1
	snapshot1, ok := tree.GetSnapshotBranch(1)
	require.True(t, ok)
	prefixSnapshot1 := NewMemStore(snapshot1, prefix)
	require.Equal(t, "value1", prefixSnapshot1.Get(bz("key")))

	// Get snapshot batch for height 2
	snapshot2, ok := tree.GetSnapshotBranch(2)
	require.True(t, ok)
	prefixSnapshot2 := NewMemStore(snapshot2, prefix)
	require.Equal(t, "value2", prefixSnapshot2.Get(bz("key")))

	// Snapshot for non-existent height
	snapshot3, ok := tree.GetSnapshotBranch(3)
	require.False(t, ok)
	require.Nil(t, snapshot3)

	// Verify commit panic for snapshot
	require.Panics(t, func() {
		prefixSnapshot1.Commit()
	})
}

// Utility for generating random key-value pairs
func genRandomMemStoreKVPairs(n int) []struct {
	key   []byte
	value string
} {
	r := rand.New(rand.NewSource(1))
	kvs := make([]struct {
		key   []byte
		value string
	}, n)

	for i := 0; i < n; i++ {
		kl := r.Intn(10) + 2 // key length min 2, max 11
		vl := r.Intn(10) + 2 // value length min 2, max 11

		kvs[i].key = make([]byte, kl)
		value := make([]byte, vl)

		for j := 0; j < kl; j++ {
			kvs[i].key[j] = byte(r.Intn(255) + 1) // non-zero bytes
		}

		for j := 0; j < vl; j++ {
			value[j] = byte(r.Intn(255) + 1) // non-zero bytes
		}

		kvs[i].value = string(value)
	}

	return kvs
}

// TestPrefixMemStoreBatch_RandomData tests wide range of tests with random data
func TestPrefixMemStoreBatch_RandomData(t *testing.T) {
	// Create base store
	tree := memstore.NewMemStoreManager()
	prefix := bz("prefix")

	// Generate random key-value pairs
	kvs := genRandomMemStoreKVPairs(50)

	// Set data
	baseBatch := tree.Branch()
	batch := NewMemStore(baseBatch, prefix)
	for _, kv := range kvs {
		batch.Set(kv.key, kv.value)
	}
	batch.Commit()
	tree.Commit(1)

	// Verify all key-value pairs
	checkBaseBatch := tree.Branch()
	checkBatch := NewMemStore(checkBaseBatch, prefix)
	for _, kv := range kvs {
		require.Equal(t, kv.value, checkBatch.Get(kv.key))
	}

	// Verify in base store (with prefix)
	verifyBatch := tree.Branch()
	for _, kv := range kvs {
		prefixedKey := append(prefix, kv.key...)
		require.Equal(t, kv.value, verifyBatch.Get(prefixedKey))
	}

	// Delete half of the keys
	deleteBaseBatch := tree.Branch()
	deleteBatch := NewMemStore(deleteBaseBatch, prefix)
	for i := 0; i < len(kvs)/2; i++ {
		deleteBatch.Delete(kvs[i].key)
	}
	deleteBatch.Commit()
	tree.Commit(1)

	// Verify deletions
	finalBaseBatch := tree.Branch()
	finalBatch := NewMemStore(finalBaseBatch, prefix)
	for i, kv := range kvs {
		if i < len(kvs)/2 {
			require.Nil(t, finalBatch.Get(kv.key))
		} else {
			require.Equal(t, kv.value, finalBatch.Get(kv.key))
		}
	}
}
