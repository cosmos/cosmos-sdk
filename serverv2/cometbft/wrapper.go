package cometbft

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"

	corecomet "cosmossdk.io/core/comet"
	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/serverv2/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

type cometABCIWrapper struct {
	app    types.ProtoApp
	logger log.Logger

	// indexEvents defines the set of events in the form {eventType}.{attributeKey},
	// which informs CometBFT what to index. If empty, all events will be indexed.
	indexEvents map[string]struct{}

	paramStore ParamStore

	// TODO: these below here I don't know yet where to put
	trace bool
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
	// TODO: should we do something different for check and re-check tx?
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

	ms := w.app.CommitMultiStore().CacheMultiStore()
	ctx := sdk.NewContext(ms, false, w.logger).WithStreamingManager(w.app.StreamingManager()).WithBlockHeader(initHeader)
	defer ms.Write() // no need to keep InitChain changes in cache, we can write them immediately

	// Store the consensus params in the BaseApp's param store. Note, this must be
	// done after the finalizeBlockState and context have been set as it's persisted
	// to state.
	if req.ConsensusParams != nil {
		err := w.StoreConsensusParams(ctx, *req.ConsensusParams)
		if err != nil {
			return nil, err
		}
	}

	// TODO: define this in the app
	if w.app.InitChainer() == nil {
		return &abci.ResponseInitChain{}, nil
	}

	// add block gas meter for any genesis transactions (allow infinite gas)
	ctx = ctx.WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	// TODO: define this in the app
	res, err := w.app.InitChainer()(ctx, req)
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
	// TODO: do an interface check to see if app implements PrepareProposal
	return w.app.PrepareProposal(req)
}

func (w *cometABCIWrapper) ProcessProposal(_ context.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	// TODO: do an interface check to see if app implements ProcessProposal
	return w.app.ProcessProposal(req)
}

func (w *cometABCIWrapper) FinalizeBlock(c context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	var events []abci.Event

	if err := w.app.CheckHalt(req.Height, req.Time); err != nil {
		return nil, err
	}

	if err := w.validateFinalizeBlockHeight(req); err != nil {
		return nil, err
	}

	if w.app.CommitMultiStore().TracingEnabled() {
		w.app.CommitMultiStore().SetTracingContext(storetypes.TraceContext(
			map[string]any{"blockHeight": req.Height},
		))
	}

	header := cmtproto.Header{
		ChainID:            w.app.ChainID(),
		Height:             req.Height,
		Time:               req.Time,
		ProposerAddress:    req.ProposerAddress,
		NextValidatorsHash: req.NextValidatorsHash,
		AppHash:            w.app.CommitMultiStore().LastCommitID().Hash,
	}

	ms := w.app.CommitMultiStore().CacheMultiStore()

	// Context is now updated with Header information.
	ctx := sdk.NewContext(ms, false, w.logger).
		WithContext(c).
		WithStreamingManager(w.app.StreamingManager()).
		WithBlockHeader(header).
		WithHeaderHash(req.Hash).
		WithHeaderInfo(coreheader.Info{
			ChainID: w.app.ChainID(),
			Height:  req.Height,
			Time:    req.Time,
			Hash:    req.Hash,
			AppHash: w.app.CommitMultiStore().LastCommitID().Hash,
		}).
		WithVoteInfos(req.DecidedLastCommit.Votes).
		WithExecMode(sdk.ExecModeFinalize).
		WithCometInfo(corecomet.Info{
			Evidence:        sdk.ToSDKEvidence(req.Misbehavior),
			ValidatorsHash:  req.NextValidatorsHash,
			ProposerAddress: req.ProposerAddress,
			LastCommit:      sdk.ToSDKCommitInfo(req.DecidedLastCommit),
		})

		// GasMeter must be set after we get a context with updated consensus params.
	ctx = ctx.WithConsensusParams(w.GetConsensusParams(ctx)).
		WithBlockGasMeter(w.getBlockGasMeter(ctx))

	// if w.checkState != nil {
	// 	w.checkState.ctx = w.checkState.ctx.
	// 		WithBlockGasMeter(gasMeter).
	// 		WithHeaderHash(req.Hash)
	// }

	if err := w.app.PreBlock(req); err != nil {
		return nil, err
	}

	beginBlock, err := w.app.BeginBlock(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: Right now BeginBlock in baseapp is gathering events, maybe do it here idk

	// First check for an abort signal after beginBlock, as it's the first place
	// we spend any significant amount of time.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// continue
	}

	events = append(events, beginBlock.Events...)

	// Iterate over all raw transactions in the proposal and attempt to execute
	// them, gathering the execution results.
	//
	// NOTE: Not all raw transactions may adhere to the sdk.Tx interface, e.g.
	// vote extensions, so skip those.
	txResults := make([]*abci.ExecTxResult, 0, len(req.Txs))
	for _, rawTx := range req.Txs {
		var response *abci.ExecTxResult

		// TODO: figure out if ValidateTX should be used here or not?
		if _, _, _, err := w.app.ValidateTX(rawTx); err == nil {
			response = w.app.DeliverTx(rawTx)
		} else {
			// In the case where a transaction included in a block proposal is malformed,
			// we still want to return a default response to comet. This is because comet
			// expects a response for each transaction included in a block proposal.
			response = sdkerrors.ResponseExecTxResultWithEvents(
				sdkerrors.ErrTxDecode,
				0,
				0,
				nil,
				false,
			)
		}

		// check after every tx if we should abort
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// continue
		}

		txResults = append(txResults, response)
	}

	if ms.TracingEnabled() {
		ms = ms.SetTracingContext(nil).(storetypes.CacheMultiStore)
	}

	endBlock, err := w.app.EndBlock(ctx)
	if err != nil {
		return nil, err
	}

	// check after endBlock if we should abort, to avoid propagating the result
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// continue
	}

	events = append(events, endBlock.Events...)
	cp := w.GetConsensusParams(ctx)

	// write changes into the branched state and get working hash
	ms.Write()
	commitHash := w.app.CommitMultiStore().WorkingHash()
	w.logger.Debug("hash of all writes", "workingHash", fmt.Sprintf("%X", commitHash))

	return &abci.ResponseFinalizeBlock{
		Events:                events,
		TxResults:             txResults,
		ValidatorUpdates:      endBlock.ValidatorUpdates,
		ConsensusParamUpdates: &cp,
		AppHash:               commitHash,
	}, nil
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

	lastBlockHeight := w.app.CommitMultiStore().LastCommitID().Version

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

func (w *cometABCIWrapper) getBlockGasMeter(ctx sdk.Context) storetypes.GasMeter {
	if maxGas := w.app.GetMaximumBlockGas(ctx); maxGas > 0 {
		return storetypes.NewGasMeter(maxGas)
	}

	return storetypes.NewInfiniteGasMeter()
}
