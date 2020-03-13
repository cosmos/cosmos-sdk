package snapshot

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	db "github.com/tendermint/tm-db"
)

// Store is a snapshot store, containing snapshot metadata and chunks.
type Store struct {
	db  db.DB
	dir string
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
		db:  db,
		dir: dir,
	}, nil
}

// Delete deletes a snapshot
func (s *Store) Delete(height uint64, format uint32) error {
	return nil
}

// Load loads a snapshot from disk
func (s *Store) Load(height uint64, format uint32) (*store.Snapshot, error) {
	return nil, nil
}

// Save saves a snapshot to disk
func (s *Store) Save(snapshot *store.Snapshot) error {
	if snapshot.Height == 0 {
		return errors.New("snapshot height cannot be 0")
	}
	current, err := s.Load(snapshot.Height, snapshot.Format)
	if err != nil {
		return err
	}
	if current != nil {
		_ = current.Close()
		return fmt.Errorf("snapshot already exists for height %v format %v",
			snapshot.Height, snapshot.Format)
	}

	metadata := &types.SnapshotMetadata{
		Height: snapshot.Height,
		Format: snapshot.Format,
	}
	i := uint32(1)
	for chunk := range snapshot.Chunks {
		defer chunk.Close()
		snapshotDir := filepath.Join(s.dir, strconv.FormatUint(snapshot.Height, 10))
		err = os.MkdirAll(snapshotDir, 0755)
		if err != nil {
			return err
		}
		file, err := os.Create(filepath.Join(snapshotDir, strconv.FormatUint(uint64(i), 10)))
		if err != nil {
			return err
		}
		defer file.Close()
		hasher := sha1.New()
		_, err = io.Copy(io.MultiWriter(file, hasher), chunk)
		if err != nil {
			return fmt.Errorf("failed to read snapshot chunk %v: %w", i, err)
		}
		err = file.Close()
		if err != nil {
			return err
		}
		err = chunk.Close()
		if err != nil {
			return err
		}
		metadata.Chunks = append(metadata.Chunks, &types.SnapshotChunkMetadata{
			Chunk:    i,
			Checksum: hasher.Sum(nil),
		})
		i++
	}

	buf := proto.NewBuffer(nil)
	err = buf.EncodeMessage(metadata)
	if err != nil {
		return fmt.Errorf("failed to encode snapshot: %w", err)
	}
	err = s.db.Set(key(snapshot.Height, snapshot.Format), buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to store snapshot: %w", err)
	}
	return nil
}

// key generates a snapshot key
// FIXME We should probably generate this in a different way.
func key(height uint64, format uint32) []byte {
	k := make([]byte, 0, 13)
	k = append(k, 0x01) // prefix for snapshot metadata

	bHeight := make([]byte, 8)
	binary.BigEndian.PutUint64(bHeight, height)
	k = append(k, bHeight...)

	bFormat := make([]byte, 4)
	binary.BigEndian.PutUint32(bFormat, format)
	k = append(k, bFormat...)

	return k
}
