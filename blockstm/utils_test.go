package blockstm

import (
	"testing"

	"github.com/test-go/testify/require"
)

type DiffEntry struct {
	Key   Key
	IsNew bool
}

func TestDiffOrderedList(t *testing.T) {
	testCases := []struct {
		name     string
		old      []Key
		new      []Key
		expected []DiffEntry
	}{
		{
			name:     "empty lists",
			old:      []Key{},
			new:      []Key{},
			expected: []DiffEntry{},
		},
		{
			name: "old is longer",
			old: []Key{
				[]byte("a"),
				[]byte("b"),
				[]byte("c"),
				[]byte("d"),
				[]byte("e"),
			},
			new: []Key{
				[]byte("b"),
				[]byte("c"),
				[]byte("f"),
			},
			expected: []DiffEntry{
				{Key: []byte("a"), IsNew: false},
				{Key: []byte("d"), IsNew: false},
				{Key: []byte("e"), IsNew: false},
				{Key: []byte("f"), IsNew: true},
			},
		},
		{
			name: "new is longer",
			old: []Key{
				[]byte("a"),
				[]byte("c"),
				[]byte("e"),
			},
			new: []Key{
				[]byte("b"),
				[]byte("c"),
				[]byte("d"),
				[]byte("e"),
				[]byte("f"),
			},
			expected: []DiffEntry{
				{Key: []byte("a"), IsNew: false},
				{Key: []byte("b"), IsNew: true},
				{Key: []byte("d"), IsNew: true},
				{Key: []byte("f"), IsNew: true},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := []DiffEntry{}
			DiffOrderedList(tc.old, tc.new, func(key Key, leftOrRight bool) bool {
				result = append(result, DiffEntry{key, leftOrRight})
				return true
			})
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestPackUnpackValidationIdx(t *testing.T) {
	testCases := []struct {
		index TxnIndex
		wave  Wave
	}{
		{0, 0},
		{1, 0},
		{0, 1},
		{12345, 67890},
		{4294967295, 4294967295}, // max uint32
	}

	for _, tc := range testCases {
		packed := PackValidationIdx(tc.index, tc.wave)
		unpackedIndex, unpackedIncarnation := UnpackValidationIdx(packed)
		require.Equal(t, tc.index, unpackedIndex)
		require.Equal(t, tc.wave, unpackedIncarnation)
	}
}

func TestArmedLock(t *testing.T) {
	var lock ArmedLock
	lock.Init()

	// TryLock should succeed initially
	require.True(t, lock.TryLock())

	// TryLock should fail when already locked
	require.False(t, lock.TryLock())

	// Unlock the lock
	lock.Unlock()

	// TryLock should fail when no work
	require.False(t, lock.TryLock())

	// add work
	lock.Arm()

	// TryLock should succeed since it's armed
	require.True(t, lock.TryLock())

	// add work
	lock.Arm()

	// TryLock should fail when already locked
	require.False(t, lock.TryLock())
}
