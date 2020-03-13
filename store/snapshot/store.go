package snapshot

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	db "github.com/tendermint/tm-db"
)

const (
	keyPrefixSnapshot = 0x01
)

// Store is a snapshot store, containing snapshot metadata and chunks.
type Store struct {
	db  db.DB
	dir string

	mtx    sync.Mutex
	saving map[uint64]bool // heights currently being saved
}

// New creates a new snapshot store. The passed database must be independent of the application
// database, to prevent it from taking snapshots of itself.
func New(db db.DB, dir string) (*Store, error) {
	if dir == "" {
		return nil, errors.New("snapshot directory not given")
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot dir %v: %w", dir, err)
	}

	return &Store{
		db:     db,
		dir:    dir,
		saving: make(map[uint64]bool),
	}, nil
}

// Active checks whether there are currently any active snapshots being saved
func (s *Store) Active() bool {
	s.mtx.Lock()
	active := false
	for _, saving := range s.saving {
		if saving {
			active = true
			break
		}
	}
	s.mtx.Unlock()
	return active
}

// Delete deletes a snapshot
func (s *Store) Delete(height uint64, format uint32) error {
	return nil
}

// Load loads a snapshot (both metadata and chunks). The chunks must be consumed and closed.
func (s *Store) Load(height uint64, format uint32) (*types.SnapshotMetadata, <-chan io.ReadCloser, error) {
	metadata, err := s.LoadMetadata(height, format)
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan io.ReadCloser)
	go func() {
		defer close(ch)
		for _, chunkMetadata := range metadata.Chunks {
			pr, pw := io.Pipe()
			ch <- pr
			hasher := sha1.New()
			chunk, err := s.LoadChunk(height, format, chunkMetadata.Chunk)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			defer chunk.Close()
			_, err = io.Copy(io.MultiWriter(pw, hasher), chunk)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			if !bytes.Equal(chunkMetadata.Checksum, hasher.Sum(nil)) {
				pw.CloseWithError(fmt.Errorf("checksum failure for chunk %v: expected %q got %q",
					chunkMetadata.Chunk, chunkMetadata.Checksum, hasher.Sum(nil)))
				return
			}
			chunk.Close()
			pw.Close()
		}
	}()

	return metadata, ch, nil
}

// LoadMetadata loads snapshot metadata from the database.
func (s *Store) LoadMetadata(height uint64, format uint32) (*types.SnapshotMetadata, error) {
	bytes, err := s.db.Get(key(height, format))
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshot metadata: %w", err)
	}
	if bytes == nil {
		return nil, fmt.Errorf("snapshot at height %v in format %v not found", height, format)
	}
	metadata := &types.SnapshotMetadata{}
	err = proto.Unmarshal(bytes, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshot metadata for height %v format %v: %w",
			height, format, err)
	}
	return metadata, nil
}

// LoadChunk loads a chunk from disk. The caller must call Close() on it when done.
func (s *Store) LoadChunk(height uint64, format uint32, chunk uint32) (io.ReadCloser, error) {
	metadata, err := s.LoadMetadata(height, format)
	if err != nil {
		return nil, err
	}
	if chunk > uint32(len(metadata.Chunks)) {
		return nil, fmt.Errorf("snapshot for height %v format %v only has %v chunks, requested %v",
			height, format, len(metadata.Chunks), chunk)
	}
	path := s.pathChunk(height, format, chunk)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch chunk %v from file %q: %w", chunk, path, err)
	}
	return file, nil
}

// Save saves a snapshot to disk
func (s *Store) Save(height uint64, format uint32, chunks <-chan io.ReadCloser) error {
	// Make sure we close all of the chunks on error
	defer func() {
		for c := range chunks {
			c.Close()
		}
	}()
	if height == 0 {
		return errors.New("snapshot height cannot be 0")
	}

	s.mtx.Lock()
	saving := s.saving[height]
	s.saving[height] = true
	s.mtx.Unlock()
	if saving {
		return fmt.Errorf("a snapshot for height %v is already being saved", height)
	}
	defer func() {
		s.mtx.Lock()
		delete(s.saving, height)
		s.mtx.Unlock()
	}()

	exists, err := s.db.Has(key(height, format))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("snapshot already exists for height %v format %v", height, format)
	}

	metadata := &types.SnapshotMetadata{
		Height: height,
		Format: format,
	}
	index := uint32(1)
	for chunk := range chunks {
		chunkMetadata, err := s.saveChunk(height, format, index, chunk)
		if err != nil {
			return err
		}
		metadata.Chunks = append(metadata.Chunks, chunkMetadata)
		index++
	}
	return s.saveMetadata(height, format, metadata)
}

// saveChunk saves a chunk to disk
func (s *Store) saveChunk(height uint64, format uint32, index uint32, chunk io.ReadCloser) (*types.SnapshotChunkMetadata, error) {
	defer chunk.Close()
	dir := s.pathSnapshot(height, format)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory %q: %w", dir, err)
	}
	path := s.pathChunk(height, format, index)
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot chunk file %q: %w", path, err)
	}
	defer file.Close()
	hasher := sha1.New()
	_, err = io.Copy(io.MultiWriter(file, hasher), chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to generate snapshot chunk file %q: %w", path, err)
	}
	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close snapshot chunk file %q: %w", path, err)
	}
	return &types.SnapshotChunkMetadata{
		Chunk:    index,
		Checksum: hasher.Sum(nil),
	}, nil
}

// saveMetadata saves snapshot metadata to the database
func (s *Store) saveMetadata(height uint64, format uint32, metadata *types.SnapshotMetadata) error {
	value, err := proto.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to encode snapshot metadata: %w", err)
	}
	err = s.db.Set(key(height, format), value)
	if err != nil {
		return fmt.Errorf("failed to store snapshot: %w", err)
	}
	return nil
}

// pathChunk generates a snapshot chunk path
func (s *Store) pathChunk(height uint64, format uint32, chunk uint32) string {
	return filepath.Join(s.pathSnapshot(height, format), strconv.FormatUint(uint64(chunk), 10))
}

// pathSnapshot generates a snapshot path
func (s *Store) pathSnapshot(height uint64, format uint32) string {
	return filepath.Join(s.dir, strconv.FormatUint(height, 10), strconv.FormatUint(uint64(format), 10))
}

// key generates a snapshot key
func key(height uint64, format uint32) []byte {
	k := make([]byte, 0, 13)
	k = append(k, keyPrefixSnapshot)

	bHeight := make([]byte, 8)
	binary.BigEndian.PutUint64(bHeight, height)
	k = append(k, bHeight...)

	bFormat := make([]byte, 4)
	binary.BigEndian.PutUint32(bFormat, format)
	k = append(k, bFormat...)

	return k
}
