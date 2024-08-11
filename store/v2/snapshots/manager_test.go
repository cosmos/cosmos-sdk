package snapshots_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/snapshots"
	"cosmossdk.io/store/v2/snapshots/types"
)

var opts = snapshots.NewSnapshotOptions(1500, 2)

func TestManager_List(t *testing.T) {
	store := setupStore(t)
	commitSnapshotter := &mockCommitSnapshotter{}
	storageSnapshotter := &mockStorageSnapshotter{}
	manager := snapshots.NewManager(store, opts, commitSnapshotter, storageSnapshotter, nil, coretesting.NewNopLogger())

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

	require.NoError(t, manager.Close())
}

func TestManager_LoadChunk(t *testing.T) {
	store := setupStore(t)
	manager := snapshots.NewManager(store, opts, &mockCommitSnapshotter{}, &mockStorageSnapshotter{}, nil, coretesting.NewNopLogger())

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
	commitSnapshotter := &mockCommitSnapshotter{
		items: items,
	}
	extSnapshotter := newExtSnapshotter(10)

	expectChunks := snapshotItems(items, extSnapshotter)
	manager := snapshots.NewManager(store, opts, commitSnapshotter, &mockStorageSnapshotter{}, nil, coretesting.NewNopLogger())
	err := manager.RegisterExtensions(extSnapshotter)
	require.NoError(t, err)

	// nil manager should return error
	_, err = (*snapshots.Manager)(nil).Create(1)
	require.Error(t, err)

	// creating a snapshot at a lower height than the latest should error
	_, err = manager.Create(3)
	require.Error(t, err)

	// creating a snapshot at a higher height should be fine, and should return it
	snapshot, err := manager.Create(5)
	require.NoError(t, err)

	assert.Equal(t, &types.Snapshot{
		Height: 5,
		Format: commitSnapshotter.SnapshotFormat(),
		Chunks: 1,
		Hash:   []uint8{0xc5, 0xf7, 0xfe, 0xea, 0xd3, 0x4d, 0x3e, 0x87, 0xff, 0x41, 0xa2, 0x27, 0xfa, 0xcb, 0x38, 0x17, 0xa, 0x5, 0xeb, 0x27, 0x4e, 0x16, 0x5e, 0xf3, 0xb2, 0x8b, 0x47, 0xd1, 0xe6, 0x94, 0x7e, 0x8b},
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
	manager := snapshots.NewManager(store, opts, &mockCommitSnapshotter{}, &mockStorageSnapshotter{}, nil, coretesting.NewNopLogger())

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
	target := &mockCommitSnapshotter{}
	storageSnapshotter := &mockStorageSnapshotter{items: map[string][]byte{}}
	extSnapshotter := newExtSnapshotter(0)
	manager := snapshots.NewManager(store, opts, target, storageSnapshotter, nil, coretesting.NewNopLogger())
	err := manager.RegisterExtensions(extSnapshotter)
	require.NoError(t, err)

	expectItems := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	chunks := snapshotItems(expectItems, newExtSnapshotter(10))

	// Restore errors on invalid format
	err = manager.Restore(types.Snapshot{
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
	assert.Equal(t, 10, len(extSnapshotter.state))

	// make sure storageSnapshotter items are properly stored
	for i, item := range target.items {
		key := fmt.Sprintf("key-%d", i)
		chunk := storageSnapshotter.items[key]
		require.NotNil(t, chunk)
		require.Equal(t, item, chunk)
	}

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
}

func TestManager_TakeError(t *testing.T) {
	snapshotter := &mockErrorCommitSnapshotter{}
	store, err := snapshots.NewStore(t.TempDir())
	require.NoError(t, err)
	manager := snapshots.NewManager(store, opts, snapshotter, &mockStorageSnapshotter{}, nil, coretesting.NewNopLogger())

	_, err = manager.Create(1)
	require.Error(t, err)
}

func TestSnapshot_Take_Restore(t *testing.T) {
	store := setupStore(t)
	items := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	commitSnapshotter := &mockCommitSnapshotter{
		items: items,
	}
	storageSnapshotter := &mockStorageSnapshotter{items: map[string][]byte{}}

	extSnapshotter := newExtSnapshotter(10)

	expectChunks := snapshotItems(items, extSnapshotter)
	manager := snapshots.NewManager(store, opts, commitSnapshotter, storageSnapshotter, nil, coretesting.NewNopLogger())
	err := manager.RegisterExtensions(extSnapshotter)
	require.NoError(t, err)

	// creating a snapshot at a higher height should be fine, and should return it
	snapshot, err := manager.Create(5)
	require.NoError(t, err)

	assert.Equal(t, &types.Snapshot{
		Height: 5,
		Format: commitSnapshotter.SnapshotFormat(),
		Chunks: 1,
		Hash:   []uint8{0xc5, 0xf7, 0xfe, 0xea, 0xd3, 0x4d, 0x3e, 0x87, 0xff, 0x41, 0xa2, 0x27, 0xfa, 0xcb, 0x38, 0x17, 0xa, 0x5, 0xeb, 0x27, 0x4e, 0x16, 0x5e, 0xf3, 0xb2, 0x8b, 0x47, 0xd1, 0xe6, 0x94, 0x7e, 0x8b},
		Metadata: types.Metadata{
			ChunkHashes: checksums(expectChunks),
		},
	}, snapshot)

	storeSnapshot, chunks, err := store.Load(snapshot.Height, snapshot.Format)
	require.NoError(t, err)
	assert.Equal(t, snapshot, storeSnapshot)
	assert.Equal(t, expectChunks, readChunks(chunks))

	err = manager.Restore(*snapshot)
	require.NoError(t, err)

	// Feeding the chunks should work
	for i, chunk := range readChunks(chunks) {
		done, err := manager.RestoreChunk(chunk)
		require.NoError(t, err)
		if i == len(chunks)-1 {
			assert.True(t, done)
		} else {
			assert.False(t, done)
		}
	}

	// The snapshot is saved in local snapshot store
	snapshots, err := store.List()
	require.NoError(t, err)
	require.Equal(t, uint64(5), snapshots[0].Height)
	require.Equal(t, types.CurrentFormat, snapshots[0].Format)

	// Starting a new restore should fail now, because the target already has contents.
	err = manager.Restore(*snapshot)
	require.Error(t, err)

	storeSnapshot, chunks, err = store.Load(snapshot.Height, snapshot.Format)
	require.NoError(t, err)
	assert.Equal(t, snapshot, storeSnapshot)
	assert.Equal(t, expectChunks, readChunks(chunks))

	// Feeding the chunks should work
	for i, chunk := range readChunks(chunks) {
		done, err := manager.RestoreChunk(chunk)
		require.NoError(t, err)
		if i == len(chunks)-1 {
			assert.True(t, done)
		} else {
			assert.False(t, done)
		}
	}

	assert.Equal(t, items, commitSnapshotter.items)
	assert.Equal(t, 10, len(extSnapshotter.state))

	snapshots, err = store.List()
	require.NoError(t, err)
	require.Equal(t, uint64(5), snapshots[0].Height)
	require.Equal(t, types.CurrentFormat, snapshots[0].Format)
}

func TestSnapshot_Take_Prune(t *testing.T) {
	store := setupStore(t)

	items := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	commitSnapshotter := &mockCommitSnapshotter{
		items: items,
	}
	extSnapshotter := newExtSnapshotter(10)

	expectChunks := snapshotItems(items, extSnapshotter)
	manager := snapshots.NewManager(store, opts, commitSnapshotter, &mockStorageSnapshotter{}, nil, coretesting.NewNopLogger())
	err := manager.RegisterExtensions(extSnapshotter)
	require.NoError(t, err)

	// creating a snapshot at height 4
	snapshot, err := manager.Create(4)
	require.NoError(t, err)

	assert.Equal(t, &types.Snapshot{
		Height: 4,
		Format: commitSnapshotter.SnapshotFormat(),
		Chunks: 1,
		Hash:   []uint8{0xc5, 0xf7, 0xfe, 0xea, 0xd3, 0x4d, 0x3e, 0x87, 0xff, 0x41, 0xa2, 0x27, 0xfa, 0xcb, 0x38, 0x17, 0xa, 0x5, 0xeb, 0x27, 0x4e, 0x16, 0x5e, 0xf3, 0xb2, 0x8b, 0x47, 0xd1, 0xe6, 0x94, 0x7e, 0x8b},
		Metadata: types.Metadata{
			ChunkHashes: checksums(expectChunks),
		},
	}, snapshot)

	pruned, err := manager.Prune(1)
	require.NoError(t, err)
	assert.EqualValues(t, 4, pruned)

	// creating a snapshot at a same height 4, should be error
	// since we prune all the previous snapshot except the latest at height 4
	_, err = manager.Create(4)
	require.Error(t, err)

	// prune all
	pruned, err = manager.Prune(0)
	require.NoError(t, err)
	assert.EqualValues(t, 1, pruned)

	// creating a snapshot at a same height 4, should be true since we prune all the previous snapshot
	snapshot, err = manager.Create(4)
	require.NoError(t, err)

	assert.Equal(t, &types.Snapshot{
		Height: 4,
		Format: commitSnapshotter.SnapshotFormat(),
		Chunks: 1,
		Hash:   []uint8{0xc5, 0xf7, 0xfe, 0xea, 0xd3, 0x4d, 0x3e, 0x87, 0xff, 0x41, 0xa2, 0x27, 0xfa, 0xcb, 0x38, 0x17, 0xa, 0x5, 0xeb, 0x27, 0x4e, 0x16, 0x5e, 0xf3, 0xb2, 0x8b, 0x47, 0xd1, 0xe6, 0x94, 0x7e, 0x8b},
		Metadata: types.Metadata{
			ChunkHashes: checksums(expectChunks),
		},
	}, snapshot)

	storeSnapshot, chunks, err := store.Load(snapshot.Height, snapshot.Format)
	require.NoError(t, err)
	assert.Equal(t, snapshot, storeSnapshot)
	assert.Equal(t, expectChunks, readChunks(chunks))

	pruned, err = manager.Prune(2)
	require.NoError(t, err)
	assert.EqualValues(t, 0, pruned)

	list, err := manager.List()
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Prune should error while a snapshot is being taken
	manager = setupBusyManager(t)
	_, err = manager.Prune(2)
	require.Error(t, err)
}

func TestSnapshot_Pruning_Take_Snapshot_Parallel(t *testing.T) {
	store := setupStore(t)

	items := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	commitSnapshotter := &mockCommitSnapshotter{
		items: items,
	}
	extSnapshotter := newExtSnapshotter(10)

	expectChunks := snapshotItems(items, extSnapshotter)
	manager := snapshots.NewManager(store, opts, commitSnapshotter, &mockStorageSnapshotter{}, nil, coretesting.NewNopLogger())
	err := manager.RegisterExtensions(extSnapshotter)
	require.NoError(t, err)

	var prunedCount uint64
	// try take snapshot and pruning parallel while prune operation begins first
	go func() {
		checkError := func() bool {
			_, err := manager.Create(4)
			return err != nil
		}

		require.Eventually(t, checkError, time.Millisecond*200, time.Millisecond)
	}()

	prunedCount, err = manager.Prune(1)
	require.NoError(t, err)
	assert.EqualValues(t, 3, prunedCount)

	// creating a snapshot at a same height 4, should be true since we prune has finished
	snapshot, err := manager.Create(4)
	require.NoError(t, err)

	assert.Equal(t, &types.Snapshot{
		Height: 4,
		Format: commitSnapshotter.SnapshotFormat(),
		Chunks: 1,
		Hash:   []uint8{0xc5, 0xf7, 0xfe, 0xea, 0xd3, 0x4d, 0x3e, 0x87, 0xff, 0x41, 0xa2, 0x27, 0xfa, 0xcb, 0x38, 0x17, 0xa, 0x5, 0xeb, 0x27, 0x4e, 0x16, 0x5e, 0xf3, 0xb2, 0x8b, 0x47, 0xd1, 0xe6, 0x94, 0x7e, 0x8b},
		Metadata: types.Metadata{
			ChunkHashes: checksums(expectChunks),
		},
	}, snapshot)

	// try take snapshot and pruning parallel while snapshot operation begins first
	go func() {
		checkError := func() bool {
			_, err = manager.Prune(1)
			return err != nil
		}

		require.Eventually(t, checkError, time.Millisecond*200, time.Millisecond)
	}()

	snapshot, err = manager.Create(5)
	require.NoError(t, err)

	assert.Equal(t, &types.Snapshot{
		Height: 5,
		Format: commitSnapshotter.SnapshotFormat(),
		Chunks: 1,
		Hash:   []uint8{0xc5, 0xf7, 0xfe, 0xea, 0xd3, 0x4d, 0x3e, 0x87, 0xff, 0x41, 0xa2, 0x27, 0xfa, 0xcb, 0x38, 0x17, 0xa, 0x5, 0xeb, 0x27, 0x4e, 0x16, 0x5e, 0xf3, 0xb2, 0x8b, 0x47, 0xd1, 0xe6, 0x94, 0x7e, 0x8b},
		Metadata: types.Metadata{
			ChunkHashes: checksums(expectChunks),
		},
	}, snapshot)
}

func TestSnapshot_SnapshotIfApplicable(t *testing.T) {
	store := setupStore(t)

	items := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	commitSnapshotter := &mockCommitSnapshotter{
		items: items,
	}
	extSnapshotter := newExtSnapshotter(10)

	snapshotOpts := snapshots.NewSnapshotOptions(1, 1)

	manager := snapshots.NewManager(store, snapshotOpts, commitSnapshotter, &mockStorageSnapshotter{}, nil, coretesting.NewNopLogger())
	err := manager.RegisterExtensions(extSnapshotter)
	require.NoError(t, err)

	manager.SnapshotIfApplicable(4)

	checkLatestHeight := func() bool {
		latestSnapshot, _ := store.GetLatest()
		return latestSnapshot.Height == 4
	}

	require.Eventually(t, checkLatestHeight, time.Second*10, time.Second)

	pruned, err := manager.Prune(1)
	require.NoError(t, err)
	require.Equal(t, uint64(0), pruned)
}
