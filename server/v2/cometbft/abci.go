package cometbft

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	abci "github.com/cometbft/cometbft/abci/types"

	coreappmgr "cosmossdk.io/core/app"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/log"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/handlers"
	"cosmossdk.io/server/v2/cometbft/mempool"
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

var _ abci.Application = (*Consensus[transaction.Tx])(nil)

type Consensus[T transaction.Tx] struct {
	app             *appmanager.AppManager[T]
	cfg             Config
	store           types.Store
	logger          log.Logger
	txCodec         transaction.Codec[T]
	streaming       streaming.Manager
	snapshotManager *snapshots.Manager
	mempool         mempool.Mempool[T]

	// this is only available after this node has committed a block (in FinalizeBlock),
	// otherwise it will be empty and we will need to query the app for the last
	// committed block. TODO(tip): check if concurrency is really needed
	lastCommittedBlock atomic.Pointer[BlockData]

	prepareProposalHandler handlers.PrepareHandler[T]
	processProposalHandler handlers.ProcessHandler[T]
	verifyVoteExt          handlers.VerifyVoteExtensionhandler
	extendVote             handlers.ExtendVoteHandler

	chainID string
}

func NewConsensus[T transaction.Tx](
	app *appmanager.AppManager[T],
	mp mempool.Mempool[T],
	store types.Store,
	cfg Config,
	txCodec transaction.Codec[T],
	logger log.Logger,
) *Consensus[T] {
	return &Consensus[T]{
		mempool: mp,
		store:   store,
		app:     app,
		cfg:     cfg,
		txCodec: txCodec,
		logger:  logger,
	}
}

func (c *Consensus[T]) SetMempool(mp mempool.Mempool[T]) {
	c.mempool = mp
}

func (c *Consensus[T]) SetStreamingManager(sm streaming.Manager) {
	c.streaming = sm
}

// SetSnapshotManager sets the snapshot manager for the Consensus.
// The snapshot manager is responsible for managing snapshots of the Consensus state.
// It allows for creating, storing, and restoring snapshots of the Consensus state.
// The provided snapshot manager will be used by the Consensus to handle snapshots.
func (c *Consensus[T]) SetSnapshotManager(sm *snapshots.Manager) {
	c.snapshotManager = sm
}

// RegisterExtensions registers the given extensions with the consensus module's snapshot manager.
// It allows additional snapshotter implementations to be used for creating and restoring snapshots.
func (c *Consensus[T]) RegisterExtensions(extensions ...snapshots.ExtensionSnapshotter) {
	if err := c.snapshotManager.RegisterExtensions(extensions...); err != nil {
		panic(fmt.Errorf("failed to register snapshot extensions: %w", err))
	}
}

func (c *Consensus[T]) SetPrepareProposalHandler(handler handlers.PrepareHandler[T]) {
	c.prepareProposalHandler = handler
}

func (c *Consensus[T]) SetProcessProposalHandler(handler handlers.ProcessHandler[T]) {
	c.processProposalHandler = handler
}

func (c *Consensus[T]) SetExtendVoteExtension(handler handlers.ExtendVoteHandler) {
	c.extendVote = handler
}

func (c *Consensus[T]) SetVerifyVoteExtension(handler handlers.VerifyVoteExtensionhandler) {
	c.verifyVoteExt = handler
}

// BlockData is used to keep some data about the last committed block. Currently
// we only use the height, the rest is not needed right now and might get removed
// in the future.
type BlockData struct {
	Height       int64
	Hash         []byte
	StateChanges []store.StateChanges
}

// CheckTx implements types.Application.
// It is called by cometbft to verify transaction validity
func (c *Consensus[T]) CheckTx(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
	decodedTx, err := c.txCodec.Decode(req.Tx)
	if err != nil {
		return nil, err
	}

	resp, err := c.app.ValidateTx(ctx, decodedTx)
	if err != nil {
		return nil, err
	}

	cometResp := &abci.CheckTxResponse{
		Code:      resp.Code,
		GasWanted: uint64ToInt64(resp.GasWanted),
		GasUsed:   uint64ToInt64(resp.GasUsed),
		Events:    intoABCIEvents(resp.Events, c.cfg.IndexEvents),
		Info:      resp.Info,
		Data:      resp.Data,
		Log:       resp.Log,
		Codespace: resp.Codespace,
	}
	if resp.Error != nil {
		cometResp.Code = 1
		cometResp.Log = resp.Error.Error()
	}
	return cometResp, nil
}

// Info implements types.Application.
func (c *Consensus[T]) Info(ctx context.Context, _ *abci.InfoRequest) (*abci.InfoResponse, error) {
	version, _, err := c.store.StateLatest()
	if err != nil {
		return nil, err
	}

	// cp, err := c.GetConsensusParams(ctx)
	// if err != nil {
	//	return nil, err
	// }

	cid, err := c.store.LastCommitID()
	if err != nil {
		return nil, err
	}

	return &abci.InfoResponse{
		Data:    c.cfg.Name,
		Version: c.cfg.Version,
		// AppVersion:       cp.GetVersion().App,
		AppVersion:       0, // TODO fetch from store?
		LastBlockHeight:  int64(version),
		LastBlockAppHash: cid.Hash,
	}, nil
}

// Query implements types.Application.
// It is called by cometbft to query application state.
func (c *Consensus[T]) Query(ctx context.Context, req *abci.QueryRequest) (*abci.QueryResponse, error) {
	// follow the query path from here
	decodedMsg, err := c.txCodec.Decode(req.Data)
	protoMsg, ok := any(decodedMsg).(transaction.Msg)
	if !ok {
		return nil, fmt.Errorf("decoded type T %T must implement core/transaction.Msg", decodedMsg)
	}

	// if no error is returned then we can handle the query with the appmanager
	// otherwise it is a KV store query
	if err == nil {
		res, err := c.app.Query(ctx, uint64(req.Height), protoMsg)
		if err != nil {
			return nil, err
		}

		return queryResponse(res)
	}

	// this error most probably means that we can't handle it with a proto message, so
	// it must be an app/p2p/store query
	path := splitABCIQueryPath(req.Path)
	if len(path) == 0 {
		return QueryResult(errorsmod.Wrap(cometerrors.ErrUnknownRequest, "no query path provided"), c.cfg.Trace), nil
	}

	var resp *abci.QueryResponse

	switch path[0] {
	case QueryPathApp:
		resp, err = c.handlerQueryApp(ctx, path, req)

	case QueryPathStore:
		resp, err = c.handleQueryStore(path, c.store, req)

	case QueryPathP2P:
		resp, err = c.handleQueryP2P(path)

	default:
		resp = QueryResult(errorsmod.Wrap(cometerrors.ErrUnknownRequest, "unknown query path"), c.cfg.Trace)
	}

	if err != nil {
		return QueryResult(err, c.cfg.Trace), nil
	}

	return resp, nil
}

// InitChain implements types.Application.
func (c *Consensus[T]) InitChain(ctx context.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	c.logger.Info("InitChain", "initialHeight", req.InitialHeight, "chainID", req.ChainId)

	// store chainID to be used later on in execution
	c.chainID = req.ChainId

	// On a new chain, we consider the init chain block height as 0, even though
	// req.InitialHeight is 1 by default.
	// TODO

	var consMessages []transaction.Msg
	if req.ConsensusParams != nil {
		consMessages = append(consMessages, &consensustypes.MsgUpdateParams{
			Authority: c.cfg.ConsensusAuthority,
			Block:     req.ConsensusParams.Block,
			Evidence:  req.ConsensusParams.Evidence,
			Validator: req.ConsensusParams.Validator,
			Abci:      req.ConsensusParams.Abci,
		})
	}

	br := &coreappmgr.BlockRequest[T]{
		Height:            uint64(req.InitialHeight - 1),
		Time:              req.Time,
		Hash:              nil,
		AppHash:           nil,
		ChainId:           req.ChainId,
		ConsensusMessages: consMessages,
		IsGenesis:         true,
	}

	blockresponse, genesisState, err := c.app.InitGenesis(
		ctx,
		br,
		req.AppStateBytes,
		c.txCodec)
	if err != nil {
		return nil, fmt.Errorf("genesis state init failure: %w", err)
	}

	// TODO necessary? where should this WARN live if it all. helpful for testing
	for _, txRes := range blockresponse.TxResults {
		if txRes.Error != nil {
			c.logger.Warn("genesis tx failed", "code", txRes.Code, "log", txRes.Log, "error", txRes.Error)
		}
	}

	validatorUpdates := intoABCIValidatorUpdates(blockresponse.ValidatorUpdates)

	stateChanges, err := genesisState.GetStateChanges()
	if err != nil {
		return nil, err
	}
	cs := &store.Changeset{
		Changes: stateChanges,
	}
	stateRoot, err := c.store.Commit(cs)
	if err != nil {
		return nil, fmt.Errorf("unable to commit the changeset: %w", err)
	}

	return &abci.InitChainResponse{
		ConsensusParams: req.ConsensusParams,
		Validators:      validatorUpdates,
		AppHash:         stateRoot,
	}, nil
}

// PrepareProposal implements types.Application.
// It is called by cometbft to prepare a proposal block.
func (c *Consensus[T]) PrepareProposal(
	ctx context.Context,
	req *abci.PrepareProposalRequest,
) (resp *abci.PrepareProposalResponse, err error) {
	if req.Height < 1 {
		return nil, errors.New("PrepareProposal called with invalid height")
	}

	decodedTxs := make([]T, len(req.Txs))
	for _, tx := range req.Txs {
		decTx, err := c.txCodec.Decode(tx)
		if err != nil {
			// TODO: vote extension meta data as a custom type to avoid possibly accepting invalid txs
			// continue even if tx decoding fails
			c.logger.Error("failed to decode tx", "err", err)
		}
		decodedTxs = append(decodedTxs, decTx)
	}

	ciCtx := contextWithCometInfo(ctx, comet.Info{
		Evidence:        toCoreEvidence(req.Misbehavior),
		ValidatorsHash:  req.NextValidatorsHash,
		ProposerAddress: req.ProposerAddress,
		LastCommit:      toCoreExtendedCommitInfo(req.LocalLastCommit),
	})

	txs, err := c.prepareProposalHandler(ciCtx, c.app, decodedTxs, req)
	if err != nil {
		return nil, err
	}

	encodedTxs := make([][]byte, len(txs))
	for i, tx := range txs {
		encodedTxs[i] = tx.Bytes()
	}

	return &abci.PrepareProposalResponse{
		Txs: encodedTxs,
	}, nil
}

// ProcessProposal implements types.Application.
// It is called by cometbft to process/verify a proposal block.
func (c *Consensus[T]) ProcessProposal(
	ctx context.Context,
	req *abci.ProcessProposalRequest,
) (*abci.ProcessProposalResponse, error) {
	decodedTxs := make([]T, len(req.Txs))
	for _, tx := range req.Txs {
		decTx, err := c.txCodec.Decode(tx)
		if err != nil {
			// TODO: vote extension meta data as a custom type to avoid possibly accepting invalid txs
			// continue even if tx decoding fails
			c.logger.Error("failed to decode tx", "err", err)
		}
		decodedTxs = append(decodedTxs, decTx)
	}

	ciCtx := contextWithCometInfo(ctx, comet.Info{
		Evidence:        toCoreEvidence(req.Misbehavior),
		ValidatorsHash:  req.NextValidatorsHash,
		ProposerAddress: req.ProposerAddress,
		LastCommit:      toCoreCommitInfo(req.ProposedLastCommit),
	})

	err := c.processProposalHandler(ciCtx, c.app, decodedTxs, req)
	if err != nil {
		c.logger.Error("failed to process proposal", "height", req.Height, "time", req.Time, "hash", fmt.Sprintf("%X", req.Hash), "err", err)
		return &abci.ProcessProposalResponse{
			Status: abci.PROCESS_PROPOSAL_STATUS_REJECT,
		}, nil
	}

	return &abci.ProcessProposalResponse{
		Status: abci.PROCESS_PROPOSAL_STATUS_ACCEPT,
	}, nil
}

// FinalizeBlock implements types.Application.
// It is called by cometbft to finalize a block.
func (c *Consensus[T]) FinalizeBlock(
	ctx context.Context,
	req *abci.FinalizeBlockRequest,
) (*abci.FinalizeBlockResponse, error) {
	if err := c.validateFinalizeBlockHeight(req); err != nil {
		return nil, err
	}

	if err := c.checkHalt(req.Height, req.Time); err != nil {
		return nil, err
	}

	// TODO evaluate this approach vs. service using context.
	// cometInfo := &consensustypes.MsgUpdateCometInfo{
	//	Authority: c.cfg.ConsensusAuthority,
	//	CometInfo: &consensustypes.CometInfo{
	//		Evidence:        req.Misbehavior,
	//		ValidatorsHash:  req.NextValidatorsHash,
	//		ProposerAddress: req.ProposerAddress,
	//		LastCommit:      req.DecidedLastCommit,
	//	},
	//}
	//
	// ctx = context.WithValue(ctx, corecontext.CometInfoKey, &comet.Info{
	// 	Evidence:        sdktypes.ToSDKEvidence(req.Misbehavior),
	// 	ValidatorsHash:  req.NextValidatorsHash,
	// 	ProposerAddress: req.ProposerAddress,
	// 	LastCommit:      sdktypes.ToSDKCommitInfo(req.DecidedLastCommit),
	// })

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

	blockReq := &coreappmgr.BlockRequest[T]{
		Height:  uint64(req.Height),
		Time:    req.Time,
		Hash:    req.Hash,
		AppHash: cid.Hash,
		ChainId: c.chainID,
		Txs:     decodedTxs,
		// ConsensusMessages: []transaction.Msg{cometInfo},
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
	err = c.mempool.Remove(decodedTxs)
	if err != nil {
		return nil, fmt.Errorf("unable to remove txs: %w", err)
	}

	c.lastCommittedBlock.Store(&BlockData{
		Height:       req.Height,
		Hash:         appHash,
		StateChanges: stateChanges,
	})

	cp, err := c.GetConsensusParams(ctx) // we get the consensus params from the latest state because we committed state above
	if err != nil {
		return nil, err
	}

	return finalizeBlockResponse(resp, cp, appHash, c.cfg.IndexEvents)
}

// Commit implements types.Application.
// It is called by cometbft to notify the application that a block was committed.
func (c *Consensus[T]) Commit(ctx context.Context, _ *abci.CommitRequest) (*abci.CommitResponse, error) {
	lastCommittedBlock := c.lastCommittedBlock.Load()

	c.snapshotManager.SnapshotIfApplicable(lastCommittedBlock.Height)

	cp, err := c.GetConsensusParams(ctx)
	if err != nil {
		return nil, err
	}

	return &abci.CommitResponse{
		RetainHeight: c.GetBlockRetentionHeight(cp, lastCommittedBlock.Height),
	}, nil
}

// Vote extensions
// VerifyVoteExtension implements types.Application.
func (c *Consensus[T]) VerifyVoteExtension(
	ctx context.Context,
	req *abci.VerifyVoteExtensionRequest,
) (*abci.VerifyVoteExtensionResponse, error) {
	// If vote extensions are not enabled, as a safety precaution, we return an
	// error.
	cp, err := c.GetConsensusParams(ctx)
	if err != nil {
		return nil, err
	}

	// Note: we verify votes extensions on VoteExtensionsEnableHeight+1. Check
	// comment in ExtendVote and ValidateVoteExtensions for more details.
	extsEnabled := cp.Abci != nil && req.Height >= cp.Abci.VoteExtensionsEnableHeight && cp.Abci.VoteExtensionsEnableHeight != 0
	if !extsEnabled {
		return nil, fmt.Errorf("vote extensions are not enabled; unexpected call to VerifyVoteExtension at height %d", req.Height)
	}

	if c.verifyVoteExt == nil {
		return nil, fmt.Errorf("vote extensions are enabled but no verify function was set")
	}

	_, latestStore, err := c.store.StateLatest()
	if err != nil {
		return nil, err
	}

	resp, err := c.verifyVoteExt(ctx, latestStore, req)
	if err != nil {
		c.logger.Error("failed to verify vote extension", "height", req.Height, "err", err)
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
	}

	return resp, err
}

// ExtendVote implements types.Application.
func (c *Consensus[T]) ExtendVote(ctx context.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
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
	extsEnabled := cp.Abci != nil && req.Height >= cp.Abci.VoteExtensionsEnableHeight && cp.Abci.VoteExtensionsEnableHeight != 0
	if !extsEnabled {
		return nil, fmt.Errorf("vote extensions are not enabled; unexpected call to ExtendVote at height %d", req.Height)
	}

	if c.verifyVoteExt == nil {
		return nil, fmt.Errorf("vote extensions are enabled but no verify function was set")
	}

	_, latestStore, err := c.store.StateLatest()
	if err != nil {
		return nil, err
	}

	resp, err := c.extendVote(ctx, latestStore, req)
	if err != nil {
		c.logger.Error("failed to verify vote extension", "height", req.Height, "err", err)
		return &abci.ExtendVoteResponse{}, nil
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
