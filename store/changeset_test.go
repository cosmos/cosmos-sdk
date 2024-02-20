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
			name:         "two stores",
			changeset:    &Changeset{Pairs: map[string]KVPairs{"storekey1": {{Key: []byte("key1"), Value: []byte("value1"), StoreKey: "storekey1"}}, "storekey2": {{Key: []byte("key2"), Value: []byte("value2"), StoreKey: "storekey2"}, {Key: []byte("key3"), Value: []byte("value3"), StoreKey: "storekey2"}}}},
			encodedSize:  1 + 1 + 9 + 1 + 1 + 4 + 1 + 6 + 1 + 9 + 1 + 1 + 4 + 1 + 6 + 1 + 4 + 1 + 6,
			encodedBytes: []byte{0x2, 0x9, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x6b, 0x65, 0x79, 0x31, 0x1, 0x4, 0x6b, 0x65, 0x79, 0x31, 0x6, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, 0x9, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x6b, 0x65, 0x79, 0x32, 0x2, 0x4, 0x6b, 0x65, 0x79, 0x32, 0x6, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0x4, 0x6b, 0x65, 0x79, 0x33, 0x6, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x33},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.changeset.encodedSize(), tc.encodedSize, "encoded size mismatch")
			encodedBytes, err := tc.changeset.Marshal()
			require.NoError(t, err, "marshal error")
			require.Equal(t, encodedBytes, tc.encodedBytes, "encoded bytes mismatch")
			cs := NewChangeset()
			require.NoError(t, cs.Unmarshal(tc.encodedBytes), "unmarshal error")
			require.Equal(t, cs, tc.changeset, "unmarshaled changeset mismatch")
		})
	}
}
