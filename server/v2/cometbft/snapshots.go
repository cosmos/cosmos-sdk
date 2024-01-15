package cometbft

import (
	"context"
	"errors"

	snapshottypes "cosmossdk.io/store/v2/snapshots/types"
	abci "github.com/cometbft/cometbft/abci/types"
)

// ApplySnapshotChunk implements types.Application.
func (c *Consensus[T]) ApplySnapshotChunk(_ context.Context, req *abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error) {
	if c.cfg.SnapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.ApplySnapshotChunkResponse{Result: abci.APPLY_SNAPSHOT_CHUNK_RESULT_ABORT}, nil
	}

	_, err := c.cfg.SnapshotManager.RestoreChunk(req.Chunk)
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
func (c *Consensus[T]) ListSnapshots(_ context.Context, ctx *abci.ListSnapshotsRequest) (resp *abci.ListSnapshotsResponse, err error) {
	if c.cfg.SnapshotManager == nil {
		return resp, nil
	}

	snapshots, err := c.cfg.SnapshotManager.List()
	if err != nil {
		c.logger.Error("failed to list snapshots", "err", err)
		return nil, err
	}

	for _, snapshot := range snapshots {
		abciSnapshot, err := snapshot.ToABCI()
		if err != nil {
			c.logger.Error("failed to convert ABCI snapshots", "err", err)
			return nil, err
		}

		resp.Snapshots = append(resp.Snapshots, &abciSnapshot)
	}

	return resp, nil
}

// LoadSnapshotChunk implements types.Application.
func (c *Consensus[T]) LoadSnapshotChunk(_ context.Context, req *abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error) {
	if c.cfg.SnapshotManager == nil {
		return &abci.LoadSnapshotChunkResponse{}, nil
	}

	chunk, err := c.cfg.SnapshotManager.LoadChunk(req.Height, req.Format, req.Chunk)
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
func (c *Consensus[T]) OfferSnapshot(_ context.Context, req *abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error) {
	if c.cfg.SnapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_ABORT}, nil
	}

	if req.Snapshot == nil {
		c.logger.Error("received nil snapshot")
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT}, nil
	}

	snapshot, err := snapshottypes.SnapshotFromABCI(req.Snapshot)
	if err != nil {
		c.logger.Error("failed to decode snapshot metadata", "err", err)
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT}, nil
	}

	err = c.cfg.SnapshotManager.Restore(snapshot)
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
