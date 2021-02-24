package snapshots

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gogo/protobuf/proto"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/snapshots/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// keyPrefixSnapshot is the prefix for snapshot database keys
	keyPrefixSnapshot byte = 0x01
)

// Store is a snapshot store, containing snapshot metadata and binary chunks.
type Store struct {
	db  db.DB
	dir string

	mtx    sync.Mutex
	saving map[uint64]bool // heights currently being saved
}

// NewStore creates a new snapshot store.
func NewStore(db db.DB, dir string) (*Store, error) {
	if dir == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "snapshot directory not given")
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to create snapshot directory %q", dir)
	}

	return &Store{
		db:     db,
		dir:    dir,
		saving: make(map[uint64]bool),
	}, nil
}

// Delete deletes a snapshot.
func (s *Store) Delete(height uint64, format uint32) error {
	s.mtx.Lock()
	saving := s.saving[height]
	s.mtx.Unlock()
	if saving {
		return sdkerrors.Wrapf(sdkerrors.ErrConflict,
			"snapshot for height %v format %v is currently being saved", height, format)
	}
	err := s.db.DeleteSync(encodeKey(height, format))
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to delete snapshot for height %v format %v",
			height, format)
	}
	err = os.RemoveAll(s.pathSnapshot(height, format))
	return sdkerrors.Wrapf(err, "failed to delete snapshot chunks for height %v format %v",
		height, format)
}

// Get fetches snapshot info from the database.
func (s *Store) Get(height uint64, format uint32) (*types.Snapshot, error) {
	bytes, err := s.db.Get(encodeKey(height, format))
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to fetch snapshot metadata for height %v format %v",
			height, format)
	}
	if bytes == nil {
		return nil, nil
	}
	snapshot := &types.Snapshot{}
	err = proto.Unmarshal(bytes, snapshot)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to decode snapshot metadata for height %v format %v",
			height, format)
	}
	if snapshot.Metadata.ChunkHashes == nil {
		snapshot.Metadata.ChunkHashes = [][]byte{}
	}
	return snapshot, nil
}

// Get fetches the latest snapshot from the database, if any.
func (s *Store) GetLatest() (*types.Snapshot, error) {
	iter, err := s.db.ReverseIterator(encodeKey(0, 0), encodeKey(uint64(math.MaxUint64), math.MaxUint32))
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to find latest snapshot")
	}
	defer iter.Close()

	var snapshot *types.Snapshot
	if iter.Valid() {
		snapshot = &types.Snapshot{}
		err := proto.Unmarshal(iter.Value(), snapshot)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to decode latest snapshot")
		}
	}
	err = iter.Error()
	return snapshot, sdkerrors.Wrap(err, "failed to find latest snapshot")
}

// List lists snapshots, in reverse order (newest first).
func (s *Store) List() ([]*types.Snapshot, error) {
	iter, err := s.db.ReverseIterator(encodeKey(0, 0), encodeKey(uint64(math.MaxUint64), math.MaxUint32))
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to list snapshots")
	}
	defer iter.Close()

	snapshots := make([]*types.Snapshot, 0)
	for ; iter.Valid(); iter.Next() {
		snapshot := &types.Snapshot{}
		err := proto.Unmarshal(iter.Value(), snapshot)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to decode snapshot info")
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, iter.Error()
}

// Load loads a snapshot (both metadata and binary chunks). The chunks must be consumed and closed.
// Returns nil if the snapshot does not exist.
func (s *Store) Load(height uint64, format uint32) (*types.Snapshot, <-chan io.ReadCloser, error) {
	snapshot, err := s.Get(height, format)
	if snapshot == nil || err != nil {
		return nil, nil, err
	}

	ch := make(chan io.ReadCloser)
	go func() {
		defer close(ch)
		for i := uint32(0); i < snapshot.Chunks; i++ {
			pr, pw := io.Pipe()
			ch <- pr
			chunk, err := s.loadChunkFile(height, format, i)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			defer chunk.Close()
			_, err = io.Copy(pw, chunk)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			chunk.Close()
			pw.Close()
		}
	}()

	return snapshot, ch, nil
}

// LoadChunk loads a chunk from disk, or returns nil if it does not exist. The caller must call
// Close() on it when done.
func (s *Store) LoadChunk(height uint64, format uint32, chunk uint32) (io.ReadCloser, error) {
	path := s.pathChunk(height, format, chunk)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return file, err
}

// loadChunkFile loads a chunk from disk, and errors if it does not exist.
func (s *Store) loadChunkFile(height uint64, format uint32, chunk uint32) (io.ReadCloser, error) {
	path := s.pathChunk(height, format, chunk)
	return os.Open(path)
}

// Prune removes old snapshots. The given number of most recent heights (regardless of format) are retained.
func (s *Store) Prune(retain uint32) (uint64, error) {
	iter, err := s.db.ReverseIterator(encodeKey(0, 0), encodeKey(uint64(math.MaxUint64), math.MaxUint32))
	if err != nil {
		return 0, sdkerrors.Wrap(err, "failed to prune snapshots")
	}
	defer iter.Close()

	pruned := uint64(0)
	prunedHeights := make(map[uint64]bool)
	skip := make(map[uint64]bool)
	for ; iter.Valid(); iter.Next() {
		height, format, err := decodeKey(iter.Key())
		if err != nil {
			return 0, sdkerrors.Wrap(err, "failed to prune snapshots")
		}
		if skip[height] || uint32(len(skip)) < retain {
			skip[height] = true
			continue
		}
		err = s.Delete(height, format)
		if err != nil {
			return 0, sdkerrors.Wrap(err, "failed to prune snapshots")
		}
		pruned++
		prunedHeights[height] = true
	}
	// Since Delete() deletes a specific format, while we want to prune a height, we clean up
	// the height directory as well
	for height, ok := range prunedHeights {
		if ok {
			err = os.Remove(s.pathHeight(height))
			if err != nil {
				return 0, sdkerrors.Wrapf(err, "failed to remove snapshot directory for height %v", height)
			}
		}
	}
	return pruned, iter.Error()
}

// Save saves a snapshot to disk, returning it.
func (s *Store) Save(
	height uint64, format uint32, chunks <-chan io.ReadCloser,
) (*types.Snapshot, error) {
	defer DrainChunks(chunks)
	if height == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "snapshot height cannot be 0")
	}

	s.mtx.Lock()
	saving := s.saving[height]
	s.saving[height] = true
	s.mtx.Unlock()
	if saving {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrConflict,
			"a snapshot for height %v is already being saved", height)
	}
	defer func() {
		s.mtx.Lock()
		delete(s.saving, height)
		s.mtx.Unlock()
	}()

	exists, err := s.db.Has(encodeKey(height, format))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrConflict,
			"snapshot already exists for height %v format %v", height, format)
	}

	snapshot := &types.Snapshot{
		Height: height,
		Format: format,
	}
	index := uint32(0)
	snapshotHasher := sha256.New()
	chunkHasher := sha256.New()
	for chunkBody := range chunks {
		defer chunkBody.Close() // nolint: staticcheck
		dir := s.pathSnapshot(height, format)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to create snapshot directory %q", dir)
		}
		path := s.pathChunk(height, format, index)
		file, err := os.Create(path)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to create snapshot chunk file %q", path)
		}
		defer file.Close() // nolint: staticcheck

		chunkHasher.Reset()
		_, err = io.Copy(io.MultiWriter(file, chunkHasher, snapshotHasher), chunkBody)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to generate snapshot chunk %v", index)
		}
		err = file.Close()
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to close snapshot chunk %v", index)
		}
		err = chunkBody.Close()
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to close snapshot chunk %v", index)
		}
		snapshot.Metadata.ChunkHashes = append(snapshot.Metadata.ChunkHashes, chunkHasher.Sum(nil))
		index++
	}
	snapshot.Chunks = index
	snapshot.Hash = snapshotHasher.Sum(nil)
	return snapshot, s.saveSnapshot(snapshot)
}

// saveSnapshot saves snapshot metadata to the database.
func (s *Store) saveSnapshot(snapshot *types.Snapshot) error {
	value, err := proto.Marshal(snapshot)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to encode snapshot metadata")
	}
	err = s.db.SetSync(encodeKey(snapshot.Height, snapshot.Format), value)
	return sdkerrors.Wrap(err, "failed to store snapshot")
}

// pathHeight generates the path to a height, containing multiple snapshot formats.
func (s *Store) pathHeight(height uint64) string {
	return filepath.Join(s.dir, strconv.FormatUint(height, 10))
}

// pathSnapshot generates a snapshot path, as a specific format under a height.
func (s *Store) pathSnapshot(height uint64, format uint32) string {
	return filepath.Join(s.pathHeight(height), strconv.FormatUint(uint64(format), 10))
}

// pathChunk generates a snapshot chunk path.
func (s *Store) pathChunk(height uint64, format uint32, chunk uint32) string {
	return filepath.Join(s.pathSnapshot(height, format), strconv.FormatUint(uint64(chunk), 10))
}

// decodeKey decodes a snapshot key.
func decodeKey(k []byte) (uint64, uint32, error) {
	if len(k) != 13 {
		return 0, 0, sdkerrors.Wrapf(sdkerrors.ErrLogic, "invalid snapshot key with length %v", len(k))
	}
	if k[0] != keyPrefixSnapshot {
		return 0, 0, sdkerrors.Wrapf(sdkerrors.ErrLogic, "invalid snapshot key prefix %x", k[0])
	}
	height := binary.BigEndian.Uint64(k[1:9])
	format := binary.BigEndian.Uint32(k[9:13])
	return height, format, nil
}

// encodeKey encodes a snapshot key.
func encodeKey(height uint64, format uint32) []byte {
	k := make([]byte, 13)
	k[0] = keyPrefixSnapshot
	binary.BigEndian.PutUint64(k[1:], height)
	binary.BigEndian.PutUint32(k[9:], format)
	return k
}
