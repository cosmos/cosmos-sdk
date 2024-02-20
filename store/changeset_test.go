package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChangesetMarshal(t *testing.T) {
	testcases := []struct {
		name         string
		changeset    *Changeset
		encodedSize  int
		encodedBytes []byte
	}{
		{
			name:         "empty",
			changeset:    NewChangeset(),
			encodedSize:  1,
			encodedBytes: []byte{0x0},
		},
		{
			name:         "one store",
			changeset:    &Changeset{Pairs: map[string]KVPairs{"storekey": {{Key: []byte("key"), Value: []byte("value"), StoreKey: "storekey"}}}},
			encodedSize:  1 + 1 + 8 + 1 + 1 + 3 + 1 + 5,
			encodedBytes: []byte{0x1, 0x8, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x6b, 0x65, 0x79, 0x1, 0x3, 0x6b, 0x65, 0x79, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65},
		},
		{
			name:        "two stores",
			changeset:   &Changeset{Pairs: map[string]KVPairs{"storekey1": {{Key: []byte("key1"), Value: []byte("value1"), StoreKey: "storekey1"}}, "storekey2": {{Key: []byte("key2"), Value: []byte("value2"), StoreKey: "storekey2"}, {Key: []byte("key3"), Value: []byte("value3"), StoreKey: "storekey2"}}}},
			encodedSize: 1 + 1 + 9 + 1 + 1 + 4 + 1 + 6 + 1 + 9 + 1 + 1 + 4 + 1 + 6 + 1 + 4 + 1 + 6,
			// encodedBytes: it is not deterministic,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// check the encoded size
			require.Equal(t, tc.changeset.encodedSize(), tc.encodedSize, "encoded size mismatch")
			// check the encoded bytes
			encodedBytes, err := tc.changeset.Marshal()
			require.NoError(t, err, "marshal error")
			if len(tc.encodedBytes) != 0 {
				require.Equal(t, encodedBytes, tc.encodedBytes, "encoded bytes mismatch")
			}
			// check the unmarshaled changeset
			cs := NewChangeset()
			require.NoError(t, cs.Unmarshal(encodedBytes), "unmarshal error")
			require.Equal(t, len(tc.changeset.Pairs), len(cs.Pairs), "unmarshaled changeset store size mismatch")
			for storeKey, pairs := range tc.changeset.Pairs {
				require.Equal(t, len(pairs), len(cs.Pairs[storeKey]), "unmarshaled changeset pairs size mismatch")
				for i, pair := range pairs {
					require.Equal(t, pair, cs.Pairs[storeKey][i], "unmarshaled changeset pair mismatch")
				}
			}
		})
	}
}
