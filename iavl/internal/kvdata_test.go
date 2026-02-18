package internal

//import (
//	"testing"
//
//	"github.com/stretchr/testify/require"
//)

//type kvDataWriterHelper struct {
//	files *ChangesetFiles
//	*KVDataWriter
//}
//
//func openTestKVDataWriter(t *testing.T) *kvDataWriterHelper {
//	t.Helper()
//
//	files, err := CreateChangesetFiles(t.TempDir(), 1, 0)
//	require.NoError(t, err)
//	t.Cleanup(func() {
//		require.NoError(t, files.Close())
//	})
//	writer := NewKVDataWriter(files.KVDataFile())
//	return &kvDataWriterHelper{
//		files:        files,
//		KVDataWriter: writer,
//	}
//}
//
//func (h *kvDataWriterHelper) openReader(t *testing.T) *KVDataReader {
//	t.Helper()
//
//	require.NoError(t, h.writer.Flush())
//	rdr, err := NewKVDataReader(h.files.KVDataFile())
//	require.NoError(t, err)
//	t.Cleanup(func() {
//		require.NoError(t, rdr.Close())
//	})
//	return rdr
//}
//
//func TestKVData_WAL(t *testing.T) {
//	writer := openTestKVDataWriter(t)
//
//	// Write WAL start
//	require.NoError(t, writer.WriteStartWAL(42))
//	// Write WAL set with short key
//	shortKey := []byte("key") // short key, should not be cached
//	shortValue := []byte("value")
//	shortKeyOffset, shortValueOffset, err := writer.WriteWALSet(shortKey, shortValue)
//	require.NoError(t, err)
//	// Write WAL set with longer key
//	longerKey := []byte("longerKey") // longer key, should be cached
//	longerValue := []byte("longerValue")
//	longKeyOffset, longerValueOffset, err := writer.WriteWALSet(longerKey, longerValue)
//	require.NoError(t, err)
//	// Write WAL delete
//	oldKey := []byte("oldKey")
//	require.NoError(t, writer.WriteWALDelete(oldKey))
//	// Write WAL commit
//	require.NoError(t, writer.WriteWALCommit(42))
//
//	// Write short key again to test caching
//	shortValue2 := []byte("value2")
//	shortKeyOffset2, shortValue2Offset, err := writer.WriteWALSet(shortKey, shortValue2)
//	require.NoError(t, err)
//	// short key should NOT be cached
//	require.NotEqual(t, shortKeyOffset, shortKeyOffset2)
//	// Write longer key again to test caching
//	longerValue2 := []byte("longerValue2")
//	longKeyOffset2, longerValue2Offset, err := writer.WriteWALSet(longerKey, longerValue2)
//	require.NoError(t, err)
//	// longer key should be cached
//	require.Equal(t, longKeyOffset, longKeyOffset2)
//
//	// Write WAL Updates
//	memKey1 := []byte("memKey1")
//	memValue1 := []byte("memValue1")
//	memNode1 := &MemNode{
//		key:   memKey1,
//		value: memValue1,
//	}
//	memValue2 := []byte("memValue2")
//	memNode2 := &MemNode{
//		key:   longerKey, // should use cached key offset
//		value: memValue2,
//	}
//	reinsertedValue := []byte("valueReinserted")
//	memNode3 := &MemNode{
//		key:   oldKey,
//		value: reinsertedValue,
//	}
//	err = writer.WriteWALUpdates([]KVUpdate{
//		{
//			SetNode: memNode1,
//		},
//		{
//			DeleteKey: oldKey,
//		},
//		{
//			SetNode: memNode2,
//		},
//		{
//			SetNode: memNode3,
//		},
//	})
//	require.NoError(t, err)
//	require.NotZero(t, memNode1.walKeyOffset)
//	require.NotZero(t, memNode1.walValueOffset)
//	require.NotZero(t, memNode2.walKeyOffset)
//	require.NotZero(t, memNode2.walValueOffset)
//	// memNode2 should use cached key offset
//	require.Equal(t, longKeyOffset, memNode2.walKeyOffset)
//
//	require.NoError(t, writer.WriteWALCommit(43))
//
//	// test caching again with some blobs
//	blobKeyOffset, err := writer.WriteKeyBlob(longerKey)
//	require.NoError(t, err)
//	// should use cached offset
//	require.Equal(t, longKeyOffset, blobKeyOffset)
//
//	blobKeyOffset2, shortValueOffset2, err := writer.WriteKeyValueBlobs(shortKey, shortValue)
//	require.NoError(t, err)
//	// short key should NOT be cached
//	require.NotEqual(t, shortKeyOffset, blobKeyOffset2)
//
//	// write invalid updates, should error
//	require.Error(t, writer.WriteWALUpdates([]KVUpdate{
//		{},
//	}))
//	require.Error(t, writer.WriteWALUpdates([]KVUpdate{
//		{
//			DeleteKey: shortKey,
//			SetNode: &MemNode{
//				key:   shortKey,
//				value: shortValue,
//			},
//		},
//	}))
//
//	// write an empty commit
//	require.NoError(t, writer.WriteWALCommit(44))
//
//	// Test empty key/value edge cases via WriteWALUpdates
//	emptyKey := []byte{}
//	emptyValue := []byte{}
//	emptyMemNode := &MemNode{key: emptyKey, value: emptyValue}
//	err = writer.WriteWALUpdates([]KVUpdate{
//		{DeleteKey: emptyKey},
//		{SetNode: emptyMemNode},
//	})
//	require.NoError(t, err)
//	require.NotZero(t, emptyMemNode.walKeyOffset)
//	require.NotZero(t, emptyMemNode.walValueOffset)
//	require.NoError(t, writer.WriteWALCommit(45))
//
//	// open reader
//	r := writer.openReader(t)
//	// Verify that the reader has a WAL
//	hasWal, startVersion := r.HasWAL()
//	require.True(t, hasWal)
//	require.Equal(t, uint64(42), startVersion)
//
//	// Verify via walReader
//	wr, err := r.ReadWAL()
//	require.NoError(t, err)
//	require.Equal(t, uint64(42), wr.Version)
//	// Entry 1: WAL Set short key
//	entryType, ok, err := wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntrySet, entryType)
//	require.Equal(t, []byte("key"), wr.Key)
//	require.Equal(t, []byte("value"), wr.Value)
//
//	// Entry 2: WAL Set longer key
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntrySet, entryType)
//	require.Equal(t, []byte("longerKey"), wr.Key)
//	require.Equal(t, []byte("longerValue"), wr.Value)
//
//	// Entry 3: WAL Delete
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntryDelete, entryType)
//	require.Equal(t, []byte("oldKey"), wr.Key)
//
//	// Entry 4: WAL Commit
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntryCommit, entryType)
//	require.Equal(t, uint64(42), wr.Version)
//
//	// Entry 5: WAL Set short key again (not cached)
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntrySet, entryType)
//	require.Equal(t, []byte("key"), wr.Key)
//	require.Equal(t, []byte("value2"), wr.Value)
//
//	// Entry 6: WAL Set longer key again (cached)
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntrySet|WALFlagCachedKey, entryType)
//	require.Equal(t, []byte("longerKey"), wr.Key)
//	require.Equal(t, []byte("longerValue2"), wr.Value)
//
//	// Entry 7-10: WAL Updates
//	for i, expected := range []struct {
//		entryType WALEntryType
//		key       []byte
//		value     []byte
//	}{
//		{WALEntrySet, memKey1, memValue1},
//		{WALEntryDelete | WALFlagCachedKey, oldKey, nil},
//		{WALEntrySet | WALFlagCachedKey, longerKey, memValue2},
//		{WALEntrySet | WALFlagCachedKey, oldKey, reinsertedValue},
//	} {
//		entryType, ok, err = wr.Next()
//		require.NoError(t, err, "WAL Update entry %d", i)
//		require.True(t, ok, "WAL Update entry %d", i)
//		require.Equal(t, expected.entryType, entryType, "WAL Update entry %d type", i)
//		require.Equal(t, expected.key, wr.Key, "WAL Update entry %d key", i)
//		if expected.value != nil {
//			require.Equal(t, expected.value, wr.Value, "WAL Update entry %d value", i)
//		}
//	}
//	// Entry 11: WAL Commit
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntryCommit, entryType)
//	require.Equal(t, uint64(43), wr.Version)
//	// Entry 12: WAL Commit (empty)
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntryCommit, entryType)
//	require.Equal(t, uint64(44), wr.Version)
//
//	// Entry 13: WAL Delete with empty key
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntryDelete, entryType)
//	require.Equal(t, emptyKey, wr.Key)
//
//	// Entry 14: WAL Set with empty key and empty value
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntrySet, entryType)
//	require.Equal(t, emptyKey, wr.Key)
//	require.Equal(t, emptyValue, wr.Value)
//
//	// Entry 15: WAL Commit
//	entryType, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.True(t, ok)
//	require.Equal(t, WALEntryCommit, entryType)
//	require.Equal(t, uint64(45), wr.Version)
//
//	// No more entries
//	_, ok, err = wr.Next()
//	require.NoError(t, err)
//	require.False(t, ok)
//
//	// Check that all offsets are readable
//	shortKeyRead, err := r.UnsafeReadBlob(int(shortKeyOffset))
//	require.NoError(t, err)
//	require.Equal(t, shortKey, shortKeyRead)
//	shortValueRead, err := r.UnsafeReadBlob(int(shortValueOffset))
//	require.NoError(t, err)
//	require.Equal(t, shortValue, shortValueRead)
//	longKeyRead, err := r.UnsafeReadBlob(int(longKeyOffset))
//	require.NoError(t, err)
//	require.Equal(t, longerKey, longKeyRead)
//	longValueRead, err := r.UnsafeReadBlob(int(longerValueOffset))
//	require.NoError(t, err)
//	require.Equal(t, longerValue, longValueRead)
//	shortKeyRead2, err := r.UnsafeReadBlob(int(shortKeyOffset2))
//	require.NoError(t, err)
//	require.Equal(t, shortKey, shortKeyRead2)
//	longValueRead2, err := r.UnsafeReadBlob(int(longerValue2Offset))
//	require.NoError(t, err)
//	require.Equal(t, longerValue2, longValueRead2)
//	shorterValue2Read, err := r.UnsafeReadBlob(int(shortValue2Offset))
//	require.NoError(t, err)
//	require.Equal(t, shortValue2, shorterValue2Read)
//	blobKeyRead, err := r.UnsafeReadBlob(int(blobKeyOffset))
//	require.NoError(t, err)
//	require.Equal(t, longerKey, blobKeyRead)
//	blobKeyRead2, err := r.UnsafeReadBlob(int(blobKeyOffset2))
//	require.NoError(t, err)
//	require.Equal(t, shortKey, blobKeyRead2)
//	shortValueRead2, err := r.UnsafeReadBlob(int(shortValueOffset2))
//	require.NoError(t, err)
//	require.Equal(t, shortValue, shortValueRead2)
//	// also check all memNode offsets
//	memKey1Read, err := r.UnsafeReadBlob(int(memNode1.walKeyOffset))
//	require.NoError(t, err)
//	require.Equal(t, memKey1, memKey1Read)
//	memValue1Read, err := r.UnsafeReadBlob(int(memNode1.walValueOffset))
//	require.NoError(t, err)
//	require.Equal(t, memValue1, memValue1Read)
//	memKey2Read, err := r.UnsafeReadBlob(int(memNode2.walKeyOffset))
//	require.NoError(t, err)
//	require.Equal(t, longerKey, memKey2Read)
//	memValue2Read, err := r.UnsafeReadBlob(int(memNode2.walValueOffset))
//	require.NoError(t, err)
//	require.Equal(t, memValue2, memValue2Read)
//	memKey3Read, err := r.UnsafeReadBlob(int(memNode3.walKeyOffset))
//	require.NoError(t, err)
//	require.Equal(t, oldKey, memKey3Read)
//	memValue3Read, err := r.UnsafeReadBlob(int(memNode3.walValueOffset))
//	require.NoError(t, err)
//	require.Equal(t, reinsertedValue, memValue3Read)
//	// check empty memNode offsets
//	emptyKeyRead, err := r.UnsafeReadBlob(int(emptyMemNode.walKeyOffset))
//	require.NoError(t, err)
//	require.Equal(t, emptyKey, emptyKeyRead)
//	emptyValueRead, err := r.UnsafeReadBlob(int(emptyMemNode.walValueOffset))
//	require.NoError(t, err)
//	require.Equal(t, emptyValue, emptyValueRead)
//}
//
//func TestKVData_BlobStore(t *testing.T) {
//	writer := openTestKVDataWriter(t)
//	// Write some blobs
//	key1 := []byte("key") // short key, should not be cached
//	key2 := []byte("a much longer key 2 that should be cached")
//	value1 := []byte("value1")
//	value2 := []byte("value2")
//
//	key1Offset, err := writer.WriteKeyBlob(key1)
//	require.NoError(t, err)
//	key2Offset, err := writer.WriteKeyBlob(key2)
//	require.NoError(t, err)
//	key1Offset2, value1Offset, err := writer.WriteKeyValueBlobs(key1, value1)
//	require.NoError(t, err)
//	key2Offset2, value2Offset, err := writer.WriteKeyValueBlobs(key2, value2)
//	require.NoError(t, err)
//	// key1 should NOT be cached
//	require.NotEqual(t, key1Offset, key1Offset2)
//	// key2 should be cached
//	require.Equal(t, key2Offset, key2Offset2)
//
//	// Test empty key/value edge cases
//	emptyKey := []byte{}
//	emptyValue := []byte{}
//	emptyKeyOffset, err := writer.WriteKeyBlob(emptyKey)
//	require.NoError(t, err)
//	emptyKeyOffset2, emptyValueOffset, err := writer.WriteKeyValueBlobs(emptyKey, emptyValue)
//	require.NoError(t, err)
//	// empty key should NOT be cached (len < 4)
//	require.NotEqual(t, emptyKeyOffset, emptyKeyOffset2)
//
//	// verify we're not in WAL mode and that WAL operations fail
//	require.False(t, writer.IsInWALMode())
//	require.Error(t, writer.WriteStartWAL(1))
//	require.Error(t, writer.WriteWALDelete(key1))
//	_, _, err = writer.WriteWALSet(key1, value1)
//	require.Error(t, err)
//	require.Error(t, writer.WriteWALUpdates([]KVUpdate{}))
//	require.Error(t, writer.WriteWALCommit(1))
//
//	// open reader
//	r := writer.openReader(t)
//
//	// verify that the reader does not have a WAL
//	hasWal, _ := r.HasWAL()
//	require.False(t, hasWal)
//
//	_, err = r.ReadWAL()
//	require.Error(t, err)
//
//	// check that all offsets are readable
//	key1Read, err := r.UnsafeReadBlob(int(key1Offset))
//	require.NoError(t, err)
//	require.Equal(t, key1, key1Read)
//	key2Read, err := r.UnsafeReadBlob(int(key2Offset))
//	require.NoError(t, err)
//	require.Equal(t, key2, key2Read)
//	key1Read2, err := r.UnsafeReadBlob(int(key1Offset2))
//	require.NoError(t, err)
//	require.Equal(t, key1, key1Read2)
//	value1Read, err := r.UnsafeReadBlob(int(value1Offset))
//	require.NoError(t, err)
//	require.Equal(t, value1, value1Read)
//	key2Read2, err := r.UnsafeReadBlob(int(key2Offset2))
//	require.NoError(t, err)
//	require.Equal(t, key2, key2Read2)
//	value2Read, err := r.UnsafeReadBlob(int(value2Offset))
//	require.NoError(t, err)
//	require.Equal(t, value2, value2Read)
//	// check empty key/value offsets
//	emptyKeyRead, err := r.UnsafeReadBlob(int(emptyKeyOffset))
//	require.NoError(t, err)
//	require.Equal(t, emptyKey, emptyKeyRead)
//	emptyKeyRead2, err := r.UnsafeReadBlob(int(emptyKeyOffset2))
//	require.NoError(t, err)
//	require.Equal(t, emptyKey, emptyKeyRead2)
//	emptyValueRead, err := r.UnsafeReadBlob(int(emptyValueOffset))
//	require.NoError(t, err)
//	require.Equal(t, emptyValue, emptyValueRead)
//}
//
//func TestKVData_SizeLimits(t *testing.T) {
//	writer := openTestKVDataWriter(t)
//
//	// Test key at max size should succeed
//	maxKey := make([]byte, MaxKeySize)
//	_, err := writer.WriteKeyBlob(maxKey)
//	require.NoError(t, err)
//
//	// Test key exceeding max size should fail
//	oversizedKey := make([]byte, MaxKeySize+1)
//	_, err = writer.WriteKeyBlob(oversizedKey)
//	require.Error(t, err)
//	require.Contains(t, err.Error(), "key size exceeds maximum")
//
//	// Test value at max size should succeed
//	maxValue := make([]byte, MaxValueSize)
//	_, _, err = writer.WriteKeyValueBlobs([]byte("k"), maxValue)
//	require.NoError(t, err)
//
//	// Test value exceeding max size should fail
//	oversizedValue := make([]byte, MaxValueSize+1)
//	_, _, err = writer.WriteKeyValueBlobs([]byte("k"), oversizedValue)
//	require.Error(t, err)
//	require.Contains(t, err.Error(), "value size exceeds maximum")
//}
