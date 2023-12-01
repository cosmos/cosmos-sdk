package commitment

import (
	"testing"
)

func TestSnapshotter(t *testing.T) {
	// generate a new tree
	// storeKey := "store"
	// tree := generateTree(storeKey)
	// require.NotNil(t, tree)

	// latestVersion := uint64(10)
	// kvCount := 10
	// for i := uint64(1); i <= latestVersion; i++ {
	// 	cs := store.NewChangeset()
	// 	for j := 0; j < kvCount; j++ {
	// 		key := []byte(fmt.Sprintf("key-%d-%d", i, j))
	// 		value := []byte(fmt.Sprintf("value-%d-%d", i, j))
	// 		cs.Add(key, value)
	// 	}
	// 	err := tree.WriteBatch(cs)
	// 	require.NoError(t, err)

	// 	_, err = tree.Commit()
	// 	require.NoError(t, err)
	// }

	// latestHash := tree.WorkingHash()

	// // create a snapshot
	// dummyExtensionItem := snapshottypes.SnapshotItem{
	// 	Item: &snapshottypes.SnapshotItem_Extension{
	// 		Extension: &snapshottypes.SnapshotExtensionMeta{
	// 			Name:   "test",
	// 			Format: 1,
	// 		},
	// 	},
	// }
	// target := generateTree("")
	// chunks := make(chan io.ReadCloser, kvCount*int(latestVersion))
	// go func() {
	// 	streamWriter := snapshots.NewStreamWriter(chunks)
	// 	require.NotNil(t, streamWriter)
	// 	defer streamWriter.Close()
	// 	err := tree.Snapshot(latestVersion, streamWriter)
	// 	require.NoError(t, err)
	// 	// write an extension metadata
	// 	err = streamWriter.WriteMsg(&dummyExtensionItem)
	// 	require.NoError(t, err)
	// }()

	// streamReader, err := snapshots.NewStreamReader(chunks)
	// chStorage := make(chan *store.KVPair, 100)
	// require.NoError(t, err)
	// nextItem, err := target.Restore(latestVersion, snapshottypes.CurrentFormat, streamReader, chStorage)
	// require.NoError(t, err)
	// require.Equal(t, *dummyExtensionItem.GetExtension(), *nextItem.GetExtension())

	// // check the store key
	// require.Equal(t, storeKey, target.storeKey)

	// // check the restored tree hash
	// require.Equal(t, latestHash, target.WorkingHash())
}
