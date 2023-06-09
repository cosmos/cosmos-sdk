package snapshots_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/snapshots/types"
)

var opts = types.NewSnapshotOptions(1500, 2)

func TestManager_List(t *testing.T) {
	store := setupStore(t)
	snapshotter := &mockSnapshotter{}
	snapshotter.SetSnapshotInterval(opts.Interval)
	manager := snapshots.NewManager(store, opts, snapshotter, nil, log.NewNopLogger())
	require.Equal(t, opts.Interval, snapshotter.GetSnapshotInterval())

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
	manager := snapshots.NewManager(store, opts, &mockSnapshotter{}, nil, log.NewNopLogger())

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
	items := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	snapshotter := &mockSnapshotter{
		items:         items,
		prunedHeights: make(map[int64]struct{}),
	}
	expectChunks := snapshotItems(items)
	manager := snapshots.NewManager(store, opts, snapshotter, nil, log.NewNopLogger())

	// nil manager should return error
	_, err := (*snapshots.Manager)(nil).Create(1)
	require.Error(t, err)

	// creating a snapshot at a lower height than the latest should error
	_, err = manager.Create(3)
	require.Error(t, err)
	_, didPruneHeight := snapshotter.prunedHeights[3]
	require.True(t, didPruneHeight)

	// creating a snapshot at a higher height should be fine, and should return it
	snapshot, err := manager.Create(5)
	require.NoError(t, err)
	_, didPruneHeight = snapshotter.prunedHeights[5]
	require.True(t, didPruneHeight)

	assert.Equal(t, &types.Snapshot{
		Height: 5,
		Format: snapshotter.SnapshotFormat(),
		Chunks: 1,
		Hash:   []uint8{0x14, 0x38, 0x97, 0x96, 0xba, 0xe4, 0x81, 0xaf, 0x6c, 0xac, 0xff, 0xa5, 0xb8, 0x7e, 0x63, 0x4b, 0xac, 0x69, 0x3f, 0x38, 0x90, 0x5c, 0x7d, 0x57, 0xb3, 0xf, 0x69, 0x73, 0xb3, 0xa0, 0xe0, 0xad},
		Metadata: types.Metadata{
			ChunkHashes: checksums(expectChunks),
		},
	}, snapshot)

	storeSnapshot, chunks, err := store.Load(snapshot.Height, snapshot.Format)
	require.NoError(t, err)
	assert.Equal(t, snapshot, storeSnapshot)
	assert.Equal(t, expectChunks, readChunks(chunks))

	// creating a snapshot while a different snapshot is being created should error
	manager = setupBusyManager(t)
	_, err = manager.Create(9)
	require.Error(t, err)
}

func TestManager_Prune(t *testing.T) {
	store := setupStore(t)
	snapshotter := &mockSnapshotter{}
	snapshotter.SetSnapshotInterval(opts.Interval)
	manager := snapshots.NewManager(store, opts, snapshotter, nil, log.NewNopLogger())

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
	target := &mockSnapshotter{
		prunedHeights: make(map[int64]struct{}),
	}
	manager := snapshots.NewManager(store, opts, target, nil, log.NewNopLogger())

	expectItems := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	chunks := snapshotItems(expectItems)

	// Restore errors on invalid format
	err := manager.Restore(types.Snapshot{
		Height:   3,
		Format:   0,
		Hash:     []byte{1, 2, 3},
		Chunks:   uint32(len(chunks)),
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnknownFormat)

	// Restore errors on no chunks
	err = manager.Restore(types.Snapshot{Height: 3, Format: types.CurrentFormat, Hash: []byte{1, 2, 3}})
	require.Error(t, err)

	// Restore errors on chunk and chunkhashes mismatch
	err = manager.Restore(types.Snapshot{
		Height:   3,
		Format:   types.CurrentFormat,
		Hash:     []byte{1, 2, 3},
		Chunks:   4,
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.Error(t, err)

	// Starting a restore works
	err = manager.Restore(types.Snapshot{
		Height:   3,
		Format:   types.CurrentFormat,
		Hash:     []byte{1, 2, 3},
		Chunks:   1,
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.NoError(t, err)

	// While the restore is in progress, any other operations fail
	_, err = manager.Create(4)
	require.Error(t, err)
	_, didPruneHeight := target.prunedHeights[4]
	require.True(t, didPruneHeight)

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

	assert.Equal(t, expectItems, target.items)

	// The snapshot is saved in local snapshot store
	snapshots, err := store.List()
	require.NoError(t, err)
	snapshot := snapshots[0]
	require.Equal(t, uint64(3), snapshot.Height)
	require.Equal(t, types.CurrentFormat, snapshot.Format)

	// Starting a new restore should fail now, because the target already has contents.
	err = manager.Restore(types.Snapshot{
		Height:   3,
		Format:   types.CurrentFormat,
		Hash:     []byte{1, 2, 3},
		Chunks:   3,
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.Error(t, err)

	// But if we clear out the target we should be able to start a new restore. This time we'll
	// fail it with a checksum error. That error should stop the operation, so that we can do
	// a prune operation right after.
	target.items = nil
	err = manager.Restore(types.Snapshot{
		Height:   3,
		Format:   types.CurrentFormat,
		Hash:     []byte{1, 2, 3},
		Chunks:   1,
		Metadata: types.Metadata{ChunkHashes: checksums(chunks)},
	})
	require.NoError(t, err)
}
