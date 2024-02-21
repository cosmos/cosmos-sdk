package proof

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetStoreProof(t *testing.T) {
	tests := []struct {
		storeInfos []StoreInfo
	}{
		{[]StoreInfo{
			{"key1", CommitID{1, []byte("value1")}},
		}},
		{[]StoreInfo{
			{"key2", CommitID{1, []byte("value2")}},
			{"key1", CommitID{1, []byte("value1")}},
		}},
		{[]StoreInfo{
			{"key3", CommitID{1, []byte("value3")}},
			{"key2", CommitID{1, []byte("value2")}},
			{"key1", CommitID{1, []byte("value1")}},
		}},
		{[]StoreInfo{
			{"key2", CommitID{1, []byte("value2")}},
			{"key1", CommitID{1, []byte("value1")}},
			{"key3", CommitID{1, []byte("value3")}},
		}},
		{[]StoreInfo{
			{"key4", CommitID{1, []byte("value4")}},
			{"key1", CommitID{1, []byte("value1")}},
			{"key3", CommitID{1, []byte("value3")}},
			{"key2", CommitID{1, []byte("value2")}},
		}},
	}

	for i, tc := range tests {
		// create a commit info
		ci := CommitInfo{
			Version:    1,
			Timestamp:  time.Now(),
			StoreInfos: tc.storeInfos,
		}
		commitHash := ci.Hash()
		// make sure the store infos are sorted
		require.Equal(t, ci.StoreInfos[0].Name, "key1")
		for _, si := range tc.storeInfos {
			// get the proof
			_, proof, err := ci.GetStoreProof(si.Name)
			require.NoError(t, err, "test case %d", i)
			// verify the proof
			expRoots, err := proof.Run([][]byte{si.CommitID.Hash})
			require.NoError(t, err, "test case %d", i)
			require.Equal(t, commitHash, expRoots[0], "test case %d", i)
		}
	}
}
