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
		{[]string{"key1"}, []string{"value1"}, "09c468a07fe9bc1f14e754cff0acbad4faf9449449288be8e1d5d1199a247034"},
		{[]string{"key1"}, []string{"value2"}, "2131d85de3a8ded5d3a72bfc657f7324138540c520de7401ac8594785a3082fb"},
		// swap order with 2 keys
		{
			[]string{"key1", "key2"},
			[]string{"value1", "value2"},
			"017788f37362dd0687beb59c0b3bfcc17a955120a4cb63dbdd4a0fdf9e07730e",
		},
		{
			[]string{"key2", "key1"},
			[]string{"value2", "value1"},
			"017788f37362dd0687beb59c0b3bfcc17a955120a4cb63dbdd4a0fdf9e07730e",
		},
		// swap order with 3 keys
		{
			[]string{"key1", "key2", "key3"},
			[]string{"value1", "value2", "value3"},
			"68f41a8a3508cb5f8eb3f1c7534a86fea9f59aa4898a5aac2f1bb92834ae2a36",
		},
		{
			[]string{"key1", "key3", "key2"},
			[]string{"value1", "value3", "value2"},
			"68f41a8a3508cb5f8eb3f1c7534a86fea9f59aa4898a5aac2f1bb92834ae2a36",
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
