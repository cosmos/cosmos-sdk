package cometbft

import (
	"context"
	"errors"
	"fmt"
	"sort"

	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/serverv2/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

var _ abci.Application = (*cometABCIWrapper)(nil)

type cometABCIWrapper struct {
	app    types.ProtoApp
	logger log.Logger

	// indexEvents defines the set of events in the form {eventType}.{attributeKey},
	// which informs CometBFT what to index. If empty, all events will be indexed.
	indexEvents map[string]struct{}

	paramStore ParamStore

	defaultProposalHandler *baseapp.DefaultProposalHandler // move this struct to this package

	// TODO: these below here I don't know yet where to put
	trace bool

	snapshotManager *snapshots.Manager
}

func NewCometABCIWrapper(app types.ProtoApp, logger log.Logger) abci.Application {
	return &cometABCIWrapper{app: app, logger: logger}
}

func (w *cometABCIWrapper) Info(_ context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	appVersion, err := w.app.AppVersion() // avoid the QueryContext given that we are always returning the latest here
	if err != nil {
		return nil, fmt.Errorf("failed getting app version: %w", err)
	}

	return &abci.ResponseInfo{
		Data:             w.app.Name(),
		Version:          w.app.Version(),
		AppVersion:       appVersion,
		LastBlockHeight:  w.app.LastBlockHeight(),
		LastBlockAppHash: w.app.AppHash(),
	}, nil
}

func (w *cometABCIWrapper) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	// return w.app.Query(ctx, req)
	// TODO: handle query
	return &abci.ResponseQuery{}, nil
}

func (w *cometABCIWrapper) CheckTx(_ context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	gInfo, result, anteEvents, err := w.app.ValidateTX(req.Tx)
	if err != nil {
		return sdkerrors.ResponseCheckTxWithEvents(err, gInfo.GasWanted, gInfo.GasUsed, anteEvents, w.trace), nil
	}

	return &abci.ResponseCheckTx{
		GasWanted: int64(gInfo.GasWanted),
		GasUsed:   int64(gInfo.GasUsed),
		Log:       result.Log,
		Data:      result.Data,
		Events:    sdk.MarkEventsToIndex(result.Events, w.indexEvents), // TODO: this event handling should be done on cometbft's package
	}, nil
}

func (w *cometABCIWrapper) InitChain(_ context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	rr := types.RequestInitChain{
		StateBytes: req.AppStateBytes,
	}

	_, err := w.app.InitChain(rr)
	if err != nil {
		return nil, err
	}

	vals := w.app.Validators()

	if len(req.Validators) > 0 {
		if len(req.Validators) != len(vals) {
			return nil, fmt.Errorf(
				"len(RequestInitChain.Validators) != len(GenesisValidators) (%d != %d)",
				len(req.Validators), len(vals),
			)
		}

		sort.Sort(abci.ValidatorUpdates(req.Validators))
		sort.Sort(abci.ValidatorUpdates(vals))

		for i := range vals {
			if !proto.Equal(&vals[i], &req.Validators[i]) {
				return nil, fmt.Errorf("genesisValidators[%d] != req.Validators[%d] ", i, i)
			}
		}
	}

	return &abci.ResponseInitChain{
		ConsensusParams: w.app.ConsensusParams(),
		Validators:      vals,
		AppHash:         w.app.AppHash(),
	}, nil
}

func (w *cometABCIWrapper) PrepareProposal(ctx context.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	if abciapp, ok := w.app.(types.HasProposal); ok {
		return abciapp.PrepareProposal(ctx, req)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return w.defaultProposalHandler.PrepareProposalHandler()(sdkCtx, req)

}

func (w *cometABCIWrapper) ProcessProposal(ctx context.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	if abciapp, ok := w.app.(types.HasProposal); ok {
		return abciapp.ProcessProposal(ctx, req)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return w.defaultProposalHandler.ProcessProposalHandler()(sdkCtx, req)
}

func (w *cometABCIWrapper) FinalizeBlock(c context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	var events []abci.Event

	if err := w.validateFinalizeBlockHeight(req); err != nil {
		return nil, err
	}

	// 	WithVoteInfos(req.DecidedLastCommit.Votes).
	// 	WithExecMode(sdk.ExecModeFinalize).
	// WithCometInfo(corecomet.Info{
	// 	Evidence:        sdk.ToSDKEvidence(req.Misbehavior),
	// 	ValidatorsHash:  req.NextValidatorsHash,
	// 	ProposerAddress: req.ProposerAddress,
	// 	LastCommit:      sdk.ToSDKCommitInfo(req.DecidedLastCommit),
	// })

	// LastCommit is used by upgrade, slashing, distribution, and simulation
	// Evidence is used by x/evidence
	// ValidatorsHash is used by x/staking
	// ProposerAddress is used by x/distribution

	// GasMeter must be set after we get a context with updated consensus params.
	// ctx = ctx.WithConsensusParams(w.GetConsensusParams(ctx)).
	// WithBlockGasMeter(w.getBlockGasMeter(ctx))

	// TODO: missing a way to pass vote infos and the stuff that currently is in comet info

	headerInfo := coreheader.Info{
		ChainID: w.app.ChainID(),
		Height:  req.Height,
		Time:    req.Time,
		Hash:    req.Hash,
		AppHash: w.app.AppHash(),
	}
	_, err := w.app.DeliverTxs(headerInfo, req.Txs)
	if err != nil {
		return nil, err
	}

	cp := w.app.ConsensusParams()

	// TODO: translate tx results

	return &abci.ResponseFinalizeBlock{
		Events: events,
		// TxResults:             txResults,
		ValidatorUpdates:      w.app.Validators(),
		ConsensusParamUpdates: cp,
		AppHash:               w.app.AppHash(),
	}, nil
}

func (w *cometABCIWrapper) ExtendVote(ctx context.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	// TODO: do an interface check to see if app implements ExtendVote
	return &abci.ResponseExtendVote{}, nil
}

func (w *cometABCIWrapper) VerifyVoteExtension(_ context.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	// TODO: do an interface check to see if app implements VerifyVoteExtension
	return &abci.ResponseVerifyVoteExtension{}, nil
}

func (w *cometABCIWrapper) Commit(ctx context.Context, _ *abci.RequestCommit) (*abci.ResponseCommit, error) {
	retainHeight := w.app.GetBlockRetentionHeight(w.app.LastBlockHeight())

	resp := &abci.ResponseCommit{
		RetainHeight: retainHeight,
	}

	err := w.app.Commit()
	if err != nil {
		return nil, err
	}

	// TODO: revise streaming
	// abciListeners := w.app.StreamingManager().ABCIListeners

	// The SnapshotIfApplicable method will create the snapshot by starting the goroutine
	w.snapshotManager.SnapshotIfApplicable(w.app.LastBlockHeight())

	return resp, nil
}

func (w *cometABCIWrapper) ListSnapshots(_ context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	resp := &abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{}}
	if w.snapshotManager == nil {
		return resp, nil
	}

	snapshots, err := w.snapshotManager.List()
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
	if w.snapshotManager == nil {
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

	err = w.snapshotManager.Restore(snapshot)
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
	if w.snapshotManager == nil {
		return &abci.ResponseLoadSnapshotChunk{}, nil
	}

	chunk, err := w.snapshotManager.LoadChunk(req.Height, req.Format, req.Chunk)
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
	if w.snapshotManager == nil {
		w.logger.Error("snapshot manager not configured")
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
	}

	_, err := w.snapshotManager.RestoreChunk(req.Chunk)
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

// StoreConsensusParams sets the consensus parameters to the BaseApp's param
// store.
func (w *cometABCIWrapper) StoreConsensusParams(ctx sdk.Context, cp cmtproto.ConsensusParams) error {
	if w.paramStore == nil {
		return errors.New("cannot store consensus params with no params store set")
	}

	return w.paramStore.Set(ctx, cp)
}

// GetConsensusParams returns the current consensus parameters from the BaseApp's
// ParamStore. If the BaseApp has no ParamStore defined, nil is returned.
func (w *cometABCIWrapper) GetConsensusParams(ctx sdk.Context) cmtproto.ConsensusParams {
	if w.paramStore == nil {
		return cmtproto.ConsensusParams{}
	}

	cp, err := w.paramStore.Get(ctx)
	if err != nil {
		panic(fmt.Errorf("consensus key is nil: %w", err))
	}

	return cp
}

func (w *cometABCIWrapper) validateFinalizeBlockHeight(req *abci.RequestFinalizeBlock) error {
	if req.Height < 1 {
		return fmt.Errorf("invalid height: %d", req.Height)
	}

	lastBlockHeight := w.app.LastBlockHeight()

	// expectedHeight holds the expected height to validate
	var expectedHeight int64
	if lastBlockHeight == 0 && w.app.InitialHeight() > 1 {
		// In this case, we're validating the first block of the chain, i.e no
		// previous commit. The height we're expecting is the initial height.
		expectedHeight = w.app.InitialHeight()
	} else {
		// This case can mean two things:
		//
		// - Either there was already a previous commit in the store, in which
		// case we increment the version from there.
		// - Or there was no previous commit, in which case we start at version 1.
		expectedHeight = lastBlockHeight + 1
	}

	if req.Height != expectedHeight {
		return fmt.Errorf("invalid height: %d; expected: %d", req.Height, expectedHeight)
	}

	return nil
}
