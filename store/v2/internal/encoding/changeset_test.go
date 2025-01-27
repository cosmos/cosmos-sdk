package encoding

import (
	"testing"

	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
)

func TestChangesetMarshal(t *testing.T) {
	testcases := []struct {
		name         string
		changeset    *corestore.Changeset
		encodedSize  int
		encodedBytes []byte
	}{
		{
			name:         "empty",
			changeset:    corestore.NewChangeset(1),
			encodedSize:  1,
			encodedBytes: []byte{0x0},
		},
		{
			name: "one store",
			changeset: &corestore.Changeset{Changes: []corestore.StateChanges{
				{
					Actor: []byte("storekey"),
					StateChanges: corestore.KVPairs{
						{Key: []byte("key"), Value: []byte("value"), Remove: false},
					},
				},
			}},
			encodedSize:  1 + 1 + 8 + 1 + 1 + 3 + 1 + 1 + 5,
			encodedBytes: []byte{0x1, 0x8, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x6b, 0x65, 0x79, 0x1, 0x3, 0x6b, 0x65, 0x79, 0x0, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65},
		},
		{
			name: "one remove store",
			changeset: &corestore.Changeset{Changes: []corestore.StateChanges{
				{
					Actor: []byte("storekey"),
					StateChanges: corestore.KVPairs{
						{Key: []byte("key"), Remove: true},
					},
				},
			}},
			encodedSize:  1 + 1 + 8 + 1 + 1 + 3 + 1,
			encodedBytes: []byte{0x1, 0x8, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x6b, 0x65, 0x79, 0x1, 0x3, 0x6b, 0x65, 0x79, 0x1},
		},
		{
			name: "two stores",
			changeset: &corestore.Changeset{Changes: []corestore.StateChanges{
				{
					Actor: []byte("storekey1"),
					StateChanges: corestore.KVPairs{
						{Key: []byte("key1"), Value: []byte("value1"), Remove: false},
					},
				},
				{
					Actor: []byte("storekey2"),
					StateChanges: corestore.KVPairs{
						{Key: []byte("key2"), Value: []byte("value2"), Remove: false},
						{Key: []byte("key1"), Remove: true},
					},
				},
			}},
			encodedSize: 2 + 1 + 9 + 1 + 1 + 4 + 1 + 6 + 1 + 9 + 1 + 1 + 4 + 1 + 1 + 6 + 1 + 4 + 1,
			// encodedBytes: it is not deterministic,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// check the encoded size
			require.Equal(t, encodedSize(tc.changeset), tc.encodedSize, "encoded size mismatch")
			// check the encoded bytes
			encodedBytes, err := MarshalChangeset(tc.changeset)
			require.NoError(t, err, "marshal error")
			if len(tc.encodedBytes) != 0 {
				require.Equal(t, encodedBytes, tc.encodedBytes, "encoded bytes mismatch")
			}
			// check the unmarshaled changeset
			cs := corestore.NewChangeset(1)
			require.NoError(t, UnmarshalChangeset(cs, encodedBytes), "unmarshal error")
			require.Equal(t, len(tc.changeset.Changes), len(cs.Changes), "unmarshaled changeset store size mismatch")
			for i, changes := range tc.changeset.Changes {
				require.Equal(t, changes.Actor, cs.Changes[i].Actor, "unmarshaled changeset store key mismatch")
				require.Equal(t, len(changes.StateChanges), len(cs.Changes[i].StateChanges), "unmarshaled changeset StateChanges size mismatch")
				for j, pair := range changes.StateChanges {
					require.Equal(t, pair, cs.Changes[i].StateChanges[j], "unmarshaled changeset pair mismatch")
				}
			}
		})
	}
}
