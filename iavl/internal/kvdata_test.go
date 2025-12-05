package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type kvDataWriterHelper struct {
	files *ChangesetFiles
	*KVDataWriter
}

func openTestKVDataWriter(t *testing.T) *kvDataWriterHelper {
	files, err := CreateChangesetFiles(t.TempDir(), 1, 0)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, files.Close())
	})
	writer := NewKVDataWriter(files.KVDataFile())
	return &kvDataWriterHelper{
		files:        files,
		KVDataWriter: writer,
	}
}

func (h *kvDataWriterHelper) openReader(t *testing.T) *KVDataReader {
	require.NoError(t, h.writer.Flush())
	rdr, err := NewKVDataReader(h.files.KVDataFile())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, rdr.Close())
	})
	return rdr
}

func TestKVData_WAL(t *testing.T) {
	writer := openTestKVDataWriter(t)

	// Write WAL start
	require.NoError(t, writer.WriteStartWAL(42))
	// Write WAL set with short key
	shortKey := []byte("key") // short key, should not be cached
	shortValue := []byte("value")
	shortKeyOffset, shortValueOffset, err := writer.WriteWALSet(shortKey, shortValue)
	require.NoError(t, err)
	// Write WAL set with longer key
	longerKey := []byte("longerKey") // longer key, should be cached
	longerValue := []byte("longerValue")
	longKeyOffset, longerValueOffset, err := writer.WriteWALSet(longerKey, longerValue)
	require.NoError(t, err)
	// Write WAL delete
	oldKey := []byte("oldKey")
	require.NoError(t, writer.WriteWALDelete(oldKey))
	// Write WAL commit
	require.NoError(t, writer.WriteWALCommit(42))

	// Write short key again to test caching
	shortValue2 := []byte("value2")
	shortKeyOffset2, shortValue2Offset, err := writer.WriteWALSet(shortKey, shortValue2)
	// short key should NOT be cached
	require.NotEqual(t, shortKeyOffset, shortKeyOffset2)
	// Write longer key again to test caching
	longerValue2 := []byte("longerValue2")
	longKeyOffset2, longerValue2Offset, err := writer.WriteWALSet(longerKey, longerValue2)
	// longer key should be cached
	require.Equal(t, longKeyOffset, longKeyOffset2)

	// Write WAL Updates
	memKey1 := []byte("memKey1")
	memValue1 := []byte("memValue1")
	memNode1 := &MemNode{
		key:   memKey1,
		value: memValue1,
	}
	memValue2 := []byte("memValue2")
	memNode2 := &MemNode{
		key:   longerKey, // should use cached key offset
		value: memValue2,
	}
	reinsertedValue := []byte("valueReinserted")
	memNode3 := &MemNode{
		key:   oldKey,
		value: reinsertedValue,
	}
	err = writer.WriteWALUpdates([]KVUpdate{
		{
			SetNode: memNode1,
		},
		{
			DeleteKey: oldKey,
		},
		{
			SetNode: memNode2,
		},
		{
			SetNode: memNode3,
		},
	})
	require.NoError(t, err)
	require.NotZero(t, memNode1.keyOffset)
	require.NotZero(t, memNode1.valueOffset)
	require.NotZero(t, memNode2.keyOffset)
	require.NotZero(t, memNode2.valueOffset)
	// memNode2 should use cached key offset
	require.Equal(t, longKeyOffset, memNode2.keyOffset)

	require.NoError(t, writer.WriteWALCommit(43))

	// test caching again with some blobs
	blobKeyOffset, err := writer.WriteKeyBlob([]byte("longerKey"))
	require.NoError(t, err)
	// should use cached offset
	require.Equal(t, longKeyOffset, blobKeyOffset)

	blobKeyOffset2, err := writer.WriteKeyBlob([]byte("key"))
	require.NoError(t, err)
	// short key should NOT be cached
	require.NotEqual(t, shortKeyOffset, blobKeyOffset2)

	// open reader
	r := writer.openReader(t)
	// Verify via WALReader
	wr, err := r.ReadWAL()
	require.NoError(t, err)
	require.Equal(t, uint64(42), wr.Version)
	// Entry 1: WAL Set short key
	entryType, ok, err := wr.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, KVEntryWALSet, entryType)
	require.Equal(t, []byte("key"), wr.Key)
	require.Equal(t, []byte("value"), wr.Value)

	// Entry 2: WAL Set longer key
	entryType, ok, err = wr.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, KVEntryWALSet, entryType)
	require.Equal(t, []byte("longerKey"), wr.Key)
	require.Equal(t, []byte("longerValue"), wr.Value)

	// Entry 3: WAL Delete
	entryType, ok, err = wr.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, KVEntryWALDelete, entryType)
	require.Equal(t, []byte("oldKey"), wr.Key)

	// Entry 4: WAL Commit
	entryType, ok, err = wr.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, KVEntryWALCommit, entryType)
	require.Equal(t, uint64(42), wr.Version)

	// Entry 5: WAL Set short key again (not cached)
	entryType, ok, err = wr.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, KVEntryWALSet, entryType)
	require.Equal(t, []byte("key"), wr.Key)
	require.Equal(t, []byte("value2"), wr.Value)

	// Entry 6: WAL Set longer key again (cached)
	entryType, ok, err = wr.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, KVEntryWALSet|KVFlagCachedKey, entryType)
	require.Equal(t, []byte("longerKey"), wr.Key)
	require.Equal(t, []byte("longerValue2"), wr.Value)

	// Entry 7-10: WAL Updates
	for i, expected := range []struct {
		entryType KVEntryType
		key       []byte
		value     []byte
	}{
		{KVEntryWALSet, memKey1, memValue1},
		{KVEntryWALDelete | KVFlagCachedKey, oldKey, nil},
		{KVEntryWALSet | KVFlagCachedKey, longerKey, memValue2},
		{KVEntryWALSet | KVFlagCachedKey, oldKey, reinsertedValue},
	} {
		entryType, ok, err = wr.Next()
		require.NoError(t, err, "WAL Update entry %d", i)
		require.True(t, ok, "WAL Update entry %d", i)
		require.Equal(t, expected.entryType, entryType, "WAL Update entry %d type", i)
		require.Equal(t, expected.key, wr.Key, "WAL Update entry %d key", i)
		if expected.value != nil {
			require.Equal(t, expected.value, wr.Value, "WAL Update entry %d value", i)
		}
	}
	// Entry 11: WAL Commit
	entryType, ok, err = wr.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, KVEntryWALCommit, entryType)
	require.Equal(t, uint64(43), wr.Version)

	// No more entries
	_, ok, err = wr.Next()
	require.NoError(t, err)
	require.False(t, ok)

	// Check that all offsets are readable
	shortKeyRead, err := r.UnsafeReadBlob(int(shortKeyOffset))
	require.NoError(t, err)
	require.Equal(t, shortKey, shortKeyRead)
	shortValueRead, err := r.UnsafeReadBlob(int(shortValueOffset))
	require.NoError(t, err)
	require.Equal(t, shortValue, shortValueRead)
	longKeyRead, err := r.UnsafeReadBlob(int(longKeyOffset))
	require.NoError(t, err)
	require.Equal(t, longerKey, longKeyRead)
	longValueRead, err := r.UnsafeReadBlob(int(longerValueOffset))
	require.NoError(t, err)
	require.Equal(t, longerValue, longValueRead)
	shortKeyRead2, err := r.UnsafeReadBlob(int(shortKeyOffset2))
	require.NoError(t, err)
	require.Equal(t, shortKey, shortKeyRead2)
	longValueRead2, err := r.UnsafeReadBlob(int(longerValue2Offset))
	require.NoError(t, err)
	require.Equal(t, longerValue2, longValueRead2)
	shorterValue2Read, err := r.UnsafeReadBlob(int(shortValue2Offset))
	require.NoError(t, err)
	require.Equal(t, shortValue2, shorterValue2Read)
	// also check all memNode offsets
	memKey1Read, err := r.UnsafeReadBlob(int(memNode1.keyOffset))
	require.NoError(t, err)
	require.Equal(t, memKey1, memKey1Read)
	memValue1Read, err := r.UnsafeReadBlob(int(memNode1.valueOffset))
	require.NoError(t, err)
	require.Equal(t, memValue1, memValue1Read)
	memKey2Read, err := r.UnsafeReadBlob(int(memNode2.keyOffset))
	require.NoError(t, err)
	require.Equal(t, longerKey, memKey2Read)
	memValue2Read, err := r.UnsafeReadBlob(int(memNode2.valueOffset))
	require.NoError(t, err)
	require.Equal(t, memValue2, memValue2Read)
	memKey3Read, err := r.UnsafeReadBlob(int(memNode3.keyOffset))
	require.NoError(t, err)
	require.Equal(t, oldKey, memKey3Read)
	memValue3Read, err := r.UnsafeReadBlob(int(memNode3.valueOffset))
	require.NoError(t, err)
	require.Equal(t, reinsertedValue, memValue3Read)
}

func TestKVData_WALStart(t *testing.T) {
	tests := []struct {
		name    string
		version uint64
	}{
		{"version 0", 0},
		{"version 1", 1},
		{"version 100", 100},
		{"large version", 1<<32 + 12345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := openTestKVDataWriter(t)

			// Write WAL start
			require.NoError(t, writer.WriteStartWAL(tt.version))

			r := writer.openReader(t)

			// Verify HasWAL
			hasWAL, startVersion := r.HasWAL()
			require.True(t, hasWAL)
			require.Equal(t, tt.version, startVersion)

			// Verify ReadWAL
			wr, err := r.ReadWAL()
			require.NoError(t, err)
			require.Equal(t, tt.version, wr.Version)
		})
	}
}

//func TestKVData_WALMode(t *testing.T) {
//	tests := []struct {
//		name           string
//		startInWalMode bool
//		op             func(t *testing.T, writer *KVDataWriter)
//	}{
//		{
//			name:           "WAL",
//			startInWalMode: true,
//			op: func(t *testing.T, writer *KVDataWriter) {
//
//			},
//		},
//	}
//}

/*func TestKVData_WALStart_EmptyFile(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

// Create empty file
f, err := os.Create(path)
require.NoError(t, err)
require.NoError(t, f.Close())

// Read phase
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

// HasWAL should return false for empty file
hasWAL, startVersion := r.HasWAL()
require.False(t, hasWAL)
require.Equal(t, uint64(0), startVersion)

// ReadWAL should return error
_, err = r.ReadWAL()
require.Error(t, err)
}

func TestKVData_WALSet(t *testing.T) {
tests := []struct {
name  string
key   []byte
value []byte
}{
{"simple kv", []byte("hello"), []byte("world")},
{"empty value", []byte("key"), []byte{}},
{"binary key", []byte{0x00, 0x01, 0x02}, []byte("value")},
{"medium key", []byte(strings.Repeat("k", 100)), []byte(strings.Repeat("v", 200))},
}

for _, tt := range tests {
t.Run(tt.name, func (t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

// Write phase
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
err = w.WriteStartWAL(1)
require.NoError(t, err)
keyOffset, valueOffset, err := w.WriteWALSet(tt.key, tt.value)
require.NoError(t, err)
require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read phase
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

// Verify via WALReader
wr, err := r.ReadWAL()
require.NoError(t, err)
entryType, ok, err := wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALSet, entryType)
require.Equal(t, tt.key, wr.Key)
require.Equal(t, tt.value, wr.Value)

// Verify offsets via UnsafeReadBlob
keyRead, err := r.UnsafeReadBlob(int(keyOffset))
require.NoError(t, err)
require.Equal(t, tt.key, keyRead)

valueRead, err := r.UnsafeReadBlob(int(valueOffset))
require.NoError(t, err)
require.Equal(t, tt.value, valueRead)

// No more entries
_, ok, err = wr.Next()
require.NoError(t, err)
require.False(t, ok)
})
}
}

func TestKVData_WALDelete(t *testing.T) {
tests := []struct {
name string
key  []byte
}{
{"simple key", []byte("deleteMe")},
{"binary key", []byte{0xFF, 0xFE, 0xFD}},
{"medium key", []byte(strings.Repeat("d", 50))},
}

for _, tt := range tests {
t.Run(tt.name, func (t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

// Write phase
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
err = w.WriteStartWAL(1)
require.NoError(t, err)
err = w.WriteWALDelete(tt.key)
require.NoError(t, err)
require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read phase
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

wr, err := r.ReadWAL()
require.NoError(t, err)
entryType, ok, err := wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALDelete, entryType)
require.Equal(t, tt.key, wr.Key)

// No more entries
_, ok, err = wr.Next()
require.NoError(t, err)
require.False(t, ok)
})
}
}

func TestKVData_WALCommit(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

// Write phase
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
err = w.WriteStartWAL(1)
require.NoError(t, err)
_, _, err = w.WriteWALSet([]byte("key1"), []byte("val1"))
require.NoError(t, err)
err = w.WriteWALCommit(1)
require.NoError(t, err)
_, _, err = w.WriteWALSet([]byte("key2"), []byte("val2"))
require.NoError(t, err)
err = w.WriteWALCommit(2)
require.NoError(t, err)
require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read phase
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

wr, err := r.ReadWAL()
require.NoError(t, err)
require.Equal(t, uint64(1), wr.Version) // Start version

// Entry 1: Set
entryType, ok, err := wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALSet, entryType)
require.Equal(t, []byte("key1"), wr.Key)

// Entry 2: Commit - version should update
entryType, ok, err = wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALCommit, entryType)
require.Equal(t, uint64(1), wr.Version)

// Entry 3: Set
entryType, ok, err = wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALSet, entryType)
require.Equal(t, []byte("key2"), wr.Key)

// Entry 4: Commit
entryType, ok, err = wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALCommit, entryType)
require.Equal(t, uint64(2), wr.Version)

// No more entries
_, ok, err = wr.Next()
require.NoError(t, err)
require.False(t, ok)
}

func TestKVData_KeyCaching(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

key := []byte("cachedKey") // > 4 bytes, will be cached

// Write phase
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
err = w.WriteStartWAL(1)
require.NoError(t, err)

// First write - should be inline
keyOffset1, _, err := w.WriteWALSet(key, []byte("value1"))
require.NoError(t, err)

// Second write of same key - should use cached offset
keyOffset2, _, err := w.WriteWALSet(key, []byte("value2"))
require.NoError(t, err)

// Both should have the same key offset
require.Equal(t, keyOffset1, keyOffset2)

require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read raw bytes to verify cached flag is set
data, err := os.ReadFile(path)
require.NoError(t, err)

// Find the second WALSet entry - it should have the cached flag (0x81)
// First entry is at offset after WALStart
foundCached := false
for i := 0; i < len(data); i++ {
if data[i] == byte(KVEntryWALSet|KVFlagCachedKey) {
foundCached = true
break
}
}
require.True(t, foundCached, "second WALSet should have cached key flag")

// Read and verify both entries work correctly
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

wr, err := r.ReadWAL()
require.NoError(t, err)

// First entry
entryType, ok, err := wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALSet, entryType)
require.Equal(t, key, wr.Key)
require.Equal(t, []byte("value1"), wr.Value)

// Second entry - should also resolve key correctly via cache
entryType, ok, err = wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALSet|KVFlagCachedKey, entryType)
require.Equal(t, key, wr.Key)
require.Equal(t, []byte("value2"), wr.Value)
}

func TestKVData_KeyCaching_ShortKeys(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

shortKey := []byte("abc") // 3 bytes, should NOT be cached

// Write phase
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
err = w.WriteStartWAL(1)
require.NoError(t, err)

// First write
_, _, err = w.WriteWALSet(shortKey, []byte("value1"))
require.NoError(t, err)

// Second write - should NOT use cache (key too short)
_, _, err = w.WriteWALSet(shortKey, []byte("value2"))
require.NoError(t, err)

require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read raw bytes - should NOT find cached flag for short keys
data, err := os.ReadFile(path)
require.NoError(t, err)

cachedCount := 0
for i := 0; i < len(data); i++ {
if data[i] == byte(KVEntryWALSet|KVFlagCachedKey) {
cachedCount++
}
}
require.Equal(t, 0, cachedCount, "short keys should not be cached")
}

func TestKVData_WALReplay(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

// Write a realistic WAL sequence
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
err = w.WriteStartWAL(10)
require.NoError(t, err)

// Version 10 operations
_, _, err = w.WriteWALSet([]byte("key1"), []byte("val1"))
require.NoError(t, err)
_, _, err = w.WriteWALSet([]byte("key2"), []byte("val2"))
require.NoError(t, err)
err = w.WriteWALDelete([]byte("oldKey"))
require.NoError(t, err)
err = w.WriteWALCommit(10)
require.NoError(t, err)

// Version 11 operations
_, _, err = w.WriteWALSet([]byte("key1"), []byte("val1_updated"))
require.NoError(t, err)
_, _, err = w.WriteWALSet([]byte("key3"), []byte("val3"))
require.NoError(t, err)
err = w.WriteWALCommit(11)
require.NoError(t, err)

require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Replay
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

wr, err := r.ReadWAL()
require.NoError(t, err)
require.Equal(t, uint64(10), wr.Version)

expectedEntries := []struct {
entryType KVEntryType
key       []byte
value     []byte
version   uint64
}{
{KVEntryWALSet, []byte("key1"), []byte("val1"), 10},
{KVEntryWALSet, []byte("key2"), []byte("val2"), 10},
{KVEntryWALDelete, []byte("oldKey"), nil, 10},
{KVEntryWALCommit, nil, nil, 10},
{KVEntryWALSet | KVFlagCachedKey, []byte("key1"), []byte("val1_updated"), 10}, // key1 cached
{KVEntryWALSet, []byte("key3"), []byte("val3"), 10},
{KVEntryWALCommit, nil, nil, 11},
}

for i, exp := range expectedEntries {
entryType, ok, err := wr.Next()
require.NoError(t, err, "entry %d", i)
require.True(t, ok, "entry %d", i)
require.Equal(t, exp.entryType, entryType, "entry %d type", i)

if exp.key != nil {
require.Equal(t, exp.key, wr.Key, "entry %d key", i)
}
if exp.value != nil {
require.Equal(t, exp.value, wr.Value, "entry %d value", i)
}
if entryType == KVEntryWALCommit {
require.Equal(t, exp.version, wr.Version, "entry %d version", i)
}
}

// No more entries
_, ok, err := wr.Next()
require.NoError(t, err)
require.False(t, ok)
}

func TestKVData_Blobs(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

key := []byte("blobKey")
value := []byte("blobValue")

// Write phase
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)

// Write key blob
keyOffset, err := w.WriteKeyBlob(key)
require.NoError(t, err)

// Write same key again - should return cached offset
keyOffset2, err := w.WriteKeyBlob(key)
require.NoError(t, err)
require.Equal(t, keyOffset, keyOffset2)

// Write key/value blobs
keyOffset3, valueOffset, err := w.WriteKeyValueBlobs(key, value)
require.NoError(t, err)
require.Equal(t, keyOffset, keyOffset3) // same key, cached

require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read phase
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

// Read key via offset
keyRead, err := r.UnsafeReadBlob(int(keyOffset))
require.NoError(t, err)
require.Equal(t, key, keyRead)

// Read value via offset
valueRead, err := r.UnsafeReadBlob(int(valueOffset))
require.NoError(t, err)
require.Equal(t, value, valueRead)
}

func TestKVData_Blobs_ValueType(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

// Write phase
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
_, valueOffset, err := w.WriteKeyValueBlobs([]byte("key"), []byte("value"))
require.NoError(t, err)
require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read raw bytes to verify value uses KVEntryValueBlob type
data, err := os.ReadFile(path)
require.NoError(t, err)

// The byte before valueOffset should be the type byte
// valueOffset points to the varint length, type byte is 1 before
require.Equal(t, byte(KVEntryValueBlob), data[valueOffset-1])
}

func TestKVData_MixedEntries(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

// Write WAL with blob entries interspersed
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
err = w.WriteStartWAL(1)
require.NoError(t, err)

// WAL set
_, _, err = w.WriteWALSet([]byte("walKey1"), []byte("walVal1"))
require.NoError(t, err)

// Blob entry (non-WAL)
blobOffset, err := w.WriteKeyBlob([]byte("blobKey"))
require.NoError(t, err)

// Another WAL set
_, _, err = w.WriteWALSet([]byte("walKey2"), []byte("walVal2"))
require.NoError(t, err)

err = w.WriteWALCommit(1)
require.NoError(t, err)

require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read phase
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

wr, err := r.ReadWAL()
require.NoError(t, err)

// WALReader.Next() should skip blob entries
// Entry 1: WAL Set
entryType, ok, err := wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALSet, entryType)
require.Equal(t, []byte("walKey1"), wr.Key)

// Entry 2: WAL Set (blob skipped)
entryType, ok, err = wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALSet, entryType)
require.Equal(t, []byte("walKey2"), wr.Key)

// Entry 3: Commit
entryType, ok, err = wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALCommit, entryType)

// Blob is still readable via offset
blobData, err := r.UnsafeReadBlob(int(blobOffset))
require.NoError(t, err)
require.Equal(t, []byte("blobKey"), blobData)
}

func TestKVData_EdgeCases_LargeKeys(t *testing.T) {
tests := []struct {
name    string
keySize int
}{
{"127 bytes (1-byte varint)", 127},
{"128 bytes (2-byte varint)", 128},
{"16383 bytes (2-byte varint max)", 16383},
{"16384 bytes (3-byte varint)", 16384},
}

for _, tt := range tests {
t.Run(tt.name, func (t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

key := make([]byte, tt.keySize)
for i := range key {
key[i] = byte(i % 256)
}
value := []byte("value")

// Write
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
err = w.WriteStartWAL(1)
require.NoError(t, err)
keyOffset, _, err := w.WriteWALSet(key, value)
require.NoError(t, err)
require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

// Via WALReader
wr, err := r.ReadWAL()
require.NoError(t, err)
entryType, ok, err := wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALSet, entryType)
require.Equal(t, key, wr.Key)
require.Equal(t, value, wr.Value)

// Via direct blob read
keyRead, err := r.UnsafeReadBlob(int(keyOffset))
require.NoError(t, err)
require.Equal(t, key, keyRead)
})
}
}

func TestKVData_EdgeCases_EmptyKey(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "kv.dat")

// Write
f, err := os.Create(path)
require.NoError(t, err)
w := NewKVDataWriter(f)
err = w.WriteStartWAL(1)
require.NoError(t, err)
_, _, err = w.WriteWALSet([]byte{}, []byte("value"))
require.NoError(t, err)
require.NoError(t, w.Flush())
require.NoError(t, f.Close())

// Read
f, err = os.Open(path)
require.NoError(t, err)
defer f.Close()
r, err := NewKVDataReader(f)
require.NoError(t, err)

wr, err := r.ReadWAL()
require.NoError(t, err)
entryType, ok, err := wr.Next()
require.NoError(t, err)
require.True(t, ok)
require.Equal(t, KVEntryWALSet, entryType)
require.Equal(t, []byte{}, wr.Key)
require.Equal(t, []byte("value"), wr.Value)
}
*/
