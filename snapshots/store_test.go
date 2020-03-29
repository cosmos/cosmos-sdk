package snapshots_test

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/snapshots/types"
)

func setup(t *testing.T) (*snapshots.Store, func()) {
	tempdir, err := ioutil.TempDir("", "snapshots")
	require.NoError(t, err)

	store, err := snapshots.NewStore(db.NewMemDB(), tempdir)
	require.NoError(t, err)

	err = store.Save(1, 1, makeChunks([][]byte{
		{1, 1, 0}, {1, 1, 1},
	}))
	require.NoError(t, err)
	err = store.Save(2, 1, makeChunks([][]byte{
		{2, 1, 0}, {2, 1, 1},
	}))
	require.NoError(t, err)
	err = store.Save(2, 2, makeChunks([][]byte{
		{2, 2, 0}, {2, 2, 1}, {2, 2, 2},
	}))
	require.NoError(t, err)
	err = store.Save(3, 2, makeChunks([][]byte{
		{3, 2, 0}, {3, 2, 1}, {3, 2, 2},
	}))
	require.NoError(t, err)

	teardown := func() {
		err := os.RemoveAll(tempdir)
		if err != nil {
			t.Logf("Failed to remove tempdir %q", tempdir)
		}
	}
	return store, teardown
}

func makeChunks(chunks [][]byte) <-chan io.ReadCloser {
	ch := make(chan io.ReadCloser, len(chunks))
	for _, chunk := range chunks {
		ch <- ioutil.NopCloser(bytes.NewReader(chunk))
	}
	close(ch)
	return ch
}

func checksum(b []byte) []byte {
	hash := sha256.Sum256(b)
	return hash[:]
}

func TestNewStore(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "snapshots")
	require.NoError(t, err)
	defer os.RemoveAll(tempdir)

	_, err = snapshots.NewStore(db.NewMemDB(), tempdir)
	require.NoError(t, err)
}

func TestNewStore_ErrNoDir(t *testing.T) {
	_, err := snapshots.NewStore(db.NewMemDB(), "")
	require.Error(t, err)
}

func TestNewStore_ErrDirFailure(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "snapshots")
	require.NoError(t, err)
	defer func() {
		os.RemoveAll(tempfile.Name())
		tempfile.Close()
	}()
	tempdir := filepath.Join(tempfile.Name(), "subdir")

	_, err = snapshots.NewStore(db.NewMemDB(), tempdir)
	require.Error(t, err)
}

func TestStore_Active(t *testing.T) {
	store, teardown := setup(t)
	defer teardown()

	require.False(t, store.Active())

	// Start saving a snapshot, but don't give it any data
	ch := make(chan io.ReadCloser)
	go store.Save(9, 1, ch)

	// The store should now be active (it's saving the above snapshot)
	time.Sleep(10 * time.Millisecond)
	require.True(t, store.Active())

	// When we close the channel, the snapshot is done
	close(ch)
	time.Sleep(10 * time.Millisecond)
	require.False(t, store.Active())
}

func TestStore_Delete(t *testing.T) {
	store, teardown := setup(t)
	defer teardown()

	// Deleting a snapshot should remove it
	err := store.Delete(2, 2)
	require.NoError(t, err)

	snapshot, err := store.Get(2, 2)
	require.NoError(t, err)
	assert.Nil(t, snapshot)

	snapshots, err := store.List()
	require.NoError(t, err)
	assert.Len(t, snapshots, 3)

	// Deleting it again should not error
	err = store.Delete(2, 2)
	require.NoError(t, err)

	// Deleting a snapshot being saved should error
	ch := make(chan io.ReadCloser)
	go store.Save(9, 1, ch)

	time.Sleep(10 * time.Millisecond)
	err = store.Delete(9, 1)
	require.Error(t, err)

	// But after it's saved it should work
	close(ch)
	time.Sleep(10 * time.Millisecond)
	err = store.Delete(9, 1)
	require.NoError(t, err)
}

func TestStore_Get(t *testing.T) {
	store, teardown := setup(t)
	defer teardown()

	// Loading a missing snapshot should return nil
	snapshot, err := store.Get(9, 9)
	require.NoError(t, err)
	assert.Nil(t, snapshot)

	// Loading a snapshot should returns its metadata
	snapshot, err = store.Get(2, 1)
	require.NoError(t, err)
	assert.Equal(t, &types.Snapshot{
		Height: 2,
		Format: 1,
		Chunks: []*types.Chunk{
			{Hash: checksum([]byte{2, 1, 0})},
			{Hash: checksum([]byte{2, 1, 1})},
		},
	}, snapshot)
}

func TestStore_List(t *testing.T) {
	store, teardown := setup(t)
	defer teardown()

	snapshots, err := store.List()
	require.NoError(t, err)

	require.Equal(t, []*types.Snapshot{
		{Height: 3, Format: 2, Chunks: []*types.Chunk{
			{Hash: checksum([]byte{3, 2, 0})},
			{Hash: checksum([]byte{3, 2, 1})},
			{Hash: checksum([]byte{3, 2, 2})},
		}},
		{Height: 2, Format: 2, Chunks: []*types.Chunk{
			{Hash: checksum([]byte{2, 2, 0})},
			{Hash: checksum([]byte{2, 2, 1})},
			{Hash: checksum([]byte{2, 2, 2})},
		}},
		{Height: 2, Format: 1, Chunks: []*types.Chunk{
			{Hash: checksum([]byte{2, 1, 0})},
			{Hash: checksum([]byte{2, 1, 1})},
		}},
		{Height: 1, Format: 1, Chunks: []*types.Chunk{
			{Hash: checksum([]byte{1, 1, 0})},
			{Hash: checksum([]byte{1, 1, 1})},
		}},
	}, snapshots)
}

func TestStore_Load(t *testing.T) {
	store, teardown := setup(t)
	defer teardown()

	// Loading a missing snapshot should return nil
	snapshot, chunks, err := store.Load(9, 9)
	require.NoError(t, err)
	assert.Nil(t, snapshot)
	assert.Nil(t, chunks)

	// Loading a snapshot should returns its metadata and chunks
	snapshot, chunks, err = store.Load(2, 1)
	require.NoError(t, err)
	assert.Equal(t, &types.Snapshot{
		Height: 2,
		Format: 1,
		Chunks: []*types.Chunk{
			{Hash: checksum([]byte{2, 1, 0})},
			{Hash: checksum([]byte{2, 1, 1})},
		},
	}, snapshot)

	for i := range snapshot.Chunks {
		reader, ok := <-chunks
		require.True(t, ok)
		chunk, err := ioutil.ReadAll(reader)
		require.NoError(t, err)
		err = reader.Close()
		require.NoError(t, err)
		assert.Equal(t, []byte{2, 1, byte(i)}, chunk)
	}
	assert.Empty(t, chunks)
}

func TestStore_LoadChunk(t *testing.T) {
	store, teardown := setup(t)
	defer teardown()

	// Loading a missing snapshot should return nil
	chunk, err := store.LoadChunk(9, 9, 0)
	require.NoError(t, err)
	assert.Nil(t, chunk)

	// Loading a missing chunk index should return nil
	chunk, err = store.LoadChunk(2, 1, 2)
	require.NoError(t, err)
	require.Nil(t, chunk)

	// Loading a chunk should returns a content reader
	chunk, err = store.LoadChunk(2, 1, 0)
	require.NoError(t, err)
	require.NotNil(t, chunk)
	body, err := ioutil.ReadAll(chunk)
	require.NoError(t, err)
	assert.Equal(t, []byte{2, 1, 0}, body)
	err = chunk.Close()
	require.NoError(t, err)
}

func TestStore_Prune(t *testing.T) {
	store, teardown := setup(t)
	defer teardown()

	// Pruning too many snapshots should be fine
	pruned, err := store.Prune(4)
	require.NoError(t, err)
	assert.EqualValues(t, 0, pruned)

	snapshots, err := store.List()
	require.NoError(t, err)
	assert.Len(t, snapshots, 4)

	// Pruning until the last two heights should leave three snapshots (for two heights)
	pruned, err = store.Prune(2)
	require.NoError(t, err)
	assert.EqualValues(t, 1, pruned)

	snapshots, err = store.List()
	require.NoError(t, err)
	require.Equal(t, []*types.Snapshot{
		{Height: 3, Format: 2, Chunks: []*types.Chunk{
			{Hash: checksum([]byte{3, 2, 0})},
			{Hash: checksum([]byte{3, 2, 1})},
			{Hash: checksum([]byte{3, 2, 2})},
		}},
		{Height: 2, Format: 2, Chunks: []*types.Chunk{
			{Hash: checksum([]byte{2, 2, 0})},
			{Hash: checksum([]byte{2, 2, 1})},
			{Hash: checksum([]byte{2, 2, 2})},
		}},
		{Height: 2, Format: 1, Chunks: []*types.Chunk{
			{Hash: checksum([]byte{2, 1, 0})},
			{Hash: checksum([]byte{2, 1, 1})},
		}},
	}, snapshots)

	// Pruning all heights should also be fine
	pruned, err = store.Prune(0)
	require.NoError(t, err)
	assert.EqualValues(t, 3, pruned)

	snapshots, err = store.List()
	require.NoError(t, err)
	assert.Empty(t, snapshots)
}

func TestStore_Save(t *testing.T) {
	store, teardown := setup(t)
	defer teardown()

	// Saving a snapshot should work (TestStore_Load checks that it loads correctly)
	err := store.Save(4, 1, makeChunks([][]byte{{1}, {2}}))
	require.NoError(t, err)

	// Saving an existing snapshot should error
	err = store.Save(4, 1, makeChunks([][]byte{{1}, {2}}))
	require.Error(t, err)

	// Saving at height 0 should error
	err = store.Save(0, 1, makeChunks([][]byte{{1}, {2}}))
	require.Error(t, err)

	// Saving at format 0 should be fine
	err = store.Save(1, 0, makeChunks([][]byte{{1}, {2}}))
	require.NoError(t, err)

	// Saving a snapshot with no chunks should be fine, as should loading it
	err = store.Save(5, 1, makeChunks([][]byte{}))
	require.NoError(t, err)
	snapshot, chunks, err := store.Load(5, 1)
	require.NoError(t, err)
	assert.Equal(t, &types.Snapshot{Height: 5, Format: 1, Chunks: []*types.Chunk{}}, snapshot)
	assert.Empty(t, chunks)

	// Saving a snapshot should error if a chunk reader returns an error, and it should empty out
	// the channel
	someErr := errors.New("boom")
	pr, pw := io.Pipe()
	err = pw.CloseWithError(someErr)
	require.NoError(t, err)

	ch := make(chan io.ReadCloser, 2)
	ch <- pr
	ch <- ioutil.NopCloser(bytes.NewBuffer([]byte{0xff}))
	close(ch)

	err = store.Save(6, 1, ch)
	require.Error(t, err)
	require.True(t, errors.Is(err, someErr))
	assert.Empty(t, ch)

	// Saving a snapshot should error if a snapshot is already in progress for the same height,
	// regardless of format. However, a different height should succeed.
	ch = make(chan io.ReadCloser)
	go store.Save(7, 1, ch)
	time.Sleep(10 * time.Millisecond)
	err = store.Save(7, 2, makeChunks(nil))
	require.Error(t, err)
	err = store.Save(8, 1, makeChunks(nil))
	require.NoError(t, err)
	close(ch)
}
