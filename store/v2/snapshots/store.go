package snapshots

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/cosmos/gogoproto/proto"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/errors/v2"
	storeerrors "cosmossdk.io/store/v2/errors"
	"cosmossdk.io/store/v2/snapshots/types"
)

const (
	// keyPrefixSnapshot is the prefix for snapshot database keys
	keyPrefixSnapshot byte = 0x01
)

// Store is a snapshot store, containing snapshot metadata and binary chunks.
type Store struct {
	dir string

	mtx    sync.Mutex
	saving map[uint64]bool // heights currently being saved
}

// NewStore creates a new snapshot store.
func NewStore(dir string) (*Store, error) {
	if dir == "" {
		return nil, fmt.Errorf("snapshot directory not given: %w", storeerrors.ErrLogic)
	}
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory %q: %w", dir, err)
	}
	err = os.MkdirAll(filepath.Join(dir, "metadata"), 0o750)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot metadata directory %q: %w", dir, err)
	}

	return &Store{
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
		return errors.Wrapf(storeerrors.ErrConflict,
			"snapshot for height %v format %v is currently being saved", height, format)
	}
	if err := os.RemoveAll(s.pathSnapshot(height, format)); err != nil {
		return errors.Wrapf(err, "failed to delete snapshot chunks for height %v format %v", height, format)
	}
	if err := os.RemoveAll(s.pathMetadata(height, format)); err != nil {
		return errors.Wrapf(err, "failed to delete snapshot metadata for height %v format %v", height, format)
	}
	return nil
}

// Get fetches snapshot info from the database.
func (s *Store) Get(height uint64, format uint32) (*types.Snapshot, error) {
	if _, err := os.Stat(s.pathMetadata(height, format)); os.IsNotExist(err) {
		return nil, nil
	}
	bytes, err := os.ReadFile(s.pathMetadata(height, format))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch snapshot metadata for height %v format %v",
			height, format)
	}
	snapshot := &types.Snapshot{}
	err = proto.Unmarshal(bytes, snapshot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode snapshot metadata for height %v format %v",
			height, format)
	}
	if snapshot.Metadata.ChunkHashes == nil {
		snapshot.Metadata.ChunkHashes = [][]byte{}
	}
	return snapshot, nil
}

// GetLatest fetches the latest snapshot from the database, if any.
func (s *Store) GetLatest() (*types.Snapshot, error) {
	metadata, err := os.ReadDir(s.pathMetadataDir())
	if err != nil {
		return nil, errors.Wrap(err, "failed to list snapshot metadata")
	}
	if len(metadata) == 0 {
		return nil, nil
	}
	// file system may not guarantee the order of the files, so we sort them lexically
	sort.Slice(metadata, func(i, j int) bool { return metadata[i].Name() < metadata[j].Name() })

	path := filepath.Join(s.pathMetadataDir(), metadata[len(metadata)-1].Name())
	if err := s.validateMetadataPath(path); err != nil {
		return nil, err
	}
	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read latest snapshot metadata %s", path)
	}

	snapshot := &types.Snapshot{}
	err = proto.Unmarshal(bz, snapshot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode latest snapshot metadata %s", path)
	}
	return snapshot, nil
}

// List lists snapshots, in reverse order (newest first).
func (s *Store) List() ([]*types.Snapshot, error) {
	metadata, err := os.ReadDir(s.pathMetadataDir())
	if err != nil {
		return nil, errors.Wrap(err, "failed to list snapshot metadata")
	}
	// file system may not guarantee the order of the files, so we sort them lexically
	sort.Slice(metadata, func(i, j int) bool { return metadata[i].Name() < metadata[j].Name() })

	snapshots := make([]*types.Snapshot, len(metadata))
	for i, entry := range metadata {
		path := filepath.Join(s.pathMetadataDir(), entry.Name())
		if err := s.validateMetadataPath(path); err != nil {
			return nil, err
		}
		bz, err := os.ReadFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read snapshot metadata %s", entry.Name())
		}
		snapshot := &types.Snapshot{}
		err = proto.Unmarshal(bz, snapshot)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode snapshot metadata %s", entry.Name())
		}
		snapshots[len(metadata)-1-i] = snapshot
	}
	return snapshots, nil
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
				_ = pw.CloseWithError(err)
				return
			}
			func() {
				defer chunk.Close()
				_, err = io.Copy(pw, chunk)
				if err != nil {
					_ = pw.CloseWithError(err)
					return
				}
				pw.Close()
			}()
		}
	}()

	return snapshot, ch, nil
}

// LoadChunk loads a chunk from disk, or returns nil if it does not exist. The caller must call
// Close() on it when done.
func (s *Store) LoadChunk(height uint64, format, chunk uint32) (io.ReadCloser, error) {
	path := s.PathChunk(height, format, chunk)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return file, err
}

// loadChunkFile loads a chunk from disk, and errors if it does not exist.
func (s *Store) loadChunkFile(height uint64, format, chunk uint32) (io.ReadCloser, error) {
	path := s.PathChunk(height, format, chunk)
	return os.Open(path)
}

// Prune removes old snapshots. The given number of most recent heights (regardless of format) are retained.
func (s *Store) Prune(retain uint32) (uint64, error) {
	metadata, err := os.ReadDir(s.pathMetadataDir())
	if err != nil {
		return 0, errors.Wrap(err, "failed to list snapshot metadata")
	}

	pruned := uint64(0)
	prunedHeights := make(map[uint64]bool)
	skip := make(map[uint64]bool)
	for i := len(metadata) - 1; i >= 0; i-- {
		height, format, err := s.parseMetadataFilename(metadata[i].Name())
		if err != nil {
			return 0, err
		}

		if skip[height] || uint32(len(skip)) < retain {
			skip[height] = true
			continue
		}
		err = s.Delete(height, format)
		if err != nil {
			return 0, errors.Wrap(err, "failed to prune snapshots")
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
				return 0, errors.Wrapf(err, "failed to remove snapshot directory for height %v", height)
			}
		}
	}
	return pruned, nil
}

// Save saves a snapshot to disk, returning it.
func (s *Store) Save(
	height uint64, format uint32, chunks <-chan io.ReadCloser,
) (*types.Snapshot, error) {
	defer DrainChunks(chunks)
	if height == 0 {
		return nil, errors.Wrap(storeerrors.ErrLogic, "snapshot height cannot be 0")
	}

	s.mtx.Lock()
	saving := s.saving[height]
	s.saving[height] = true
	s.mtx.Unlock()
	if saving {
		return nil, errors.Wrapf(storeerrors.ErrConflict,
			"a snapshot for height %v is already being saved", height)
	}
	defer func() {
		s.mtx.Lock()
		delete(s.saving, height)
		s.mtx.Unlock()
	}()

	snapshot := &types.Snapshot{
		Height: height,
		Format: format,
	}

	// create height directory or do nothing
	if err := os.MkdirAll(s.pathHeight(height), 0o750); err != nil {
		return nil, errors.Wrapf(err, "failed to create snapshot directory for height %v", height)
	}
	// create format directory or fail (if for example the format directory already exists)
	if err := os.Mkdir(s.pathSnapshot(height, format), 0o750); err != nil {
		return nil, errors.Wrapf(err, "failed to create snapshot directory for height %v format %v", height, format)
	}

	index := uint32(0)
	snapshotHasher := sha256.New()
	chunkHasher := sha256.New()
	for chunkBody := range chunks {
		if err := s.saveChunk(chunkBody, index, snapshot, chunkHasher, snapshotHasher); err != nil {
			return nil, err
		}
		index++
	}
	snapshot.Chunks = index
	snapshot.Hash = snapshotHasher.Sum(nil)
	return snapshot, s.saveSnapshot(snapshot)
}

// saveChunk saves the given chunkBody with the given index to its appropriate path on disk.
// The hash of the chunk is appended to the snapshot's metadata,
// and the overall snapshot hash is updated with the chunk content too.
func (s *Store) saveChunk(chunkBody io.ReadCloser, index uint32, snapshot *types.Snapshot, chunkHasher, snapshotHasher hash.Hash) (err error) {
	defer func() {
		if cErr := chunkBody.Close(); cErr != nil {
			err = errors.Wrapf(cErr, "failed to close snapshot chunk body %d", index)
		}
	}()

	path := s.PathChunk(snapshot.Height, snapshot.Format, index)
	chunkFile, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create snapshot chunk file %q", path)
	}

	defer func() {
		if cErr := chunkFile.Close(); cErr != nil {
			err = errors.Wrapf(cErr, "failed to close snapshot chunk file %d", index)
		}
	}()

	chunkHasher.Reset()
	if _, err := io.Copy(io.MultiWriter(chunkFile, chunkHasher, snapshotHasher), chunkBody); err != nil {
		return errors.Wrapf(err, "failed to generate snapshot chunk %d", index)
	}

	snapshot.Metadata.ChunkHashes = append(snapshot.Metadata.ChunkHashes, chunkHasher.Sum(nil))
	return nil
}

// saveChunkContent save the chunk to disk
func (s *Store) saveChunkContent(chunk []byte, index uint32, snapshot *types.Snapshot) error {
	path := s.PathChunk(snapshot.Height, snapshot.Format, index)
	return os.WriteFile(path, chunk, 0o600)
}

// saveSnapshot saves snapshot metadata to the database.
func (s *Store) saveSnapshot(snapshot *types.Snapshot) error {
	value, err := proto.Marshal(snapshot)
	if err != nil {
		return errors.Wrap(err, "failed to encode snapshot metadata")
	}
	err = os.WriteFile(s.pathMetadata(snapshot.Height, snapshot.Format), value, 0o600)
	if err != nil {
		return errors.Wrap(err, "failed to write snapshot metadata")
	}
	return nil
}

// pathHeight generates the path to a height, containing multiple snapshot formats.
func (s *Store) pathHeight(height uint64) string {
	return filepath.Join(s.dir, strconv.FormatUint(height, 10))
}

// pathSnapshot generates a snapshot path, as a specific format under a height.
func (s *Store) pathSnapshot(height uint64, format uint32) string {
	return filepath.Join(s.pathHeight(height), strconv.FormatUint(uint64(format), 10))
}

func (s *Store) pathMetadataDir() string {
	return filepath.Join(s.dir, "metadata")
}

// pathMetadata generates a snapshot metadata path.
func (s *Store) pathMetadata(height uint64, format uint32) string {
	return filepath.Join(s.pathMetadataDir(), fmt.Sprintf("%020d-%08d", height, format))
}

// PathChunk generates a snapshot chunk path.
func (s *Store) PathChunk(height uint64, format, chunk uint32) string {
	return filepath.Join(s.pathSnapshot(height, format), strconv.FormatUint(uint64(chunk), 10))
}

func (s *Store) parseMetadataFilename(filename string) (height uint64, format uint32, err error) {
	parts := strings.Split(filename, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid snapshot metadata filename %s", filename)
	}
	height, err = strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "invalid snapshot metadata filename %s", filename)
	}
	var f uint64
	f, err = strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "invalid snapshot metadata filename %s", filename)
	}
	format = uint32(f)
	if filename != filepath.Base(s.pathMetadata(height, format)) {
		return 0, 0, fmt.Errorf("invalid snapshot metadata filename %s", filename)
	}
	return height, format, nil
}

func (s *Store) validateMetadataPath(path string) error {
	dir, f := filepath.Split(path)
	if dir != fmt.Sprintf("%s/", s.pathMetadataDir()) {
		return fmt.Errorf("invalid snapshot metadata path %s", path)
	}
	_, _, err := s.parseMetadataFilename(f)
	return err
}

// legacyV1DecodeKey decodes a legacy snapshot key used in a raw kv store.
func legacyV1DecodeKey(k []byte) (uint64, uint32, error) {
	if len(k) != 13 {
		return 0, 0, errors.Wrapf(storeerrors.ErrLogic, "invalid snapshot key with length %v", len(k))
	}
	if k[0] != keyPrefixSnapshot {
		return 0, 0, errors.Wrapf(storeerrors.ErrLogic, "invalid snapshot key prefix %x", k[0])
	}

	height := binary.BigEndian.Uint64(k[1:9])
	format := binary.BigEndian.Uint32(k[9:13])
	return height, format, nil
}

// legacyV1EncodeKey encodes a snapshot key for use in a raw kv store.
func legacyV1EncodeKey(height uint64, format uint32) []byte {
	k := make([]byte, 13)
	k[0] = keyPrefixSnapshot
	binary.BigEndian.PutUint64(k[1:], height)
	binary.BigEndian.PutUint32(k[9:], format)
	return k
}

func (s *Store) MigrateFromV1(db corestore.KVStore) error {
	itr, err := db.Iterator(legacyV1EncodeKey(0, 0), legacyV1EncodeKey(math.MaxUint64, math.MaxUint32))
	if err != nil {
		return err
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		height, format, err := legacyV1DecodeKey(itr.Key())
		if err != nil {
			return err
		}
		if err := os.WriteFile(s.pathMetadata(height, format), itr.Value(), 0o600); err != nil {
			return errors.Wrapf(err, "failed to write snapshot metadata %q", s.pathMetadata(height, format))
		}
	}
	return nil
}
