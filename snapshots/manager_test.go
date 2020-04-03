package snapshots_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/snapshots"
)

func TestManager_List(t *testing.T) {
	store, teardown := setupStore(t)
	defer teardown()
	manager := snapshots.NewManager(store, nil)

	mgrList, err := manager.List()
	require.NoError(t, err)
	storeList, err := store.List()
	require.NoError(t, err)

	require.NotEmpty(t, storeList)
	assert.Equal(t, storeList, mgrList)

	// list should not block or error on busy managers
	manager, teardown = setupBusyManager(t)
	defer teardown()
	list, err := manager.List()
	require.NoError(t, err)
	assert.Equal(t, []*snapshots.Snapshot{}, list)
}

func TestManager_LoadChunk(t *testing.T) {
	store, teardown := setupStore(t)
	defer teardown()
	manager := snapshots.NewManager(store, nil)

	// Existing chunk should return body
	chunk, err := manager.LoadChunk(2, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, []byte{2, 1, 1}, chunk)

	// Missing chunk should return nil
	chunk, err = manager.LoadChunk(2, 1, 9)
	require.NoError(t, err)
	assert.Nil(t, chunk)

	// LoadChunk should not block or error on busy managers
	manager, teardown = setupBusyManager(t)
	defer teardown()
	chunk, err = manager.LoadChunk(2, 1, 0)
	require.NoError(t, err)
	assert.Nil(t, chunk)
}

func TestManager_Take(t *testing.T) {
	store, teardown := setupStore(t)
	defer teardown()
	snapshotter := &mockSnapshotter{
		chunks: [][]byte{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
		},
	}
	manager := snapshots.NewManager(store, snapshotter)

	// nil manager should return error
	_, err := (*snapshots.Manager)(nil).Take(1)
	require.Error(t, err)

	// taking a snapshot at a lower height than the latest should error
	_, err = manager.Take(3)
	require.Error(t, err)

	// taking a snapshot at a higher height should be fine, and should return it
	snapshot, err := manager.Take(5)
	require.NoError(t, err)
	assert.Equal(t, &snapshots.Snapshot{
		Height: 5,
		Format: snapshots.CurrentFormat,
		ChunkHashes: [][]byte{
			checksum([]byte{1, 2, 3}),
			checksum([]byte{4, 5, 6}),
			checksum([]byte{7, 8, 9}),
		},
	}, snapshot)

	storeSnapshot, chunks, err := store.Load(snapshot.Height, snapshot.Format)
	require.NoError(t, err)
	assert.Equal(t, snapshot, storeSnapshot)
	assert.Equal(t, [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, readChunks(chunks))

	// taking a snapshot while a different snapshot is being taken should error
	manager, teardown = setupBusyManager(t)
	defer teardown()
	_, err = manager.Take(9)
	require.Error(t, err)
}

func TestManager_Prune(t *testing.T) {
	store, teardown := setupStore(t)
	defer teardown()
	manager := snapshots.NewManager(store, nil)

	pruned, err := manager.Prune(2)
	require.NoError(t, err)
	assert.EqualValues(t, 1, pruned)

	list, err := manager.List()
	require.NoError(t, err)
	assert.Len(t, list, 3)

	// Prune should error while a snapshot is being taken
	manager, teardown = setupBusyManager(t)
	defer teardown()
	_, err = manager.Prune(2)
	require.Error(t, err)
}

func TestManager_Restore(t *testing.T) {
	store, teardown := setupStore(t)
	defer teardown()
	target := &mockSnapshotter{}
	manager := snapshots.NewManager(store, target)

	chunks := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	chunkHashes := checksums(chunks)

	// Restore errors on invalid format
	err := manager.Restore(snapshots.Snapshot{
		Height:      3,
		Format:      0,
		ChunkHashes: chunkHashes,
	})
	require.Error(t, err)
	require.Equal(t, err, snapshots.ErrUnknownFormat)

	// Restore errors on no chunks
	err = manager.Restore(snapshots.Snapshot{Height: 3, Format: 1})
	require.Error(t, err)

	// Starting a restore works
	err = manager.Restore(snapshots.Snapshot{
		Height:      3,
		Format:      1,
		ChunkHashes: chunkHashes,
	})
	require.NoError(t, err)

	// While the restore is in progress, any other operations fail
	_, err = manager.Take(4)
	require.Error(t, err)

	// Feeding the chunks should work
	for i, chunk := range chunks {
		done, err := manager.RestoreChunk(chunk)
		require.NoError(t, err)
		if i == len(chunks)-1 {
			assert.True(t, done)
		} else {
			assert.False(t, done)
		}
	}

	assert.Equal(t, chunks, target.chunks)

	// Starting a new restore should fail now, because the target already has contents.
	err = manager.Restore(snapshots.Snapshot{
		Height:      3,
		Format:      1,
		ChunkHashes: chunkHashes,
	})
	require.Error(t, err)

	// But if we clear out the target we should be able to start a new restore. This time we'll
	// fail it with a checksum error. That error should stop the operation, so that we can do
	// a prune operation right after.
	target.chunks = nil
	err = manager.Restore(snapshots.Snapshot{
		Height:      3,
		Format:      1,
		ChunkHashes: chunkHashes,
	})
	require.NoError(t, err)

	_, err = manager.RestoreChunk([]byte{9, 9, 9})
	require.Error(t, err)

	_, err = manager.Prune(1)
	require.NoError(t, err)
}
