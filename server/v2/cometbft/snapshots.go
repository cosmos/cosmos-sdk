package cometbft

import (
	"context"
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"

	snapshottypes "cosmossdk.io/store/v2/snapshots/types"
)

// ApplySnapshotChunk implements types.Application.
func (c *Consensus[T]) ApplySnapshotChunk(_ context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	if c.cfg.SnapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
	}

	_, err := c.cfg.SnapshotManager.RestoreChunk(req.Chunk)
	switch {
	case err == nil:
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrChunkHashMismatch):
		c.logger.Error(
			"chunk checksum mismatch; rejecting sender and requesting refetch",
			"chunk", req.Index,
			"sender", req.Sender,
			"err", err,
		)
		return &abci.ResponseApplySnapshotChunk{
			Result:        abci.ResponseApplySnapshotChunk_RETRY,
			RefetchChunks: []uint32{req.Index},
			RejectSenders: []string{req.Sender},
		}, nil

	default:
		c.logger.Error("failed to restore snapshot", "err", err)
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
	}
}

// ListSnapshots implements types.Application.
func (c *Consensus[T]) ListSnapshots(_ context.Context, ctx *abci.RequestListSnapshots) (resp *abci.ResponseListSnapshots, err error) {
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
func (c *Consensus[T]) LoadSnapshotChunk(_ context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	if c.cfg.SnapshotManager == nil {
		return &abci.ResponseLoadSnapshotChunk{}, nil
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

	return &abci.ResponseLoadSnapshotChunk{Chunk: chunk}, nil
}

// OfferSnapshot implements types.Application.
func (c *Consensus[T]) OfferSnapshot(_ context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	if c.cfg.SnapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}, nil
	}

	if req.Snapshot == nil {
		c.logger.Error("received nil snapshot")
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil
	}

	snapshot, err := snapshottypes.SnapshotFromABCI(req.Snapshot)
	if err != nil {
		c.logger.Error("failed to decode snapshot metadata", "err", err)
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil
	}

	err = c.cfg.SnapshotManager.Restore(snapshot)
	switch {
	case err == nil:
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrUnknownFormat):
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT_FORMAT}, nil

	case errors.Is(err, snapshottypes.ErrInvalidMetadata):
		c.logger.Error(
			"rejecting invalid snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil

	default:
		c.logger.Error(
			"failed to restore snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)

		// We currently don't support resetting the IAVL stores and retrying a
		// different snapshot, so we ask CometBFT to abort all snapshot restoration.
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}, nil
	}
}
