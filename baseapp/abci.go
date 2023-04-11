package baseapp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/rootmulti"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	"github.com/cockroachdb/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Supported ABCI Query prefixes and paths
const (
	QueryPathApp    = "app"
	QueryPathCustom = "custom"
	QueryPathP2P    = "p2p"
	QueryPathStore  = "store"

	QueryPathBroadcastTx = "/cosmos.tx.v1beta1.Service/BroadcastTx"
)

func (app *BaseApp) InitChain(_ context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	if req.ChainId != app.chainID {
		return nil, fmt.Errorf("invalid chain-id on InitChain; expected: %s, got: %s", app.chainID, req.ChainId)
	}

	// On a new chain, we consider the init chain block height as 0, even though
	// req.InitialHeight is 1 by default.
	initHeader := cmtproto.Header{ChainID: req.ChainId, Time: req.Time}

	app.logger.Info("InitChain", "initialHeight", req.InitialHeight, "chainID", req.ChainId)

	// Set the initial height, which will be used to determine if we are proposing
	// or processing the first block or not.
	app.initialHeight = req.InitialHeight

	// if req.InitialHeight is > 1, then we set the initial version on all stores
	if req.InitialHeight > 1 {
<<<<<<< HEAD
		app.initialHeight = req.InitialHeight
		initHeader = cmtproto.Header{ChainID: req.ChainId, Height: req.InitialHeight, Time: req.Time}

		if err := app.cms.SetInitialVersion(req.InitialHeight); err != nil {
			return nil, err
||||||| 401d0d72c9
		app.initialHeight = req.InitialHeight
		initHeader = cmtproto.Header{ChainID: req.ChainId, Height: req.InitialHeight, Time: req.Time}
		err := app.cms.SetInitialVersion(req.InitialHeight)
		if err != nil {
			panic(err)
=======
		initHeader.Height = req.InitialHeight
		if err := app.cms.SetInitialVersion(req.InitialHeight); err != nil {
			panic(err)
>>>>>>> main
		}
	}

	// initialize states with a correct header
	app.setState(execModeFinalize, initHeader)
	app.setState(execModeCheck, initHeader)

	// Store the consensus params in the BaseApp's param store. Note, this must be
	// done after the finalizeBlockState and context have been set as it's persisted
	// to state.
	if req.ConsensusParams != nil {
		err := app.StoreConsensusParams(app.finalizeBlockState.ctx, *req.ConsensusParams)
		if err != nil {
			return nil, err
		}
	}

	if app.initChainer == nil {
		return &abci.ResponseInitChain{}, nil
	}

	// add block gas meter for any genesis transactions (allow infinite gas)
	app.finalizeBlockState.ctx = app.finalizeBlockState.ctx.WithBlockGasMeter(storetypes.NewInfiniteGasMeter())

	res, err := app.initChainer(app.finalizeBlockState.ctx, req)
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
	if !app.LastCommitID().IsZero() {
		appHash = app.LastCommitID().Hash
	} else {
		// $ echo -n '' | sha256sum
		// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		emptyHash := sha256.Sum256([]byte{})
		appHash = emptyHash[:]
	}

	// NOTE: We don't commit, since FinalizeBlock for a block at height
	// initial_height starts from this state.
	return &abci.ResponseInitChain{
		ConsensusParams: res.ConsensusParams,
		Validators:      res.Validators,
		AppHash:         appHash,
	}, nil
}

func (app *BaseApp) Info(_ context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	lastCommitID := app.cms.LastCommitID()

	return &abci.ResponseInfo{
		Data:             app.name,
		Version:          app.version,
		AppVersion:       app.appVersion,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}, nil
}

// Query implements the ABCI interface. It delegates to CommitMultiStore if it
// implements Queryable.
func (app *BaseApp) Query(_ context.Context, req *abci.RequestQuery) (resp *abci.ResponseQuery, err error) {
	// add panic recovery for all queries
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/pull/8039
	defer func() {
		if r := recover(); r != nil {
			resp = sdkerrors.QueryResult(errorsmod.Wrapf(sdkerrors.ErrPanic, "%v", r), app.trace)
		}
	}()

	// when a client did not provide a query height, manually inject the latest
	if req.Height == 0 {
		req.Height = app.LastBlockHeight()
	}

	telemetry.IncrCounter(1, "query", "count")
	telemetry.IncrCounter(1, "query", req.Path)
	defer telemetry.MeasureSince(time.Now(), req.Path)

	if req.Path == QueryPathBroadcastTx {
		return sdkerrors.QueryResult(errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "can't route a broadcast tx message"), app.trace), nil
	}

	// handle gRPC routes first rather than calling splitPath because '/' characters
	// are used as part of gRPC paths
	if grpcHandler := app.grpcQueryRouter.Route(req.Path); grpcHandler != nil {
		return app.handleQueryGRPC(grpcHandler, req), nil
	}

	path := SplitABCIQueryPath(req.Path)
	if len(path) == 0 {
		return sdkerrors.QueryResult(errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "no query path provided"), app.trace), nil
	}

	switch path[0] {
	case QueryPathApp:
		// "/app" prefix for special application queries
		resp = handleQueryApp(app, path, req)

	case QueryPathStore:
		resp = handleQueryStore(app, path, req)

	case QueryPathP2P:
		resp = handleQueryP2P(app, path)

	default:
		resp = sdkerrors.QueryResult(errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "unknown query path"), app.trace)
	}

	return resp, nil
}

// ListSnapshots implements the ABCI interface. It delegates to app.snapshotManager if set.
func (app *BaseApp) ListSnapshots(_ context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	resp := &abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{}}
	if app.snapshotManager == nil {
		return resp, nil
	}

	snapshots, err := app.snapshotManager.List()
	if err != nil {
		app.logger.Error("failed to list snapshots", "err", err)
		return nil, err
	}

	for _, snapshot := range snapshots {
		abciSnapshot, err := snapshot.ToABCI()
		if err != nil {
			app.logger.Error("failed to list snapshots", "err", err)
			return nil, err
		}

		resp.Snapshots = append(resp.Snapshots, &abciSnapshot)
	}

	return resp, nil
}

// LoadSnapshotChunk implements the ABCI interface. It delegates to app.snapshotManager if set.
func (app *BaseApp) LoadSnapshotChunk(_ context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	if app.snapshotManager == nil {
		return &abci.ResponseLoadSnapshotChunk{}, nil
	}

	chunk, err := app.snapshotManager.LoadChunk(req.Height, req.Format, req.Chunk)
	if err != nil {
		app.logger.Error(
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

// OfferSnapshot implements the ABCI interface. It delegates to app.snapshotManager if set.
func (app *BaseApp) OfferSnapshot(_ context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	if app.snapshotManager == nil {
		app.logger.Error("snapshot manager not configured")
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}, nil
	}

	if req.Snapshot == nil {
		app.logger.Error("received nil snapshot")
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil
	}

	snapshot, err := snapshottypes.SnapshotFromABCI(req.Snapshot)
	if err != nil {
		app.logger.Error("failed to decode snapshot metadata", "err", err)
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, err
	}

	// TODO: Examine non-ACCEPT return paths and if errors should be returned or
	// not.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/12272
	err = app.snapshotManager.Restore(snapshot)
	switch {
	case err == nil:
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrUnknownFormat):
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT_FORMAT}, err

	case errors.Is(err, snapshottypes.ErrInvalidMetadata):
		app.logger.Error(
			"rejecting invalid snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, err

	default:
		app.logger.Error(
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

// ApplySnapshotChunk implements the ABCI interface. It delegates to app.snapshotManager if set.
func (app *BaseApp) ApplySnapshotChunk(_ context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	if app.snapshotManager == nil {
		app.logger.Error("snapshot manager not configured")
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
	}

	// TODO: Examine non-ACCEPT return paths and if errors should be returned or
	// not.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/12272
	_, err := app.snapshotManager.RestoreChunk(req.Chunk)
	switch {
	case err == nil:
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrChunkHashMismatch):
		app.logger.Error(
			"chunk checksum mismatch; rejecting sender and requesting refetch",
			"chunk", req.Index,
			"sender", req.Sender,
			"err", err,
		)
		return &abci.ResponseApplySnapshotChunk{
			Result:        abci.ResponseApplySnapshotChunk_RETRY,
			RefetchChunks: []uint32{req.Index},
			RejectSenders: []string{req.Sender},
		}, err

	default:
		app.logger.Error("failed to restore snapshot", "err", err)
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
	}
}

func (app *BaseApp) legacyBeginBlock(req *abci.RequestFinalizeBlock) sdk.LegacyResponseBeginBlock {
	var (
		resp sdk.LegacyResponseBeginBlock
		err  error
	)

	// TODO: Fill out fields...
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/12272
	legacyReq := sdk.LegacyRequestBeginBlock{}

	if app.beginBlocker != nil {
		resp, err = app.beginBlocker(app.finalizeBlockState.ctx, legacyReq)
		if err != nil {
			panic(err)
		}

		resp.Response.Events = sdk.MarkEventsToIndex(resp.Response.Events, app.indexEvents)
	}

	// call the streaming service hook with the BeginBlock messages
	for _, abciListener := range app.streamingManager.ABCIListeners {
		ctx := app.finalizeBlockState.ctx
		blockHeight := ctx.BlockHeight()

		// TODO: Figure out what to do with ListenBeginBlock and the types necessary
		// as we cannot have the store sub-module depend on the root SDK module.
		//
		// Ref: https://github.com/cosmos/cosmos-sdk/issues/12272
		if err := abciListener.ListenBeginBlock(ctx, req, res); err != nil {
			app.logger.Error("BeginBlock listening hook failed", "height", blockHeight, "err", err)
		}
	}

	return resp
}

func (app *BaseApp) legacyEndBlock(req *abci.RequestFinalizeBlock) sdk.LegacyResponseEndBlock {
	var (
		resp sdk.LegacyResponseEndBlock
		err  error
	)

	// TODO: Fill out!
	legacyReq := sdk.LegacyRequestEndBlock{}

	if app.endBlocker != nil {
		resp, err = app.endBlocker(app.finalizeBlockState.ctx, legacyReq)
		if err != nil {
			panic(err)
		}

		resp.Response.Events = sdk.MarkEventsToIndex(resp.Response.Events, app.indexEvents)
	}

	cp := app.GetConsensusParams(app.finalizeBlockState.ctx)
	resp.Response.ConsensusParamUpdates = &cp

	// call the streaming service hook with the EndBlock messages
	for _, abciListener := range app.streamingManager.ABCIListeners {
		ctx := app.finalizeBlockState.ctx
		blockHeight := ctx.BlockHeight()

		// TODO: Figure out what to do with ListenEndBlock and the types necessary
		// as we cannot have the store sub-module depend on the root SDK module.
		//
		// Ref: https://github.com/cosmos/cosmos-sdk/issues/12272
		if err := abciListener.ListenEndBlock(ctx, req, res); err != nil {
			app.logger.Error("EndBlock listening hook failed", "height", blockHeight, "err", err)
		}
	}

	return resp
}

// CheckTx implements the ABCI interface and executes a tx in CheckTx mode. In
// CheckTx mode, messages are not executed. This means messages are only validated
// and only the AnteHandler is executed. State is persisted to the BaseApp's
// internal CheckTx state if the AnteHandler passes. Otherwise, the ResponseCheckTx
// will contain relevant error information. Regardless of tx execution outcome,
// the ResponseCheckTx will contain relevant gas execution context.
func (app *BaseApp) CheckTx(_ context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	var mode execMode

	switch {
	case req.Type == abci.CheckTxType_New:
		mode = execModeCheck

	case req.Type == abci.CheckTxType_Recheck:
		mode = execModeReCheck

	default:
		return nil, fmt.Errorf("unknown RequestCheckTx type: %s", req.Type)
	}

	gInfo, result, anteEvents, err := app.runTx(mode, req.Tx)
	if err != nil {
		return sdkerrors.ResponseCheckTxWithEvents(err, gInfo.GasWanted, gInfo.GasUsed, anteEvents, app.trace), nil
	}

	return &abci.ResponseCheckTx{
		GasWanted: int64(gInfo.GasWanted), // TODO: Should type accept unsigned ints?
		GasUsed:   int64(gInfo.GasUsed),   // TODO: Should type accept unsigned ints?
		Log:       result.Log,
		Data:      result.Data,
		Events:    sdk.MarkEventsToIndex(result.Events, app.indexEvents),
	}, nil
}

// PrepareProposal implements the PrepareProposal ABCI method and returns a
// ResponsePrepareProposal object to the client. The PrepareProposal method is
// responsible for allowing the block proposer to perform application-dependent
// work in a block before proposing it.
//
// Transactions can be modified, removed, or added by the application. Since the
// application maintains its own local mempool, it will ignore the transactions
// provided to it in RequestPrepareProposal. Instead, it will determine which
// transactions to return based on the mempool's semantics and the MaxTxBytes
// provided by the client's request.
//
// Ref: https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-060-abci-1.0.md
// Ref: https://github.com/cometbft/cometbft/blob/main/spec/abci/abci%2B%2B_basic_concepts.md
func (app *BaseApp) PrepareProposal(_ context.Context, req *abci.RequestPrepareProposal) (resp *abci.ResponsePrepareProposal, err error) {
	if app.prepareProposal == nil {
		return nil, errors.New("PrepareProposal method not set")
	}

	// Always reset state given that PrepareProposal can timeout and be called
	// again in a subsequent round.
	header := cmtproto.Header{
		ChainID:            app.chainID,
		Height:             req.Height,
		Time:               req.Time,
		ProposerAddress:    req.ProposerAddress,
		NextValidatorsHash: req.NextValidatorsHash,
	}
	app.setState(execModePrepareProposal, header)

	// CometBFT must never call PrepareProposal with a height of 0.
	//
	// Ref: https://github.com/cometbft/cometbft/blob/059798a4f5b0c9f52aa8655fa619054a0154088c/spec/core/state.md?plain=1#L37-L38
	if req.Height < 1 {
		return nil, errors.New("PrepareProposal called with invalid height")
	}

	app.prepareProposalState.ctx = app.getContextForProposal(app.prepareProposalState.ctx, req.Height).
		WithVoteInfos(app.voteInfos).
		WithBlockHeight(req.Height).
		WithBlockTime(req.Time).
		WithProposer(req.ProposerAddress)

	app.prepareProposalState.ctx = app.prepareProposalState.ctx.
		WithConsensusParams(app.GetConsensusParams(app.prepareProposalState.ctx)).
		WithBlockGasMeter(app.getBlockGasMeter(app.prepareProposalState.ctx))

	defer func() {
		if err := recover(); err != nil {
			app.logger.Error(
				"panic recovered in PrepareProposal",
				"height", req.Height,
				"time", req.Time,
				"panic", err,
			)

			resp = &abci.ResponsePrepareProposal{Txs: req.Txs}
		}
	}()

	resp, err = app.prepareProposal(app.prepareProposalState.ctx, req)
	if err != nil {
		app.logger.Error("failed to prepare proposal", "height", req.Height, "error", err)
		return &abci.ResponsePrepareProposal{Txs: [][]byte{}}, nil
	}

	return resp, nil
}

// ProcessProposal implements the ProcessProposal ABCI method and returns a
// ResponseProcessProposal object to the client. The ProcessProposal method is
// responsible for allowing execution of application-dependent work in a proposed
// block. Note, the application defines the exact implementation details of
// ProcessProposal. In general, the application must at the very least ensure
// that all transactions are valid. If all transactions are valid, then we inform
// CometBFT that the Status is ACCEPT. However, the application is also able
// to implement optimizations such as executing the entire proposed block
// immediately.
//
// If a panic is detected during execution of an application's ProcessProposal
// handler, it will be recovered and we will reject the proposal.
//
// Ref: https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-060-abci-1.0.md
// Ref: https://github.com/cometbft/cometbft/blob/main/spec/abci/abci%2B%2B_basic_concepts.md
func (app *BaseApp) ProcessProposal(_ context.Context, req *abci.RequestProcessProposal) (resp *abci.ResponseProcessProposal, err error) {
	if app.processProposal == nil {
		return nil, errors.New("app.ProcessProposal is not set")
	}

	// CometBFT must never call ProcessProposal with a height of 0.
	// Ref: https://github.com/cometbft/cometbft/blob/059798a4f5b0c9f52aa8655fa619054a0154088c/spec/core/state.md?plain=1#L37-L38
	if req.Height < 1 {
		return nil, errors.New("ProcessProposal called with invalid height")
	}

	// Always reset state given that ProcessProposal can timeout and be called
	// again in a subsequent round.
	header := cmtproto.Header{
		ChainID:            app.chainID,
		Height:             req.Height,
		Time:               req.Time,
		ProposerAddress:    req.ProposerAddress,
		NextValidatorsHash: req.NextValidatorsHash,
	}
	app.setState(execModeProcessProposal, header)

	// Since the application can get access to FinalizeBlock state and write to it,
	// we must be sure to reset it in case ProcessProposal timeouts and is called
	// again in a subsequent round.
	app.setState(execModeFinalize, header)

	app.processProposalState.ctx = app.getContextForProposal(app.processProposalState.ctx, req.Height).
		WithVoteInfos(app.voteInfos).
		WithBlockHeight(req.Height).
		WithBlockTime(req.Time).
		WithHeaderHash(req.Hash).
		WithProposer(req.ProposerAddress)

	app.processProposalState.ctx = app.processProposalState.ctx.
		WithConsensusParams(app.GetConsensusParams(app.processProposalState.ctx)).
		WithBlockGasMeter(app.getBlockGasMeter(app.processProposalState.ctx))

	defer func() {
		if err := recover(); err != nil {
			app.logger.Error(
				"panic recovered in ProcessProposal",
				"height", req.Height,
				"time", req.Time,
				"hash", fmt.Sprintf("%X", req.Hash),
				"panic", err,
			)
			resp = &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}
		}
	}()

	resp, err = app.processProposal(app.processProposalState.ctx, req)
	if err != nil {
		app.logger.Error("failed to process proposal", "height", req.Height, "error", err)
		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
	}

	return resp, nil
}

// ExtendVote implements the ExtendVote ABCI method and returns a ResponseExtendVote.
// It calls the application's ExtendVote handler which is responsible for performing
// application-specific business logic when sending a pre-commit for the NEXT
// block height. The extensions response may be non-deterministic but must always
// be returned, even if empty.
//
// Agreed upon vote extensions are made available to the proposer of the next
// height and are committed in subsequent height, i.e. H+2. An error is returned
// if vote extensions are not enabled or if extendVote fails or panics.
func (app *BaseApp) ExtendVote(_ context.Context, req *abci.RequestExtendVote) (resp *abci.ResponseExtendVote, err error) {
	// Always reset state given that ExtendVote and VerifyVoteExtension can timeout
	// and be called again in a subsequent round.
	emptyHeader := cmtproto.Header{ChainID: app.chainID, Height: req.Height}
	app.setState(execModeVoteExtension, emptyHeader)

	// If vote extensions are not enabled, as a safety precaution, we return an
	// error.
	cp := app.GetConsensusParams(app.voteExtensionState.ctx)
	if cp.Abci != nil && cp.Abci.VoteExtensionsEnableHeight <= 0 {
		return nil, fmt.Errorf("vote extensions are not enabled; unexpected call to ExtendVote at height %d", req.Height)
	}

	app.voteExtensionState.ctx = app.voteExtensionState.ctx.
		WithConsensusParams(cp).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter()).
		WithBlockHeight(req.Height).
		WithHeaderHash(req.Hash)

	// add a deferred recover handler in case extendVote panics
	defer func() {
		if err := recover(); err != nil {
			app.logger.Error(
				"panic recovered in ExtendVote",
				"height", req.Height,
				"hash", fmt.Sprintf("%X", req.Hash),
				"panic", err,
			)
			err = fmt.Errorf("recovered application panic in ExtendVote: %w", err)
		}
	}()

	resp, err = app.extendVote(app.voteExtensionState.ctx, req)
	if err != nil {
		app.logger.Error("failed to extend vote", "height", req.Height, "error", err)
		return nil, fmt.Errorf("failed to extend vote: %w", err)
	}

	return resp, err
}

// VerifyVoteExtension implements the VerifyVoteExtension ABCI method and returns
// a ResponseVerifyVoteExtension. It calls the applications' VerifyVoteExtension
// handler which is responsible for performing application-specific business
// logic in verifying a vote extension from another validator during the pre-commit
// phase. The response MUST be deterministic. An error is returned if vote
// extensions are not enabled or if verifyVoteExt fails or panics.
func (app *BaseApp) VerifyVoteExtension(_ context.Context, req *abci.RequestVerifyVoteExtension) (resp *abci.ResponseVerifyVoteExtension, err error) {
	// If vote extensions are not enabled, as a safety precaution, we return an
	// error.
	cp := app.GetConsensusParams(app.voteExtensionState.ctx)
	if cp.Abci != nil && cp.Abci.VoteExtensionsEnableHeight <= 0 {
		return nil, fmt.Errorf("vote extensions are not enabled; unexpected call to VerifyVoteExtension at height %d", req.Height)
	}

	// add a deferred recover handler in case verifyVoteExt panics
	defer func() {
		if err := recover(); err != nil {
			app.logger.Error(
				"panic recovered in VerifyVoteExtension",
				"height", req.Height,
				"hash", fmt.Sprintf("%X", req.Hash),
				"validator", fmt.Sprintf("%X", req.ValidatorAddress),
				"panic", err,
			)
			err = fmt.Errorf("recovered application panic in VerifyVoteExtension: %w", err)
		}
	}()

	resp, err = app.verifyVoteExt(app.voteExtensionState.ctx, req)
	if err != nil {
		app.logger.Error("failed to verify vote extension", "height", req.Height, "error", err)
		return nil, fmt.Errorf("failed to verify vote extension: %w", err)
	}

	return resp, err
}

func (app *BaseApp) legacyDeliverTx(tx []byte) *abci.ExecTxResult {
	gInfo := sdk.GasInfo{}
	resultStr := "successful"

	var resp *abci.ExecTxResult
	defer func() {
		// call the streaming service hook with the EndBlock messages
		for _, abciListener := range app.streamingManager.ABCIListeners {
			ctx := app.finalizeBlockState.ctx
			blockHeight := ctx.BlockHeight()

			// TODO: Figure out what to do with ListenDeliverTx and the types necessary
			// as we cannot have the store sub-module depend on the root SDK module.
			//
			// Ref: https://github.com/cosmos/cosmos-sdk/issues/12272
			if err := abciListener.ListenDeliverTx(ctx, req, resp); err != nil {
				app.logger.Error("DeliverTx listening hook failed", "height", blockHeight, "err", err)
			}
		}
	}()

	defer func() {
		telemetry.IncrCounter(1, "tx", "count")
		telemetry.IncrCounter(1, "tx", resultStr)
		telemetry.SetGauge(float32(gInfo.GasUsed), "tx", "gas", "used")
		telemetry.SetGauge(float32(gInfo.GasWanted), "tx", "gas", "wanted")
	}()

	gInfo, result, anteEvents, err := app.runTx(execModeFinalize, tx)
	if err != nil {
		resultStr = "failed"
		resp = sdkerrors.ResponseExecTxResultWithEvents(
			err,
			gInfo.GasWanted,
			gInfo.GasUsed,
			sdk.MarkEventsToIndex(anteEvents, app.indexEvents),
			app.trace,
		)
		return resp
	}

	resp = &abci.ExecTxResult{
		GasWanted: int64(gInfo.GasWanted),
		GasUsed:   int64(gInfo.GasUsed),
		Log:       result.Log,
		Data:      result.Data,
		Events:    sdk.MarkEventsToIndex(result.Events, app.indexEvents),
	}

	return resp
}

func (app *BaseApp) FinalizeBlock(_ context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	if err := app.validateFinalizeBlockHeight(req); err != nil {
		return nil, err
	}

	if app.cms.TracingEnabled() {
		app.cms.SetTracingContext(storetypes.TraceContext(
			map[string]any{"blockHeight": req.Height},
		))
	}

	app.voteInfos = req.DecidedLastCommit.Votes

	header := cmtproto.Header{
		ChainID:            app.chainID,
		Height:             req.Height,
		Time:               req.Time,
		ProposerAddress:    req.ProposerAddress,
		NextValidatorsHash: req.NextValidatorsHash,
	}

	// Initialize the FinalizeBlock state. If this is the first block, it should
	// already be initialized in InitChain. Otherwise app.finalizeBlockState will be
	// nil, since it is reset on Commit.
	if app.finalizeBlockState == nil {
		app.setState(execModeFinalize, header)
	} else {
		// In the first block, app.finalizeBlockState.ctx will already be initialized
		// by InitChain. Context is now updated with Header information.
		app.finalizeBlockState.ctx = app.finalizeBlockState.ctx.
			WithBlockHeader(header).
			WithBlockHeight(req.Height)
	}

	gasMeter := app.getBlockGasMeter(app.finalizeBlockState.ctx)

	app.finalizeBlockState.ctx = app.finalizeBlockState.ctx.
		WithBlockGasMeter(gasMeter).
		WithHeaderHash(req.Hash).
		WithConsensusParams(app.GetConsensusParams(app.finalizeBlockState.ctx)).
		WithVoteInfos(req.DecidedLastCommit.Votes).
		WithMisbehavior(req.Misbehavior)

	if app.checkState != nil {
		app.checkState.ctx = app.checkState.ctx.
			WithBlockGasMeter(gasMeter).
			WithHeaderHash(req.Hash)
	}

	beginBlockResp := app.legacyBeginBlock(req)

	txResults := make([]*abci.ExecTxResult, len(req.Txs))
	for i, rawTx := range req.Txs {
		txResults[i] = app.legacyDeliverTx(rawTx)
	}

	if app.finalizeBlockState.ms.TracingEnabled() {
		app.finalizeBlockState.ms = app.finalizeBlockState.ms.SetTracingContext(nil).(storetypes.CacheMultiStore)
	}

	endBlockResp := app.legacyEndBlock(req)

	// TODO: Populate fields.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/12272
	return &abci.ResponseFinalizeBlock{
		Events:                nil,
		TxResults:             txResults,
		ValidatorUpdates:      nil,
		ConsensusParamUpdates: nil,
		AppHash:               app.flushCommit().Hash,
	}, nil
}

// Commit implements the ABCI interface. It will commit all state that exists in
// the deliver state's multi-store and includes the resulting commit ID in the
// returned abci.ResponseCommit. Commit will set the check state based on the
// latest header and reset the deliver state. Also, if a non-zero halt height is
// defined in config, Commit will execute a deferred function call to check
// against that height and gracefully halt if it matches the latest committed
// height.
func (app *BaseApp) Commit(_ context.Context, _ *abci.RequestCommit) (*abci.ResponseCommit, error) {
	header := app.finalizeBlockState.ctx.BlockHeader()
	retainHeight := app.GetBlockRetentionHeight(header.Height)

	rms, ok := app.cms.(*rootmulti.Store)
	if ok {
		rms.SetCommitHeader(header)
	}

	resp := &abci.ResponseCommit{
		RetainHeight: retainHeight,
	}

	abciListeners := app.streamingManager.ABCIListeners
	if len(abciListeners) > 0 {
		ctx := app.finalizeBlockState.ctx
		blockHeight := ctx.BlockHeight()
		changeSet := app.cms.PopStateCache()

		for _, abciListener := range abciListeners {
			if err := abciListener.ListenCommit(ctx, *resp, changeSet); err != nil {
				app.logger.Error("Commit listening hook failed", "height", blockHeight, "err", err)
			}
		}
	}

	// Reset the CheckTx state to the latest committed.
	//
	// NOTE: This is safe because CometBFT holds a lock on the mempool for
	// Commit. Use the header from this latest block.
	app.setState(execModeCheck, header)

	app.finalizeBlockState = nil

	var halt bool
	switch {
	case app.haltHeight > 0 && uint64(header.Height) >= app.haltHeight:
		halt = true

	case app.haltTime > 0 && header.Time.Unix() >= int64(app.haltTime):
		halt = true
	}

	if halt {
		// Halt the binary and allow CometBFT to receive the ResponseCommit
		// response with the commit ID hash. This will allow the node to successfully
		// restart and process blocks assuming the halt configuration has been
		// reset or moved to a more distant value.
		app.halt()
	}

	go app.snapshotManager.SnapshotIfApplicable(header.Height)

	return resp, nil
}

// flushCommit writes all state transitions that occurred during FinalizeBlock.
// These writes are persisted to the root multi-store (app.cms) and flushed to
// disk. This means when the ABCI client requests Commit(), the application
// state transitions are already flushed to disk and as a result, we already have
// an application Merkle root.
func (app *BaseApp) flushCommit() storetypes.CommitID {
	app.finalizeBlockState.ms.Write()

	commitID := app.cms.Commit()
	app.logger.Info("commit synced", "commit", fmt.Sprintf("%X", commitID))

	return commitID
}

// halt attempts to gracefully shutdown the node via SIGINT and SIGTERM falling
// back on os.Exit if both fail.
func (app *BaseApp) halt() {
	app.logger.Info("halting node per configuration", "height", app.haltHeight, "time", app.haltTime)

	p, err := os.FindProcess(os.Getpid())
	if err == nil {
		// attempt cascading signals in case SIGINT fails (os dependent)
		sigIntErr := p.Signal(syscall.SIGINT)
		sigTermErr := p.Signal(syscall.SIGTERM)

		if sigIntErr == nil || sigTermErr == nil {
			return
		}
	}

	// Resort to exiting immediately if the process could not be found or killed
	// via SIGINT/SIGTERM signals.
	app.logger.Info("failed to send SIGINT/SIGTERM; exiting...")
	os.Exit(0)
}

func handleQueryApp(app *BaseApp, path []string, req *abci.RequestQuery) *abci.ResponseQuery {
	if len(path) >= 2 {
		switch path[1] {
		case "simulate":
			txBytes := req.Data

			gInfo, res, err := app.Simulate(txBytes)
			if err != nil {
				return sdkerrors.QueryResult(errorsmod.Wrap(err, "failed to simulate tx"), app.trace)
			}

			simRes := &sdk.SimulationResponse{
				GasInfo: gInfo,
				Result:  res,
			}

			bz, err := codec.ProtoMarshalJSON(simRes, app.interfaceRegistry)
			if err != nil {
				return sdkerrors.QueryResult(errorsmod.Wrap(err, "failed to JSON encode simulation response"), app.trace)
			}

			return &abci.ResponseQuery{
				Codespace: sdkerrors.RootCodespace,
				Height:    req.Height,
				Value:     bz,
			}

		case "version":
			return &abci.ResponseQuery{
				Codespace: sdkerrors.RootCodespace,
				Height:    req.Height,
				Value:     []byte(app.version),
			}

		default:
			return sdkerrors.QueryResult(errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query: %s", path), app.trace)
		}
	}

	return sdkerrors.QueryResult(
		errorsmod.Wrap(
			sdkerrors.ErrUnknownRequest,
			"expected second parameter to be either 'simulate' or 'version', neither was present",
		), app.trace)
}

func handleQueryStore(app *BaseApp, path []string, req *abci.RequestQuery) *abci.ResponseQuery {
	// "/store" prefix for store queries
	queryable, ok := app.cms.(storetypes.Queryable)
	if !ok {
		return sdkerrors.QueryResult(errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "multi-store does not support queries"), app.trace)
	}

	req.Path = "/" + strings.Join(path[1:], "/")

	if req.Height <= 1 && req.Prove {
		return sdkerrors.QueryResult(
			errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"cannot query with proof when height <= 1; please provide a valid height",
			), app.trace)
	}

	// TODO: Update Query interface method accept a pointer to RequestQuery.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/12272
	resp := queryable.Query(req)
	resp.Height = req.Height

	return resp
}

func handleQueryP2P(app *BaseApp, path []string) *abci.ResponseQuery {
	// "/p2p" prefix for p2p queries
	if len(path) < 4 {
		return sdkerrors.QueryResult(errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "path should be p2p filter <addr|id> <parameter>"), app.trace)
	}

	var resp *abci.ResponseQuery

	cmd, typ, arg := path[1], path[2], path[3]
	switch cmd {
	case "filter":
		switch typ {
		case "addr":
			resp = app.FilterPeerByAddrPort(arg)

		case "id":
			resp = app.FilterPeerByID(arg)
		}

	default:
		resp = sdkerrors.QueryResult(errorsmod.Wrap(sdkerrors.ErrUnknownRequest, "expected second parameter to be 'filter'"), app.trace)
	}

	return resp
}

// SplitABCIQueryPath splits a string path using the delimiter '/'.
//
// e.g. "this/is/funny" becomes []string{"this", "is", "funny"}
func SplitABCIQueryPath(requestPath string) (path []string) {
	path = strings.Split(requestPath, "/")

	// first element is empty string
	if len(path) > 0 && path[0] == "" {
		path = path[1:]
	}

	return path
}

// ValidateVoteExtensions defines a helper function for verifying vote extension
// signatures that may be passed or manually injected into a block proposal from
// a proposer in ProcessProposal. It returns an error if any signature is invalid
// or if unexpected vote extensions and/or signatures are found.
func (app *BaseApp) ValidateVoteExtensions(ctx sdk.Context, currentHeight int64, extCommit abci.ExtendedCommitInfo) error {
	cp := ctx.ConsensusParams()
	extsEnabled := cp.Abci != nil && cp.Abci.VoteExtensionsEnableHeight > 0

	marshalDelimitedFn := func(msg proto.Message) ([]byte, error) {
		var buf bytes.Buffer
		if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	for _, vote := range extCommit.Votes {
		if !extsEnabled {
			if len(vote.VoteExtension) > 0 {
				return fmt.Errorf("vote extensions disabled; received non-empty vote extension at height %d", currentHeight)
			}
			if len(vote.ExtensionSignature) > 0 {
				return fmt.Errorf("vote extensions disabled; received non-empty vote extension signature at height %d", currentHeight)
			}

			continue
		}

		if len(vote.ExtensionSignature) == 0 {
			return fmt.Errorf("vote extensions enabled; received empty vote extension signature at height %d", currentHeight)
		}

		valConsAddr := cmtcrypto.Address(vote.Validator.Address)

		validator, err := app.valStore.GetValidatorByConsAddr(ctx, valConsAddr)
		if err != nil {
			return fmt.Errorf("failed to get validator %X: %w", valConsAddr, err)
		}

		cmtPubKeyProto, err := validator.CmtConsPublicKey()
		if err != nil {
			return fmt.Errorf("failed to get validator %X public key: %w", valConsAddr, err)
		}

		cmtPubKey, err := cryptoenc.PubKeyFromProto(cmtPubKeyProto)
		if err != nil {
			return fmt.Errorf("failed to convert validator %X public key: %w", valConsAddr, err)
		}

		cve := cmtproto.CanonicalVoteExtension{
			Extension: vote.VoteExtension,
			Height:    currentHeight - 1, // the vote extension was signed in the previous height
			Round:     int64(extCommit.Round),
			ChainId:   app.chainID,
		}

		extSignBytes, err := marshalDelimitedFn(&cve)
		if err != nil {
			return fmt.Errorf("failed to encode CanonicalVoteExtension: %w", err)
		}

		if !cmtPubKey.VerifySignature(extSignBytes, vote.ExtensionSignature) {
			return fmt.Errorf("failed to verify validator %X vote extension signature", valConsAddr)
		}
	}

	return nil
}

// FilterPeerByAddrPort filters peers by address/port.
func (app *BaseApp) FilterPeerByAddrPort(info string) *abci.ResponseQuery {
	if app.addrPeerFilter != nil {
		return app.addrPeerFilter(info)
	}

	return &abci.ResponseQuery{}
}

// FilterPeerByID filters peers by node ID.
func (app *BaseApp) FilterPeerByID(info string) *abci.ResponseQuery {
	if app.idPeerFilter != nil {
		return app.idPeerFilter(info)
	}

	return &abci.ResponseQuery{}
}

// getContextForProposal returns the correct Context for PrepareProposal and
// ProcessProposal. We use finalizeBlockState on the first block to be able to
// access any state changes made in InitChain.
func (app *BaseApp) getContextForProposal(ctx sdk.Context, height int64) sdk.Context {
	if height == 1 {
		ctx, _ = app.finalizeBlockState.ctx.CacheContext()

		// clear all context data set during InitChain to avoid inconsistent behavior
		ctx = ctx.WithBlockHeader(cmtproto.Header{})
		return ctx
	}

	return ctx
}

func (app *BaseApp) handleQueryGRPC(handler GRPCQueryHandler, req *abci.RequestQuery) *abci.ResponseQuery {
	ctx, err := app.CreateQueryContext(req.Height, req.Prove)
	if err != nil {
		return sdkerrors.QueryResult(err, app.trace)
	}

	resp, err := handler(ctx, req)
	if err != nil {
		resp = sdkerrors.QueryResult(gRPCErrorToSDKError(err), app.trace)
		resp.Height = req.Height
		return resp
	}

	return resp
}

func gRPCErrorToSDKError(err error) error {
	status, ok := grpcstatus.FromError(err)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	switch status.Code() {
	case codes.NotFound:
		return errorsmod.Wrap(sdkerrors.ErrKeyNotFound, err.Error())

	case codes.InvalidArgument:
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())

	case codes.FailedPrecondition:
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())

	case codes.Unauthenticated:
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, err.Error())

	default:
		return errorsmod.Wrap(sdkerrors.ErrUnknownRequest, err.Error())
	}
}

func checkNegativeHeight(height int64) error {
	if height < 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot query with height < 0; please provide a valid height")
	}

	return nil
}

// createQueryContext creates a new sdk.Context for a query, taking as args
// the block height and whether the query needs a proof or not.
func (app *BaseApp) CreateQueryContext(height int64, prove bool) (sdk.Context, error) {
	if err := checkNegativeHeight(height); err != nil {
		return sdk.Context{}, err
	}

	// use custom query multi-store if provided
	qms := app.qms
	if qms == nil {
		qms = app.cms.(storetypes.MultiStore)
	}

	lastBlockHeight := qms.LatestVersion()
	if lastBlockHeight == 0 {
		return sdk.Context{}, errorsmod.Wrapf(sdkerrors.ErrInvalidHeight, "%s is not ready; please wait for first block", app.Name())
	}

	if height > lastBlockHeight {
		return sdk.Context{},
			errorsmod.Wrap(
				sdkerrors.ErrInvalidHeight,
				"cannot query with height in the future; please provide a valid height",
			)
	}

	// when a client did not provide a query height, manually inject the latest
	if height == 0 {
		height = lastBlockHeight
	}

	if height <= 1 && prove {
		return sdk.Context{},
			errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"cannot query with proof when height <= 1; please provide a valid height",
			)
	}

	cacheMS, err := qms.CacheMultiStoreWithVersion(height)
	if err != nil {
		return sdk.Context{},
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"failed to load state at height %d; %s (latest height: %d)", height, err, lastBlockHeight,
			)
	}

	// branch the commit multi-store for safety
	ctx := sdk.NewContext(cacheMS, app.checkState.ctx.BlockHeader(), true, app.logger).
		WithMinGasPrices(app.minGasPrices).
		WithBlockHeight(height)

	if height != lastBlockHeight {
		rms, ok := app.cms.(*rootmulti.Store)
		if ok {
			cInfo, err := rms.GetCommitInfo(height)
			if cInfo != nil && err == nil {
				ctx = ctx.WithBlockTime(cInfo.Timestamp)
			}
		}
	}

	return ctx, nil
}

// GetBlockRetentionHeight returns the height for which all blocks below this height
// are pruned from CometBFT. Given a commitment height and a non-zero local
// minRetainBlocks configuration, the retentionHeight is the smallest height that
// satisfies:
//
// - Unbonding (safety threshold) time: The block interval in which validators
// can be economically punished for misbehavior. Blocks in this interval must be
// auditable e.g. by the light client.
//
// - Logical store snapshot interval: The block interval at which the underlying
// logical store database is persisted to disk, e.g. every 10000 heights. Blocks
// since the last IAVL snapshot must be available for replay on application restart.
//
// - State sync snapshots: Blocks since the oldest available snapshot must be
// available for state sync nodes to catch up (oldest because a node may be
// restoring an old snapshot while a new snapshot was taken).
//
// - Local (minRetainBlocks) config: Archive nodes may want to retain more or
// all blocks, e.g. via a local config option min-retain-blocks. There may also
// be a need to vary retention for other nodes, e.g. sentry nodes which do not
// need historical blocks.
func (app *BaseApp) GetBlockRetentionHeight(commitHeight int64) int64 {
	// pruning is disabled if minRetainBlocks is zero
	if app.minRetainBlocks == 0 {
		return 0
	}

	minNonZero := func(x, y int64) int64 {
		switch {
		case x == 0:
			return y

		case y == 0:
			return x

		case x < y:
			return x

		default:
			return y
		}
	}

	// Define retentionHeight as the minimum value that satisfies all non-zero
	// constraints. All blocks below (commitHeight-retentionHeight) are pruned
	// from CometBFT.
	var retentionHeight int64

	// Define the number of blocks needed to protect against misbehaving validators
	// which allows light clients to operate safely. Note, we piggy back of the
	// evidence parameters instead of computing an estimated number of blocks based
	// on the unbonding period and block commitment time as the two should be
	// equivalent.
	cp := app.GetConsensusParams(app.finalizeBlockState.ctx)
	if cp.Evidence != nil && cp.Evidence.MaxAgeNumBlocks > 0 {
		retentionHeight = commitHeight - cp.Evidence.MaxAgeNumBlocks
	}

	if app.snapshotManager != nil {
		snapshotRetentionHeights := app.snapshotManager.GetSnapshotBlockRetentionHeights()
		if snapshotRetentionHeights > 0 {
			retentionHeight = minNonZero(retentionHeight, commitHeight-snapshotRetentionHeights)
		}
	}

	v := commitHeight - int64(app.minRetainBlocks)
	retentionHeight = minNonZero(retentionHeight, v)

	if retentionHeight <= 0 {
		// prune nothing in the case of a non-positive height
		return 0
	}

	return retentionHeight
}
