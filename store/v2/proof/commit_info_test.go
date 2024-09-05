package proof

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetStoreProof(t *testing.T) {
	tests := []struct {
		storeInfos []StoreInfo
	}{
		{[]StoreInfo{
			{[]byte("key1"), CommitID{1, []byte("value1")}, "iavl"},
		}},
		{[]StoreInfo{
			{[]byte("key2"), CommitID{1, []byte("value2")}, "iavl"},
			{[]byte("key1"), CommitID{1, []byte("value1")}, "iavl"},
		}},
		{[]StoreInfo{
			{[]byte("key3"), CommitID{1, []byte("value3")}, "iavl"},
			{[]byte("key2"), CommitID{1, []byte("value2")}, "iavl"},
			{[]byte("key1"), CommitID{1, []byte("value1")}, "iavl"},
		}},
		{[]StoreInfo{
			{[]byte("key2"), CommitID{1, []byte("value2")}, "iavl"},
			{[]byte("key1"), CommitID{1, []byte("value1")}, "iavl"},
			{[]byte("key3"), CommitID{1, []byte("value3")}, "iavl"},
		}},
		{[]StoreInfo{
			{[]byte("key4"), CommitID{1, []byte("value4")}, "iavl"},
			{[]byte("key1"), CommitID{1, []byte("value1")}, "iavl"},
			{[]byte("key3"), CommitID{1, []byte("value3")}, "iavl"},
			{[]byte("key2"), CommitID{1, []byte("value2")}, "iavl"},
		}},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// create a commit info
			ci := CommitInfo{
				Version:    1,
				Timestamp:  time.Now(),
				StoreInfos: tc.storeInfos,
			}
			commitHash := ci.Hash()
			// make sure the store infos are sorted
			require.Equal(t, ci.StoreInfos[0].Name, []byte("key1"))
			for _, si := range tc.storeInfos {
				// get the proof
				_, proof, err := ci.GetStoreProof(si.Name)
				require.NoError(t, err, "test case %d", i)
				// verify the proof
				expRoots, err := proof.Run([][]byte{si.CommitID.Hash})
				require.NoError(t, err, "test case %d", i)
				require.Equal(t, commitHash, expRoots[0], "test case %d", i)

				bz, err := ci.Marshal()
				require.NoError(t, err)
				var ci2 CommitInfo
				err = ci2.Unmarshal(bz)
				require.NoError(t, err)
				require.True(t, ci.Timestamp.Equal(ci2.Timestamp))
				ci2.Timestamp = ci.Timestamp
				require.Equal(t, ci, ci2)
			}
		})
	}
}
