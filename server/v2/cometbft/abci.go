package cometbft

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	abci "github.com/cometbft/cometbft/abci/types"
	abciproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	gogoproto "github.com/cosmos/gogoproto/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/collections"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors/v2"
	"cosmossdk.io/log"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/client/grpc/cmtservice"
	"cosmossdk.io/server/v2/cometbft/handlers"
	"cosmossdk.io/server/v2/cometbft/mempool"
	"cosmossdk.io/server/v2/cometbft/types"
	cometerrors "cosmossdk.io/server/v2/cometbft/types/errors"
	"cosmossdk.io/server/v2/streaming"
	"cosmossdk.io/store/v2/snapshots"
	consensustypes "cosmossdk.io/x/consensus/types"
)

var _ abci.Application = (*Consensus[transaction.Tx])(nil)

type Consensus[T transaction.Tx] struct {
	logger           log.Logger
	appName, version string
	app              appmanager.AppManager[T]
	appCloser        func() error
	txCodec          transaction.Codec[T]
	store            types.Store
	streaming        streaming.Manager
	listener         *appdata.Listener
	snapshotManager  *snapshots.Manager
	mempool          mempool.Mempool[T]

	cfg           Config
	indexedEvents map[string]struct{}
	chainID       string

	initialHeight uint64
	// this is only available after this node has committed a block (in FinalizeBlock),
	// otherwise it will be empty and we will need to query the app for the last
	// committed block.
	lastCommittedHeight atomic.Int64

	prepareProposalHandler handlers.PrepareHandler[T]
	processProposalHandler handlers.ProcessHandler[T]
	verifyVoteExt          handlers.VerifyVoteExtensionhandler
	extendVote             handlers.ExtendVoteHandler
	checkTxHandler         handlers.CheckTxHandler[T]

	addrPeerFilter types.PeerFilter // filter peers by address and port
	idPeerFilter   types.PeerFilter // filter peers by node ID

	queryHandlersMap map[string]appmodulev2.Handler
	getProtoRegistry func() (*protoregistry.Files, error)
}

func NewConsensus[T transaction.Tx](
	logger log.Logger,
	appName string,
	app appmanager.AppManager[T],
	appCloser func() error,
	mp mempool.Mempool[T],
	indexedEvents map[string]struct{},
	queryHandlersMap map[string]appmodulev2.Handler,
	store types.Store,
	cfg Config,
	txCodec transaction.Codec[T],
	chainId string,
) *Consensus[T] {
	return &Consensus[T]{
		appName:                appName,
		version:                getCometBFTServerVersion(),
		app:                    app,
		appCloser:              appCloser,
		cfg:                    cfg,
		store:                  store,
		logger:                 logger,
		txCodec:                txCodec,
		streaming:              streaming.Manager{},
		snapshotManager:        nil,
		mempool:                mp,
		lastCommittedHeight:    atomic.Int64{},
		prepareProposalHandler: nil,
		processProposalHandler: nil,
		verifyVoteExt:          nil,
		extendVote:             nil,
		chainID:                chainId,
		indexedEvents:          indexedEvents,
		initialHeight:          0,
		queryHandlersMap:       queryHandlersMap,
		getProtoRegistry:       sync.OnceValues(gogoproto.MergedRegistry),
	}
}

// SetStreamingManager sets the streaming manager for the consensus module.
func (c *Consensus[T]) SetStreamingManager(sm streaming.Manager) {
	c.streaming = sm
}

// RegisterSnapshotExtensions registers the given extensions with the consensus module's snapshot manager.
// It allows additional snapshotter implementations to be used for creating and restoring snapshots.
func (c *Consensus[T]) RegisterSnapshotExtensions(extensions ...snapshots.ExtensionSnapshotter) error {
	if err := c.snapshotManager.RegisterExtensions(extensions...); err != nil {
		return fmt.Errorf("failed to register snapshot extensions: %w", err)
	}

	return nil
}

// CheckTx implements types.Application.
// It is called by cometbft to verify transaction validity
func (c *Consensus[T]) CheckTx(ctx context.Context, req *abciproto.CheckTxRequest) (*abciproto.CheckTxResponse, error) {
	decodedTx, err := c.txCodec.Decode(req.Tx)
	if err != nil {
		return nil, err
	}

	if c.checkTxHandler == nil {
		resp, err := c.app.ValidateTx(ctx, decodedTx)
		// we do not want to return a cometbft error, but a check tx response with the error
		if err != nil && !errors.Is(err, resp.Error) {
			return nil, err
		}

		events, err := intoABCIEvents(resp.Events, c.indexedEvents)
		if err != nil {
			return nil, err
		}

		cometResp := &abciproto.CheckTxResponse{
			Code:      0,
			GasWanted: uint64ToInt64(resp.GasWanted),
			GasUsed:   uint64ToInt64(resp.GasUsed),
			Events:    events,
		}
		if resp.Error != nil {
			space, code, log := errorsmod.ABCIInfo(resp.Error, c.cfg.AppTomlConfig.Trace)
			cometResp.Code = code
			cometResp.Codespace = space
			cometResp.Log = log
		}

		return cometResp, nil
	}

	return c.checkTxHandler(c.app.ValidateTx)
}

// Info implements types.Application.
func (c *Consensus[T]) Info(ctx context.Context, _ *abciproto.InfoRequest) (*abciproto.InfoResponse, error) {
	version, _, err := c.store.StateLatest()
	if err != nil {
		return nil, err
	}

	// if height is 0, we dont know the consensus params
	var appVersion uint64 = 0
	if version > 0 {
		cp, err := c.GetConsensusParams(ctx)
		// if the consensus params are not found, we set the app version to 0
		// in the case that the start version is > 0
		if cp == nil || errors.Is(err, collections.ErrNotFound) {
			appVersion = 0
		} else if err != nil {
			return nil, err
		} else {
			appVersion = cp.Version.GetApp()
		}
		if err != nil {
			return nil, err
		}
	}

	cid, err := c.store.LastCommitID()
	if err != nil {
		return nil, err
	}

	return &abciproto.InfoResponse{
		Data:             c.appName,
		Version:          c.version,
		AppVersion:       appVersion,
		LastBlockHeight:  int64(version),
		LastBlockAppHash: cid.Hash,
	}, nil
}

// Query implements types.Application.
// It is called by cometbft to query application state.
func (c *Consensus[T]) Query(ctx context.Context, req *abciproto.QueryRequest) (resp *abciproto.QueryResponse, err error) {
	resp, isGRPC, err := c.maybeRunGRPCQuery(ctx, req)
	if isGRPC {
		return resp, err
	}

	// this error most probably means that we can't handle it with a proto message, so
	// it must be an app/p2p/store query
	path := splitABCIQueryPath(req.Path)
	if len(path) == 0 {
		return QueryResult(errorsmod.Wrap(cometerrors.ErrUnknownRequest, "no query path provided"), c.cfg.AppTomlConfig.Trace), nil
	}

	switch path[0] {
	case cmtservice.QueryPathApp:
		resp, err = c.handlerQueryApp(ctx, path, req)

	case cmtservice.QueryPathStore:
		resp, err = c.handleQueryStore(path, c.store, req)

	case cmtservice.QueryPathP2P:
		resp, err = c.handleQueryP2P(path)

	default:
		resp = QueryResult(errorsmod.Wrap(cometerrors.ErrUnknownRequest, "unknown query path"), c.cfg.AppTomlConfig.Trace)
	}

	if err != nil {
		return QueryResult(err, c.cfg.AppTomlConfig.Trace), nil
	}

	return resp, nil
}

func (c *Consensus[T]) maybeRunGRPCQuery(ctx context.Context, req *abci.QueryRequest) (resp *abciproto.QueryResponse, isGRPC bool, err error) {
	// if this fails then we cannot serve queries anymore
	registry, err := c.getProtoRegistry()
	if err != nil {
		return nil, false, err
	}

	path := strings.TrimPrefix(req.Path, "/")
	pathFullName := protoreflect.FullName(strings.ReplaceAll(path, "/", "."))

	// in order to check if it's a gRPC query we ensure that there's a descriptor
	// for the path, if such descriptor exists, and it is a method descriptor
	// then we assume this is a gRPC query.
	desc, err := registry.FindDescriptorByName(pathFullName)
	if err != nil {
		return nil, false, err
	}

	md, isGRPC := desc.(protoreflect.MethodDescriptor)
	if !isGRPC {
		return nil, false, nil
	}

	handler, found := c.queryHandlersMap[string(md.Input().FullName())]
	if !found {
		return nil, true, fmt.Errorf("no query handler found for %s", req.Path)
	}
	protoRequest := handler.MakeMsg()
	err = gogoproto.Unmarshal(req.Data, protoRequest) // TODO: use codec
	if err != nil {
		return nil, true, fmt.Errorf("unable to decode gRPC request with path %s from ABCI.Query: %w", req.Path, err)
	}
	res, err := c.app.Query(ctx, uint64(req.Height), protoRequest)
	if err != nil {
		resp := QueryResult(err, c.cfg.AppTomlConfig.Trace)
		resp.Height = req.Height
		return resp, true, err

	}

	resp, err = queryResponse(res, req.Height)
	return resp, isGRPC, err
}

// InitChain implements types.Application.
func (c *Consensus[T]) InitChain(ctx context.Context, req *abciproto.InitChainRequest) (*abciproto.InitChainResponse, error) {
	c.logger.Info("InitChain", "initialHeight", req.InitialHeight, "chainID", req.ChainId)

	// store chainID to be used later on in execution
	c.chainID = req.ChainId

	// TODO: check if we need to load the config from genesis.json or config.toml
	c.initialHeight = uint64(req.InitialHeight)
	if c.initialHeight == 0 { // If initial height is 0, set it to 1
		c.initialHeight = 1
	}

	if req.ConsensusParams != nil {
		ctx = context.WithValue(ctx, corecontext.CometParamsInitInfoKey, &consensustypes.MsgUpdateParams{
			Block:     req.ConsensusParams.Block,
			Evidence:  req.ConsensusParams.Evidence,
			Validator: req.ConsensusParams.Validator,
			Abci:      req.ConsensusParams.Abci,
			Synchrony: req.ConsensusParams.Synchrony,
			Feature:   req.ConsensusParams.Feature,
		})
	}

	ci, err := c.store.LastCommitID()
	if err != nil {
		return nil, err
	}

	// populate hash with empty byte slice instead of nil
	bz := sha256.Sum256([]byte{})

	br := &server.BlockRequest[T]{
		Height:    uint64(req.InitialHeight - 1),
		Time:      req.Time,
		Hash:      bz[:],
		AppHash:   ci.Hash,
		ChainId:   req.ChainId,
		IsGenesis: true,
	}

	blockresponse, genesisState, err := c.app.InitGenesis(
		ctx,
		br,
		req.AppStateBytes,
		c.txCodec)
	if err != nil {
		return nil, fmt.Errorf("genesis state init failure: %w", err)
	}

	for _, txRes := range blockresponse.TxResults {
		if err := txRes.Error; err != nil {
			space, code, log := errorsmod.ABCIInfo(err, c.cfg.AppTomlConfig.Trace)
			c.logger.Warn("genesis tx failed", "codespace", space, "code", code, "log", log)
		}
	}

	validatorUpdates := intoABCIValidatorUpdates(blockresponse.ValidatorUpdates)

	// set the initial version of the store
	if err := c.store.SetInitialVersion(uint64(req.InitialHeight)); err != nil {
		return nil, fmt.Errorf("failed to set initial version: %w", err)
	}

	stateChanges, err := genesisState.GetStateChanges()
	if err != nil {
		return nil, err
	}
	cs := &store.Changeset{
		Changes: stateChanges,
	}
	stateRoot, err := c.store.WorkingHash(cs)
	if err != nil {
		return nil, fmt.Errorf("unable to write the changeset: %w", err)
	}

	return &abciproto.InitChainResponse{
		ConsensusParams: req.ConsensusParams,
		Validators:      validatorUpdates,
		AppHash:         stateRoot,
	}, nil
}

// PrepareProposal implements types.Application.
// It is called by cometbft to prepare a proposal block.
func (c *Consensus[T]) PrepareProposal(
	ctx context.Context,
	req *abciproto.PrepareProposalRequest,
) (resp *abciproto.PrepareProposalResponse, err error) {
	if req.Height < 1 {
		return nil, errors.New("PrepareProposal called with invalid height")
	}

	if c.prepareProposalHandler == nil {
		return nil, errors.New("no prepare proposal function was set")
	}

	ciCtx := contextWithCometInfo(ctx, comet.Info{
		Evidence:        toCoreEvidence(req.Misbehavior),
		ValidatorsHash:  req.NextValidatorsHash,
		ProposerAddress: req.ProposerAddress,
		LastCommit:      toCoreExtendedCommitInfo(req.LocalLastCommit),
	})

	txs, err := c.prepareProposalHandler(ciCtx, c.app, c.txCodec, req)
	if err != nil {
		return nil, err
	}

	encodedTxs := make([][]byte, len(txs))
	for i, tx := range txs {
		encodedTxs[i] = tx.Bytes()
	}

	return &abciproto.PrepareProposalResponse{
		Txs: encodedTxs,
	}, nil
}

// ProcessProposal implements types.Application.
// It is called by cometbft to process/verify a proposal block.
func (c *Consensus[T]) ProcessProposal(
	ctx context.Context,
	req *abciproto.ProcessProposalRequest,
) (*abciproto.ProcessProposalResponse, error) {
	if req.Height < 1 {
		return nil, errors.New("ProcessProposal called with invalid height")
	}

	if c.processProposalHandler == nil {
		return nil, errors.New("no process proposal function was set")
	}

	ciCtx := contextWithCometInfo(ctx, comet.Info{
		Evidence:        toCoreEvidence(req.Misbehavior),
		ValidatorsHash:  req.NextValidatorsHash,
		ProposerAddress: req.ProposerAddress,
		LastCommit:      toCoreCommitInfo(req.ProposedLastCommit),
	})

	err := c.processProposalHandler(ciCtx, c.app, c.txCodec, req)
	if err != nil {
		c.logger.Error("failed to process proposal", "height", req.Height, "time", req.Time, "hash", fmt.Sprintf("%X", req.Hash), "err", err)
		return &abciproto.ProcessProposalResponse{
			Status: abciproto.PROCESS_PROPOSAL_STATUS_REJECT,
		}, nil
	}

	return &abciproto.ProcessProposalResponse{
		Status: abciproto.PROCESS_PROPOSAL_STATUS_ACCEPT,
	}, nil
}

// FinalizeBlock implements types.Application.
// It is called by cometbft to finalize a block.
func (c *Consensus[T]) FinalizeBlock(
	ctx context.Context,
	req *abciproto.FinalizeBlockRequest,
) (*abciproto.FinalizeBlockResponse, error) {
	if err := c.validateFinalizeBlockHeight(req); err != nil {
		return nil, err
	}

	if err := c.checkHalt(req.Height, req.Time); err != nil {
		return nil, err
	}

	// we don't need to deliver the block in the genesis block
	if req.Height == int64(c.initialHeight) {
		appHash, err := c.store.Commit(store.NewChangeset())
		if err != nil {
			return nil, fmt.Errorf("unable to commit the changeset: %w", err)
		}
		c.lastCommittedHeight.Store(req.Height)
		return &abciproto.FinalizeBlockResponse{
			AppHash: appHash,
		}, nil
	}

	// TODO(tip): can we expect some txs to not decode? if so, what we do in this case? this does not seem to be the case,
	// considering that prepare and process always decode txs, assuming they're the ones providing txs we should never
	// have a tx that fails decoding.
	decodedTxs, err := decodeTxs(req.Txs, c.txCodec)
	if err != nil {
		return nil, err
	}

	cid, err := c.store.LastCommitID()
	if err != nil {
		return nil, err
	}

	blockReq := &server.BlockRequest[T]{
		Height:  uint64(req.Height),
		Time:    req.Time,
		Hash:    req.Hash,
		AppHash: cid.Hash,
		ChainId: c.chainID,
		Txs:     decodedTxs,
	}

	ciCtx := contextWithCometInfo(ctx, comet.Info{
		Evidence:        toCoreEvidence(req.Misbehavior),
		ValidatorsHash:  req.NextValidatorsHash,
		ProposerAddress: req.ProposerAddress,
		LastCommit:      toCoreCommitInfo(req.DecidedLastCommit),
	})

	resp, newState, err := c.app.DeliverBlock(ciCtx, blockReq)
	if err != nil {
		return nil, err
	}

	// after we get the changeset we can produce the commit hash,
	// from the store.
	stateChanges, err := newState.GetStateChanges()
	if err != nil {
		return nil, err
	}
	appHash, err := c.store.Commit(&store.Changeset{Changes: stateChanges})
	if err != nil {
		return nil, fmt.Errorf("unable to commit the changeset: %w", err)
	}

	var events []event.Event
	events = append(events, resp.PreBlockEvents...)
	events = append(events, resp.BeginBlockEvents...)
	for _, tx := range resp.TxResults {
		events = append(events, tx.Events...)
	}
	events = append(events, resp.EndBlockEvents...)

	// listen to state streaming changes in accordance with the block
	err = c.streamDeliverBlockChanges(ctx, req.Height, req.Txs, resp.TxResults, events, stateChanges)
	if err != nil {
		return nil, err
	}

	// remove txs from the mempool
	for _, tx := range decodedTxs {
		if err = c.mempool.Remove(tx); err != nil {
			return nil, fmt.Errorf("unable to remove tx: %w", err)
		}
	}

	c.lastCommittedHeight.Store(req.Height)

	cp, err := c.GetConsensusParams(ctx) // we get the consensus params from the latest state because we committed state above
	if err != nil {
		return nil, err
	}

	return finalizeBlockResponse(resp, cp, appHash, c.indexedEvents, c.cfg.AppTomlConfig.Trace)
}

// Commit implements types.Application.
// It is called by cometbft to notify the application that a block was committed.
func (c *Consensus[T]) Commit(ctx context.Context, _ *abciproto.CommitRequest) (*abciproto.CommitResponse, error) {
	lastCommittedHeight := c.lastCommittedHeight.Load()

	c.snapshotManager.SnapshotIfApplicable(lastCommittedHeight)

	cp, err := c.GetConsensusParams(ctx)
	if err != nil {
		return nil, err
	}

	return &abci.CommitResponse{
		RetainHeight: c.GetBlockRetentionHeight(cp, lastCommittedHeight),
	}, nil
}

// Vote extensions

// VerifyVoteExtension implements types.Application.
func (c *Consensus[T]) VerifyVoteExtension(
	ctx context.Context,
	req *abciproto.VerifyVoteExtensionRequest,
) (*abciproto.VerifyVoteExtensionResponse, error) {
	// If vote extensions are not enabled, as a safety precaution, we return an
	// error.
	cp, err := c.GetConsensusParams(ctx)
	if err != nil {
		return nil, err
	}

	// Note: we verify votes extensions on VoteExtensionsEnableHeight+1. Check
	// comment in ExtendVote and ValidateVoteExtensions for more details.
	// Since Abci was deprecated, should check both Feature & Abci
	extsEnabled := cp.Feature.VoteExtensionsEnableHeight != nil && req.Height >= cp.Feature.VoteExtensionsEnableHeight.Value && cp.Feature.VoteExtensionsEnableHeight.Value != 0
	if !extsEnabled {
		// check abci params
		extsEnabled = cp.Abci != nil && req.Height >= cp.Abci.VoteExtensionsEnableHeight && cp.Abci.VoteExtensionsEnableHeight != 0
		if !extsEnabled {
			return nil, fmt.Errorf("vote extensions are not enabled; unexpected call to VerifyVoteExtension at height %d", req.Height)
		}
	}

	if c.verifyVoteExt == nil {
		return nil, errors.New("vote extensions are enabled but no verify function was set")
	}

	_, latestStore, err := c.store.StateLatest()
	if err != nil {
		return nil, err
	}

	resp, err := c.verifyVoteExt(ctx, latestStore, req)
	if err != nil {
		c.logger.Error("failed to verify vote extension", "height", req.Height, "err", err)
		return &abciproto.VerifyVoteExtensionResponse{Status: abciproto.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
	}

	return resp, err
}

// ExtendVote implements types.Application.
func (c *Consensus[T]) ExtendVote(ctx context.Context, req *abciproto.ExtendVoteRequest) (*abciproto.ExtendVoteResponse, error) {
	// If vote extensions are not enabled, as a safety precaution, we return an
	// error.
	cp, err := c.GetConsensusParams(ctx)
	if err != nil {
		return nil, err
	}

	// Note: In this case, we do want to extend vote if the height is equal or
	// greater than VoteExtensionsEnableHeight. This defers from the check done
	// in ValidateVoteExtensions and PrepareProposal in which we'll check for
	// vote extensions on VoteExtensionsEnableHeight+1.
	// Since Abci was deprecated, should check both Feature & Abci
	extsEnabled := cp.Feature.VoteExtensionsEnableHeight != nil && req.Height >= cp.Feature.VoteExtensionsEnableHeight.Value && cp.Feature.VoteExtensionsEnableHeight.Value != 0
	if !extsEnabled {
		// check abci params
		extsEnabled = cp.Abci != nil && req.Height >= cp.Abci.VoteExtensionsEnableHeight && cp.Abci.VoteExtensionsEnableHeight != 0
		if !extsEnabled {
			return nil, fmt.Errorf("vote extensions are not enabled; unexpected call to ExtendVote at height %d", req.Height)
		}
	}

	if c.extendVote == nil {
		return nil, errors.New("vote extensions are enabled but no extend function was set")
	}

	_, latestStore, err := c.store.StateLatest()
	if err != nil {
		return nil, err
	}

	resp, err := c.extendVote(ctx, latestStore, req)
	if err != nil {
		c.logger.Error("failed to extend vote", "height", req.Height, "err", err)
		return &abciproto.ExtendVoteResponse{}, nil
	}

	return resp, err
}

func decodeTxs[T transaction.Tx](rawTxs [][]byte, codec transaction.Codec[T]) ([]T, error) {
	txs := make([]T, len(rawTxs))
	for i, rawTx := range rawTxs {
		tx, err := codec.Decode(rawTx)
		if err != nil {
			return nil, fmt.Errorf("unable to decode tx: %d: %w", i, err)
		}
		txs[i] = tx
	}
	return txs, nil
}
