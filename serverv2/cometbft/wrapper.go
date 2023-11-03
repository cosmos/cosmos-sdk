package cometbft

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"

	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/serverv2/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

type cometABCIWrapper struct {
	app    types.ProtoApp
	logger log.Logger

	// Comment this code later, now I'm avoiding noise (volatile states)
	checkState           *state
	prepareProposalState *state
	processProposalState *state
	finalizeBlockState   *state

	paramStore ParamStore
}

func NewCometABCIWrapper(app types.ProtoApp, logger log.Logger) abci.Application {
	return &cometABCIWrapper{app: app, logger: logger}
}

func (w *cometABCIWrapper) Info(_ context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	lastCommitID := w.app.CommitMultiStore().LastCommitID()
	appVersion := InitialAppVersion
	if lastCommitID.Version > 0 {
		ctx, err := w.app.CreateQueryContext(lastCommitID.Version, false)
		if err != nil {
			return nil, fmt.Errorf("failed creating query context: %w", err)
		}
		appVersion, err = w.app.AppVersion(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed getting app version: %w", err)
		}
	}

	return &abci.ResponseInfo{
		Data:             w.app.Name(),
		Version:          w.app.Version(),
		AppVersion:       appVersion,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}, nil
}

func (w *cometABCIWrapper) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	return w.app.Query(ctx, req)
}

func (w *cometABCIWrapper) CheckTx(_ context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return w.app.CheckTx(req)
}

func (w *cometABCIWrapper) InitChain(_ context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	if req.ChainId != w.app.ChainID() {
		return nil, fmt.Errorf("invalid chain-id on InitChain; expected: %s, got: %s", w.app.ChainID(), req.ChainId)
	}

	// On a new chain, we consider the init chain block height as 0, even though
	// req.InitialHeight is 1 by default.
	initHeader := cmtproto.Header{ChainID: req.ChainId, Time: req.Time}
	w.logger.Info("InitChain", "initialHeight", req.InitialHeight, "chainID", req.ChainId)

	// Set the initial height, which will be used to determine if we are proposing
	// or processing the first block or not.
	w.app.SetInitialHeight(req.InitialHeight)
	if w.app.InitialHeight() == 0 { // If initial height is 0, set it to 1
		w.app.SetInitialHeight(1)
	}

	// if req.InitialHeight is > 1, then we set the initial version on all stores
	if req.InitialHeight > 1 {
		initHeader.Height = req.InitialHeight
		if err := w.app.CommitMultiStore().SetInitialVersion(req.InitialHeight); err != nil {
			return nil, err
		}
	}

	// initialize states with a correct header
	w.setState(execModeFinalize, initHeader)
	w.setState(execModeCheck, initHeader)

	// Store the consensus params in the BaseApp's param store. Note, this must be
	// done after the finalizeBlockState and context have been set as it's persisted
	// to state.
	if req.ConsensusParams != nil {
		err := w.StoreConsensusParams(w.finalizeBlockState.ctx, *req.ConsensusParams)
		if err != nil {
			return nil, err
		}
	}

	defer func() {
		// InitChain represents the state of the application BEFORE the first block,
		// i.e. the genesis block. This means that when processing the app's InitChain
		// handler, the block height is zero by default. However, after Commit is called
		// the height needs to reflect the true block height.
		initHeader.Height = req.InitialHeight
		w.checkState.ctx = w.checkState.ctx.WithBlockHeader(initHeader).
			WithHeaderInfo(coreheader.Info{
				ChainID: req.ChainId,
				Height:  req.InitialHeight,
				Time:    req.Time,
			})
		w.finalizeBlockState.ctx = w.finalizeBlockState.ctx.WithBlockHeader(initHeader).
			WithHeaderInfo(coreheader.Info{
				ChainID: req.ChainId,
				Height:  req.InitialHeight,
				Time:    req.Time,
			})
	}()

	// TODO: define this in the app
	if w.app.InitChainer() == nil {
		return &abci.ResponseInitChain{}, nil
	}

	// add block gas meter for any genesis transactions (allow infinite gas)
	w.finalizeBlockState.ctx = w.finalizeBlockState.ctx.WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	// TODO: define this in the app
	res, err := w.app.InitChainer()(w.finalizeBlockState.ctx, req)
	if err != nil {
		return nil, err
	}

	if len(req.Validators) > 0 {
		if len(req.Validators) != len(res.Validators) {
			return nil, fmt.Errorf(
				"len(RequestInitChain.Validators) != len(GenesisValidators) (%d != %d)",
				len(req.Validators), len(res.Validators),
			)
		}

		sort.Sort(abci.ValidatorUpdates(req.Validators))
		sort.Sort(abci.ValidatorUpdates(res.Validators))

		for i := range res.Validators {
			if !proto.Equal(&res.Validators[i], &req.Validators[i]) {
				return nil, fmt.Errorf("genesisValidators[%d] != req.Validators[%d] ", i, i)
			}
		}
	}

	// In the case of a new chain, AppHash will be the hash of an empty string.
	// During an upgrade, it'll be the hash of the last committed block.
	var appHash []byte
	if !w.app.CommitMultiStore().LastCommitID().IsZero() {
		appHash = w.app.CommitMultiStore().LastCommitID().Hash
	} else {
		// $ echo -n '' | sha256sum
		// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		emptyHash := sha256.Sum256([]byte{})
		appHash = emptyHash[:]
	}

	// NOTE: We don't commit, but FinalizeBlock for block InitialHeight starts from
	// this FinalizeBlockState.
	return &abci.ResponseInitChain{
		ConsensusParams: res.ConsensusParams,
		Validators:      res.Validators,
		AppHash:         appHash,
	}, nil
}

func (w *cometABCIWrapper) PrepareProposal(_ context.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	return w.app.PrepareProposal(req)
}

func (w *cometABCIWrapper) ProcessProposal(_ context.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	return w.app.ProcessProposal(req)
}

func (w *cometABCIWrapper) FinalizeBlock(_ context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return w.app.FinalizeBlock(req)
}

func (w *cometABCIWrapper) ExtendVote(ctx context.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	return w.app.ExtendVote(ctx, req)
}

func (w *cometABCIWrapper) VerifyVoteExtension(_ context.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	return w.app.VerifyVoteExtension(req)
}

func (w *cometABCIWrapper) Commit(_ context.Context, _ *abci.RequestCommit) (*abci.ResponseCommit, error) {
	return w.app.Commit()
}

func (w *cometABCIWrapper) ListSnapshots(_ context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	resp := &abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{}}
	if w.app.SnapshotManager() == nil {
		return resp, nil
	}

	snapshots, err := w.app.SnapshotManager().List()
	if err != nil {
		w.logger.Error("failed to list snapshots", "err", err)
		return nil, err
	}

	for _, snapshot := range snapshots {
		abciSnapshot, err := snapshot.ToABCI()
		if err != nil {
			w.logger.Error("failed to convert ABCI snapshots", "err", err)
			return nil, err
		}

		resp.Snapshots = append(resp.Snapshots, &abciSnapshot)
	}

	return resp, nil
}

func (w *cometABCIWrapper) OfferSnapshot(_ context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	if w.app.SnapshotManager() == nil {
		w.logger.Error("snapshot manager not configured")
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}, nil
	}

	if req.Snapshot == nil {
		w.logger.Error("received nil snapshot")
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil
	}

	// TODO: SnapshotFromABCI should be moved to this package or out of the SDK
	snapshot, err := snapshottypes.SnapshotFromABCI(req.Snapshot)
	if err != nil {
		w.logger.Error("failed to decode snapshot metadata", "err", err)
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil
	}

	err = w.app.SnapshotManager().Restore(snapshot)
	switch {
	case err == nil:
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrUnknownFormat):
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT_FORMAT}, nil

	case errors.Is(err, snapshottypes.ErrInvalidMetadata):
		w.logger.Error(
			"rejecting invalid snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil

	default:
		w.logger.Error(
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

func (w *cometABCIWrapper) LoadSnapshotChunk(_ context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	if w.app.SnapshotManager() == nil {
		return &abci.ResponseLoadSnapshotChunk{}, nil
	}

	chunk, err := w.app.SnapshotManager().LoadChunk(req.Height, req.Format, req.Chunk)
	if err != nil {
		w.logger.Error(
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

func (w *cometABCIWrapper) ApplySnapshotChunk(_ context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	if w.app.SnapshotManager() == nil {
		w.logger.Error("snapshot manager not configured")
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
	}

	_, err := w.app.SnapshotManager().RestoreChunk(req.Chunk)
	switch {
	case err == nil:
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrChunkHashMismatch):
		w.logger.Error(
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
		w.logger.Error("failed to restore snapshot", "err", err)
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
	}
}

func (w *cometABCIWrapper) StoreConsensusParams(ctx sdk.Context, cp cmtproto.ConsensusParams) error {
	if w.paramStore == nil {
		return errors.New("cannot store consensus params with no params store set")
	}

	return w.paramStore.Set(ctx, cp)
}

// TODO: maybe we can avoid this, I had some ideas about it
// setState sets the BaseApp's state for the corresponding mode with a branched
// multi-store (i.e. a CacheMultiStore) and a new Context with the same
// multi-store branch, and provided header.
func (w *cometABCIWrapper) setState(mode execMode, header cmtproto.Header) {
	ms := w.app.CommitMultiStore().CacheMultiStore()
	baseState := &state{
		ms:  ms,
		ctx: sdk.NewContext(ms, false, w.logger).WithStreamingManager(w.app.StreamingManager()).WithBlockHeader(header),
	}

	switch mode {
	case execModeCheck:
		baseState.ctx = baseState.ctx.WithIsCheckTx(true).WithMinGasPrices(w.app.MinGasPrices())
		w.checkState = baseState

	case execModePrepareProposal:
		w.prepareProposalState = baseState

	case execModeProcessProposal:
		w.processProposalState = baseState

	case execModeFinalize:
		w.finalizeBlockState = baseState

	default:
		panic(fmt.Sprintf("invalid runTxMode for setState: %d", mode))
	}
}
