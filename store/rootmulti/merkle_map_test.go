package rootmulti

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleMap(t *testing.T) {
	tests := []struct {
		keys   []string
		values []string // each string gets converted to []byte in test
		want   string
	}{
		{[]string{"key1"}, []string{"value1"}, "a44d3cc7daba1a4600b00a2434b30f8b970652169810d6dfa9fb1793a2189324"},
		{[]string{"key1"}, []string{"value2"}, "0638e99b3445caec9d95c05e1a3fc1487b4ddec6a952ff337080360b0dcc078c"},
		// swap order with 2 keys
		{
			[]string{"key1", "key2"},
			[]string{"value1", "value2"},
			"8fd19b19e7bb3f2b3ee0574027d8a5a4cec370464ea2db2fbfa5c7d35bb0cff3",
		},
		{
			[]string{"key2", "key1"},
			[]string{"value2", "value1"},
			"8fd19b19e7bb3f2b3ee0574027d8a5a4cec370464ea2db2fbfa5c7d35bb0cff3",
		},
		// swap order with 3 keys
		{
			[]string{"key1", "key2", "key3"},
			[]string{"value1", "value2", "value3"},
			"1dd674ec6782a0d586a903c9c63326a41cbe56b3bba33ed6ff5b527af6efb3dc",
		},
		{
			[]string{"key1", "key3", "key2"},
			[]string{"value1", "value3", "value2"},
			"1dd674ec6782a0d586a903c9c63326a41cbe56b3bba33ed6ff5b527af6efb3dc",
		},
	}
	for i, tc := range tests {
		db := newMerkleMap()
		for i := 0; i < len(tc.keys); i++ {
			db.set(tc.keys[i], []byte(tc.values[i]))
		}

		got := db.hash()
		assert.Equal(t, tc.want, fmt.Sprintf("%x", got), "Hash didn't match on tc %d", i)
	}
}
