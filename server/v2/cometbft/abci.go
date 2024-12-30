package cometbft

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	abci "github.com/cometbft/cometbft/abci/types"
	abciproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	"cosmossdk.io/server/v2/cometbft/handlers"
	"cosmossdk.io/server/v2/cometbft/mempool"
	"cosmossdk.io/server/v2/cometbft/oe"
	"cosmossdk.io/server/v2/cometbft/types"
	cometerrors "cosmossdk.io/server/v2/cometbft/types/errors"
	"cosmossdk.io/server/v2/streaming"
	"cosmossdk.io/store/v2/snapshots"
	consensustypes "cosmossdk.io/x/consensus/types"
)

const (
	QueryPathApp   = "app"
	QueryPathP2P   = "p2p"
	QueryPathStore = "store"
)

var _ abci.Application = (*consensus[transaction.Tx])(nil)

// consensus contains the implementation of the ABCI interface for CometBFT.
type consensus[T transaction.Tx] struct {
	logger           log.Logger
	appName, version string
	app              appmanager.AppManager[T]
	store            types.Store
	listener         *appdata.Listener
	snapshotManager  *snapshots.Manager
	streamingManager streaming.Manager
	mempool          mempool.Mempool[T]
	appCodecs        AppCodecs[T]

	cfg               Config
	chainID           string
	indexedABCIEvents map[string]struct{}

	initialHeight uint64
	// this is only available after this node has committed a block (in FinalizeBlock),
	// otherwise it will be empty and we will need to query the app for the last
	// committed block.
	lastCommittedHeight atomic.Int64

	prepareProposalHandler handlers.PrepareHandler[T]
	processProposalHandler handlers.ProcessHandler[T]
	verifyVoteExt          handlers.VerifyVoteExtensionHandler
	extendVote             handlers.ExtendVoteHandler
	checkTxHandler         handlers.CheckTxHandler[T]

	// optimisticExec contains the context required for Optimistic Execution,
	// including the goroutine handling.This is experimental and must be enabled
	// by developers.
	optimisticExec *oe.OptimisticExecution[T]

	addrPeerFilter types.PeerFilter // filter peers by address and port
	idPeerFilter   types.PeerFilter // filter peers by node ID

	queryHandlersMap map[string]appmodulev2.Handler
	getProtoRegistry func() (*protoregistry.Files, error)
	cfgMap           server.ConfigMap
}

// CheckTx implements types.Application.
// It is called by cometbft to verify transaction validity
func (c *consensus[T]) CheckTx(ctx context.Context, req *abciproto.CheckTxRequest) (*abciproto.CheckTxResponse, error) {
	decodedTx, err := c.appCodecs.TxCodec.Decode(req.Tx)
	if err != nil {
		return nil, err
	}

	if c.checkTxHandler == nil {
		resp, err := c.app.ValidateTx(ctx, decodedTx)
		// we do not want to return a cometbft error, but a check tx response with the error
		if err != nil && !errors.Is(err, resp.Error) {
			return nil, err
		}

		events := make([]abci.Event, 0)
		if !c.cfg.AppTomlConfig.DisableABCIEvents {
			events, err = intoABCIEvents(
				resp.Events,
				c.indexedABCIEvents,
				c.cfg.AppTomlConfig.DisableIndexABCIEvents,
			)
			if err != nil {
				return nil, err
			}
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
func (c *consensus[T]) Info(ctx context.Context, _ *abciproto.InfoRequest) (*abciproto.InfoResponse, error) {
	version, _, err := c.store.StateLatest()
	if err != nil {
		return nil, err
	}

	// if height is 0, we dont know the consensus params
	var appVersion uint64 = 0
	if version > 0 {
		cp, err := GetConsensusParams(ctx, c.app)
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
func (c *consensus[T]) Query(ctx context.Context, req *abciproto.QueryRequest) (resp *abciproto.QueryResponse, err error) {
	resp, isGRPC, err := c.maybeRunGRPCQuery(ctx, req)
	if isGRPC {
		return resp, err
	}

	// when a client did not provide a query height, manually inject the latest
	// for modules queries, AppManager does it automatically
	if req.Height == 0 {
		latestVersion, err := c.store.GetLatestVersion()
		if err != nil {
			return nil, err
		}
		req.Height = int64(latestVersion)
	}

	// this error most probably means that we can't handle it with a proto message, so
	// it must be an app/p2p/store query
	path := splitABCIQueryPath(req.Path)
	if len(path) == 0 {
		return queryResult(errorsmod.Wrap(cometerrors.ErrUnknownRequest, "no query path provided"), c.cfg.AppTomlConfig.Trace), nil
	}

	switch path[0] {
	case QueryPathApp:
		resp, err = c.handleQueryApp(ctx, path, req)

	case QueryPathStore:
		resp, err = c.handleQueryStore(path, req)

	case QueryPathP2P:
		resp, err = c.handleQueryP2P(path)

	default:
		resp = queryResult(errorsmod.Wrapf(cometerrors.ErrUnknownRequest, "unknown query path %s", req.Path), c.cfg.AppTomlConfig.Trace)
	}

	if err != nil {
		return queryResult(err, c.cfg.AppTomlConfig.Trace), nil
	}

	return resp, nil
}

func (c *consensus[T]) maybeRunGRPCQuery(ctx context.Context, req *abci.QueryRequest) (resp *abciproto.QueryResponse, isGRPC bool, err error) {
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

	var handlerFullName string
	md, isGRPC := desc.(protoreflect.MethodDescriptor)
	if !isGRPC {
		handlerFullName = string(desc.FullName())
	} else {
		handlerFullName = string(md.Input().FullName())
	}

	// special case for non-module services as they are external gRPC registered on the grpc server component
	// and not on the app itself, so it won't pass the router afterwards.

	externalResp, err := c.maybeHandleExternalServices(ctx, req)
	if err != nil {
		return nil, true, err
	} else if externalResp != nil {
		resp, err = queryResponse(externalResp, req.Height)
		return resp, true, err
	}

	handler, found := c.queryHandlersMap[handlerFullName]
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
		resp := gRPCErrorToSDKError(err)
		resp.Height = req.Height
		return resp, true, nil
	}

	resp, err = queryResponse(res, req.Height)
	return resp, true, err
}

// InitChain implements types.Application.
func (c *consensus[T]) InitChain(ctx context.Context, req *abciproto.InitChainRequest) (*abciproto.InitChainResponse, error) {
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

	blockResponse, genesisState, err := c.app.InitGenesis(
		ctx,
		br,
		req.AppStateBytes,
		c.appCodecs.TxCodec)
	if err != nil {
		return nil, fmt.Errorf("genesis state init failure: %w", err)
	}

	for _, txRes := range blockResponse.TxResults {
		if err := txRes.Error; err != nil {
			space, code, txLog := errorsmod.ABCIInfo(err, c.cfg.AppTomlConfig.Trace)
			c.logger.Warn("genesis tx failed", "codespace", space, "code", code, "log", txLog)
		}
	}

	validatorUpdates := intoABCIValidatorUpdates(blockResponse.ValidatorUpdates)

	if err := c.store.SetInitialVersion(uint64(req.InitialHeight - 1)); err != nil {
		return nil, fmt.Errorf("failed to set initial version: %w", err)
	}

	stateChanges, err := genesisState.GetStateChanges()
	if err != nil {
		return nil, err
	}
	cs := &store.Changeset{
		Version: uint64(req.InitialHeight - 1),
		Changes: stateChanges,
	}
	stateRoot, err := c.store.Commit(cs)
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
func (c *consensus[T]) PrepareProposal(
	ctx context.Context,
	req *abciproto.PrepareProposalRequest,
) (resp *abciproto.PrepareProposalResponse, err error) {
	if req.Height < 1 {
		return nil, errors.New("PrepareProposal called with invalid height")
	}

	if c.prepareProposalHandler == nil {
		return nil, errors.New("no prepare proposal function was set")
	}

	// Abort any running OE so it cannot overlap with `PrepareProposal`. This could happen if optimistic
	// `internalFinalizeBlock` from previous round takes a long time, but consensus has moved on to next round.
	// Overlap is undesirable, since `internalFinalizeBlock` and `PrepareProposal` could share access to
	// in-memory structs depending on application implementation.
	// No-op if OE is not enabled.
	// Similar call to Abort() is done in `ProcessProposal`.
	c.optimisticExec.Abort()

	ciCtx := contextWithCometInfo(ctx, comet.Info{
		Evidence:        toCoreEvidence(req.Misbehavior),
		ValidatorsHash:  req.NextValidatorsHash,
		ProposerAddress: req.ProposerAddress,
		LastCommit:      toCoreExtendedCommitInfo(req.LocalLastCommit),
	})

	txs, err := c.prepareProposalHandler(ciCtx, c.app, c.appCodecs.TxCodec, req, c.chainID)
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
func (c *consensus[T]) ProcessProposal(
	ctx context.Context,
	req *abciproto.ProcessProposalRequest,
) (*abciproto.ProcessProposalResponse, error) {
	if req.Height < 1 {
		return nil, errors.New("ProcessProposal called with invalid height")
	}

	if c.processProposalHandler == nil {
		return nil, errors.New("no process proposal function was set")
	}

	// Since the application can get access to FinalizeBlock state and write to it,
	// we must be sure to reset it in case ProcessProposal timeouts and is called
	// again in a subsequent round. However, we only want to do this after we've
	// processed the first block, as we want to avoid overwriting the finalizeState
	// after state changes during InitChain.
	if req.Height > int64(c.initialHeight) {
		// abort any running OE
		c.optimisticExec.Abort()
	}

	ciCtx := contextWithCometInfo(ctx, comet.Info{
		Evidence:        toCoreEvidence(req.Misbehavior),
		ValidatorsHash:  req.NextValidatorsHash,
		ProposerAddress: req.ProposerAddress,
		LastCommit:      toCoreCommitInfo(req.ProposedLastCommit),
	})

	err := c.processProposalHandler(ciCtx, c.app, c.appCodecs.TxCodec, req, c.chainID)
	if err != nil {
		c.logger.Error("failed to process proposal", "height", req.Height, "time", req.Time, "hash", fmt.Sprintf("%X", req.Hash), "err", err)
		return &abciproto.ProcessProposalResponse{
			Status: abciproto.PROCESS_PROPOSAL_STATUS_REJECT,
		}, nil
	}

	// Only execute optimistic execution if the proposal is accepted, OE is
	// enabled and the block height is greater than the initial height. During
	// the first block we'll be carrying state from InitChain, so it would be
	// impossible for us to easily revert.
	// After the first block has been processed, the next blocks will get executed
	// optimistically, so that when the ABCI client calls `FinalizeBlock` the app
	// can have a response ready.
	if req.Height > int64(c.initialHeight) {
		c.optimisticExec.Execute(req)
	}

	return &abciproto.ProcessProposalResponse{
		Status: abciproto.PROCESS_PROPOSAL_STATUS_ACCEPT,
	}, nil
}

// FinalizeBlock implements types.Application.
// It is called by cometbft to finalize a block.
func (c *consensus[T]) FinalizeBlock(
	ctx context.Context,
	req *abciproto.FinalizeBlockRequest,
) (*abciproto.FinalizeBlockResponse, error) {
	var (
		resp       *server.BlockResponse
		newState   store.WriterMap
		decodedTxs []T
		err        error
	)

	if c.optimisticExec.Initialized() {
		// check if the hash we got is the same as the one we are executing
		aborted := c.optimisticExec.AbortIfNeeded(req.Hash)

		// Wait for the OE to finish, regardless of whether it was aborted or not
		res, optimistErr := c.optimisticExec.WaitResult()

		if !aborted {
			if res != nil {
				resp = res.Resp
				newState = res.StateChanges
				decodedTxs = res.DecodedTxs
			}

			if optimistErr != nil {
				return nil, optimistErr
			}
		}

		c.optimisticExec.Reset()
	}

	if resp == nil { // if we didn't run OE, run the normal finalize block
		resp, newState, decodedTxs, err = c.internalFinalizeBlock(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	// after we get the changeset we can produce the commit hash,
	// from the store.
	stateChanges, err := newState.GetStateChanges()
	if err != nil {
		return nil, err
	}
	appHash, err := c.store.Commit(&store.Changeset{Version: uint64(req.Height), Changes: stateChanges})
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
	err = c.streamDeliverBlockChanges(ctx, req.Height, req.Txs, decodedTxs, resp.TxResults, events, stateChanges)
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

	cp, err := GetConsensusParams(ctx, c.app) // we get the consensus params from the latest state because we committed state above
	if err != nil {
		return nil, err
	}

	return finalizeBlockResponse(
		resp,
		cp,
		appHash,
		c.indexedABCIEvents,
		c.cfg.AppTomlConfig,
	)
}

func (c *consensus[T]) internalFinalizeBlock(
	ctx context.Context,
	req *abciproto.FinalizeBlockRequest,
) (*server.BlockResponse, store.WriterMap, []T, error) {
	if err := c.validateFinalizeBlockHeight(req); err != nil {
		return nil, nil, nil, err
	}

	if err := c.checkHalt(req.Height, req.Time); err != nil {
		return nil, nil, nil, err
	}

	// TODO(tip): can we expect some txs to not decode? if so, what we do in this case? this does not seem to be the case,
	// considering that prepare and process always decode txs, assuming they're the ones providing txs we should never
	// have a tx that fails decoding.
	decodedTxs, err := decodeTxs(c.logger, req.Txs, c.appCodecs.TxCodec)
	if err != nil {
		return nil, nil, nil, err
	}

	cid, err := c.store.LastCommitID()
	if err != nil {
		return nil, nil, nil, err
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

	resp, stateChanges, err := c.app.DeliverBlock(ciCtx, blockReq)

	return resp, stateChanges, decodedTxs, err
}

// Commit implements types.Application.
// It is called by cometbft to notify the application that a block was committed.
func (c *consensus[T]) Commit(ctx context.Context, _ *abciproto.CommitRequest) (*abciproto.CommitResponse, error) {
	lastCommittedHeight := c.lastCommittedHeight.Load()

	c.snapshotManager.SnapshotIfApplicable(lastCommittedHeight)

	cp, err := GetConsensusParams(ctx, c.app)
	if err != nil {
		return nil, err
	}

	return &abci.CommitResponse{
		RetainHeight: c.GetBlockRetentionHeight(cp, lastCommittedHeight),
	}, nil
}

// Vote extensions

// VerifyVoteExtension implements types.Application.
func (c *consensus[T]) VerifyVoteExtension(
	ctx context.Context,
	req *abciproto.VerifyVoteExtensionRequest,
) (*abciproto.VerifyVoteExtensionResponse, error) {
	// If vote extensions are not enabled, as a safety precaution, we return an
	// error.
	cp, err := GetConsensusParams(ctx, c.app)
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
func (c *consensus[T]) ExtendVote(ctx context.Context, req *abciproto.ExtendVoteRequest) (*abciproto.ExtendVoteResponse, error) {
	// If vote extensions are not enabled, as a safety precaution, we return an
	// error.
	cp, err := GetConsensusParams(ctx, c.app)
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

func decodeTxs[T transaction.Tx](logger log.Logger, rawTxs [][]byte, codec transaction.Codec[T]) ([]T, error) {
	txs := make([]T, len(rawTxs))
	for i, rawTx := range rawTxs {
		tx, err := codec.Decode(rawTx)
		if err != nil {
			// do not return an error here, as we want to deliver the block even if some txs are invalid
			logger.Debug("failed to decode tx", "err", err)
			txs[i] = RawTx(rawTx).(T) // allows getting the raw bytes down the line
			continue
		}
		txs[i] = tx
	}
	return txs, nil
}
