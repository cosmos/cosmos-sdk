package store

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
			"ad2b0c23dbd3376440a5347fba02ff35cfad7930daa5e733930315b6fbb03b26",
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
			"92cd50420c22d0c79f64dd1b04bfd5f5d73265f7ac37e65cf622f3cf8b963805",
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
