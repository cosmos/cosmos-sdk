package cometbft

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/store/v2/snapshots"
	snapshottypes "cosmossdk.io/store/v2/snapshots/types"
)

// GetSnapshotStore returns a snapshot store for the given application options.
// It creates a directory for storing snapshots if it doesn't exist.
// It initializes a GoLevelDB database for storing metadata of the snapshots.
// The snapshot store is then created using the initialized database and directory.
// If any error occurs during the process, it is returned along with a nil snapshot store.
func GetSnapshotStore(rootDir string) (*snapshots.Store, error) {
	snapshotDir := filepath.Join(rootDir, "data", "snapshots")
	if err := os.MkdirAll(snapshotDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	snapshotStore, err := snapshots.NewStore(snapshotDir)
	if err != nil {
		return nil, err
	}

	return snapshotStore, nil
}

// ApplySnapshotChunk implements types.Application.
func (c *consensus[T]) ApplySnapshotChunk(_ context.Context, req *abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error) {
	if c.snapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.ApplySnapshotChunkResponse{Result: abci.APPLY_SNAPSHOT_CHUNK_RESULT_ABORT}, nil
	}

	_, err := c.snapshotManager.RestoreChunk(req.Chunk)
	switch {
	case err == nil:
		return &abci.ApplySnapshotChunkResponse{Result: abci.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrChunkHashMismatch):
		c.logger.Error(
			"chunk checksum mismatch; rejecting sender and requesting refetch",
			"chunk", req.Index,
			"sender", req.Sender,
			"err", err,
		)
		return &abci.ApplySnapshotChunkResponse{
			Result:        abci.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY,
			RefetchChunks: []uint32{req.Index},
			RejectSenders: []string{req.Sender},
		}, nil

	default:
		c.logger.Error("failed to restore snapshot", "err", err)
		return &abci.ApplySnapshotChunkResponse{Result: abci.APPLY_SNAPSHOT_CHUNK_RESULT_ABORT}, nil
	}
}

// ListSnapshots implements types.Application.
func (c *consensus[T]) ListSnapshots(_ context.Context, ctx *abci.ListSnapshotsRequest) (*abci.ListSnapshotsResponse, error) {
	if c.snapshotManager == nil {
		return nil, nil
	}

	snapshots, err := c.snapshotManager.List()
	if err != nil {
		c.logger.Error("failed to list snapshots", "err", err)
		return nil, err
	}

	resp := &abci.ListSnapshotsResponse{}
	for _, snapshot := range snapshots {
		abciSnapshot, err := snapshotToABCI(snapshot)
		if err != nil {
			c.logger.Error("failed to convert ABCI snapshots", "err", err)
			return nil, err
		}

		resp.Snapshots = append(resp.Snapshots, &abciSnapshot)
	}

	return resp, nil
}

// LoadSnapshotChunk implements types.Application.
func (c *consensus[T]) LoadSnapshotChunk(_ context.Context, req *abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error) {
	if c.snapshotManager == nil {
		return &abci.LoadSnapshotChunkResponse{}, nil
	}

	chunk, err := c.snapshotManager.LoadChunk(req.Height, req.Format, req.Chunk)
	if err != nil {
		c.logger.Error(
			"failed to load snapshot chunk",
			"height", req.Height,
			"format", req.Format,
			"chunk", req.Chunk,
			"err", err,
		)
		return nil, err
	}

	return &abci.LoadSnapshotChunkResponse{Chunk: chunk}, nil
}

// OfferSnapshot implements types.Application.
func (c *consensus[T]) OfferSnapshot(_ context.Context, req *abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error) {
	if c.snapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_ABORT}, nil
	}

	if req.Snapshot == nil {
		c.logger.Error("received nil snapshot")
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT}, nil
	}

	snapshot, err := snapshotFromABCI(req.Snapshot)
	if err != nil {
		c.logger.Error("failed to decode snapshot metadata", "err", err)
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT}, nil
	}

	err = c.snapshotManager.Restore(snapshot)
	switch {
	case err == nil:
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrUnknownFormat):
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT_FORMAT}, nil

	case errors.Is(err, snapshottypes.ErrInvalidMetadata):
		c.logger.Error(
			"rejecting invalid snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT}, nil

	default:
		c.logger.Error(
			"failed to restore snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)

		// We currently don't support resetting the IAVL stores and retrying a
		// different snapshot, so we ask CometBFT to abort all snapshot restoration.
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_ABORT}, nil
	}
}

// Converts an ABCI snapshot to a snapshot. Mainly to decode the SDK metadata.
func snapshotFromABCI(in *abci.Snapshot) (snapshottypes.Snapshot, error) {
	snapshot := snapshottypes.Snapshot{
		Height: in.Height,
		Format: in.Format,
		Chunks: in.Chunks,
		Hash:   in.Hash,
	}
	err := proto.Unmarshal(in.Metadata, &snapshot.Metadata)
	if err != nil {
		return snapshottypes.Snapshot{}, fmt.Errorf("failed to unmarshal snapshot metadata: %w", err)
	}
	return snapshot, nil
}

// Converts a Snapshot to its ABCI representation. Mainly to encode the SDK metadata.
func snapshotToABCI(s *snapshottypes.Snapshot) (abci.Snapshot, error) {
	out := abci.Snapshot{
		Height: s.Height,
		Format: s.Format,
		Chunks: s.Chunks,
		Hash:   s.Hash,
	}
	var err error
	out.Metadata, err = proto.Marshal(&s.Metadata)
	if err != nil {
		return abci.Snapshot{}, fmt.Errorf("failed to marshal snapshot metadata: %w", err)
	}
	return out, nil
}
