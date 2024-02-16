package proof

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProofFromBytesSlices(t *testing.T) {
	tests := []struct {
		keys   []string
		values []string
		want   string
	}{
		{[]string{}, []string{}, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
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
			"55d4bce1c53b7d394bd41bbfc2b239cc2e1c7e36423612a97181c47e79bb713c",
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
			"443382fbb629e0d50e86d6ea49e22aa4e27ba50262730b0122cec36860c903a2",
		},
	}
	for i, tc := range tests {
		var err error
		leaves := make([][]byte, len(tc.keys))
		for j, key := range tc.keys {
			leaves[j], err = LeafHash([]byte(key), []byte(tc.values[j]))
			require.NoError(t, err)
		}
		for j := range leaves {
			buf := make([][]byte, len(leaves))
			copy(buf, leaves)
			rootHash, inners := ProofFromByteSlices(buf, j)
			require.Equal(t, tc.want, fmt.Sprintf("%x", rootHash), "test case %d", i)
			commitmentOp := ConvertCommitmentOp(inners, []byte(tc.keys[j]), []byte(tc.values[j]))
			expRoots, err := commitmentOp.Run([][]byte{[]byte(tc.values[j])})
			require.NoError(t, err)
			require.Equal(t, tc.want, fmt.Sprintf("%x", expRoots[0]), "test case %d", i)
		}

	}
}
