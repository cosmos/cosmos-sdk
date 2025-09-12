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
