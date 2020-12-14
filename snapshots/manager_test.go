package snapshots_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/snapshots/types"
)

func TestManager_List(t *testing.T) {
	store := setupStore(t)
	manager := snapshots.NewManager(store, nil)

	mgrList, err := manager.List()
	require.NoError(t, err)
	storeList, err := store.List()
	require.NoError(t, err)

	require.NotEmpty(t, storeList)
	assert.Equal(t, storeList, mgrList)

	// list should not block or error on busy managers
	manager = setupBusyManager(t)
	list, err := manager.List()
	require.NoError(t, err)
	assert.Equal(t, []*types.Snapshot{}, list)
}

func TestManager_LoadChunk(t *testing.T) {
	store := setupStore(t)
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
	manager = setupBusyManager(t)
	chunk, err = manager.LoadChunk(2, 1, 0)
	require.NoError(t, err)
	assert.Nil(t, chunk)
}

func TestManager_Take(t *testing.T) {
	store := setupStore(t)
	snapshotter := &mockSnapshotter{
		chunks: [][]byte{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
		},
	}
	manager := snapshots.NewManager(store, snapshotter)

	// nil manager should return error
	_, err := (*snapshots.Manager)(nil).Create(1)
	require.Error(t, err)

	// creating a snapshot at a lower height than the latest should error
	_, err = manager.Create(3)
	require.Error(t, err)

	// creating a snapshot at a higher height should be fine, and should return it
	snapshot, err := manager.Create(5)
	require.NoError(t, err)
	assert.Equal(t, &types.Snapshot{
		Height: 5,
		Format: types.CurrentFormat,
		Chunks: 3,
		Hash:   []uint8{0x47, 0xe4, 0xee, 0x7f, 0x21, 0x1f, 0x73, 0x26, 0x5d, 0xd1, 0x76, 0x58, 0xf6, 0xe2, 0x1c, 0x13, 0x18, 0xbd, 0x6c, 0x81, 0xf3, 0x75, 0x98, 0xe2, 0xa, 0x27, 0x56, 0x29, 0x95, 0x42, 0xef, 0xcf},
		Metadata: types.Metadata{
			ChunkHashes: checksums([][]byte{
				{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}),
		},
	}, snapshot)

	storeSnapshot, chunks, err := store.Load(snapshot.Height, snapshot.Format)
	require.NoError(t, err)
	assert.Equal(t, snapshot, storeSnapshot)
	assert.Equal(t, [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, readChunks(chunks))

	// creating a snapshot while a different snapshot is being created should error
	manager = setupBusyManager(t)
	_, err = manager.Create(9)
	require.Error(t, err)
}

func TestManager_Prune(t *testing.T) {
	store := setupStore(t)
	manager := snapshots.NewManager(store, nil)

	pruned, err := manager.Prune(2)
	require.NoError(t, err)
	assert.EqualValues(t, 1, pruned)

	list, err := manager.List()
	require.NoError(t, err)
	assert.Len(t, list, 3)

	// Prune should error while a snapshot is being taken
	manager = setupBusyManager(t)
	_, err = manager.Prune(2)
	require.Error(t, err)
}

func TestManager_Restore(t *testing.T) {
	store := setupStore(t)
	target := &mockSnapshotter{}
	manager := snapshots.NewManager(store, target)

	chunks := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	// Restore errors on invalid format
	err := manager.Restore(types.Snapshot{
		Height:   3,
		Format:   0,
		Hash:     []byte{1, 2, 3},
		Chunks:   uint32(len(chunks)),
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.Error(t, err)
	require.Equal(t, types.ErrUnknownFormat, err)

	// Restore errors on no chunks
	err = manager.Restore(types.Snapshot{Height: 3, Format: 1, Hash: []byte{1, 2, 3}})
	require.Error(t, err)

	// Restore errors on chunk and chunkhashes mismatch
	err = manager.Restore(types.Snapshot{
		Height:   3,
		Format:   1,
		Hash:     []byte{1, 2, 3},
		Chunks:   4,
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.Error(t, err)

	// Starting a restore works
	err = manager.Restore(types.Snapshot{
		Height:   3,
		Format:   1,
		Hash:     []byte{1, 2, 3},
		Chunks:   3,
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.NoError(t, err)

	// While the restore is in progress, any other operations fail
	_, err = manager.Create(4)
	require.Error(t, err)

	_, err = manager.Prune(1)
	require.Error(t, err)

	// Feeding an invalid chunk should error due to invalid checksum, but not abort restoration.
	_, err = manager.RestoreChunk([]byte{9, 9, 9})
	require.Error(t, err)
	require.True(t, errors.Is(err, types.ErrChunkHashMismatch))

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
	err = manager.Restore(types.Snapshot{
		Height:   3,
		Format:   1,
		Hash:     []byte{1, 2, 3},
		Chunks:   3,
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.Error(t, err)

	// But if we clear out the target we should be able to start a new restore. This time we'll
	// fail it with a checksum error. That error should stop the operation, so that we can do
	// a prune operation right after.
	target.chunks = nil
	err = manager.Restore(types.Snapshot{
		Height:   3,
		Format:   1,
		Hash:     []byte{1, 2, 3},
		Chunks:   3,
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.NoError(t, err)
}
