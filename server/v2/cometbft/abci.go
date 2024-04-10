package cometbft

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"cosmossdk.io/core/header"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	consensusv1 "cosmossdk.io/api/cosmos/consensus/v1"
	coreappmgr "cosmossdk.io/core/app"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/handlers"
	"cosmossdk.io/server/v2/cometbft/mempool"
	"cosmossdk.io/server/v2/cometbft/types"
	cometerrors "cosmossdk.io/server/v2/cometbft/types/errors"
	"cosmossdk.io/server/v2/streaming"
	"cosmossdk.io/store/v2/snapshots"
	abci "github.com/cometbft/cometbft/abci/types"
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
	c.snapshotManager.RegisterExtensions(extensions...)
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
func (c *Consensus[T]) CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	decodedTx, err := c.txCodec.Decode(req.Tx)
	if err != nil {
		return nil, err
	}

	resp, err := c.app.ValidateTx(ctx, decodedTx)
	if err != nil {
		return nil, err
	}

	cometResp := &abci.ResponseCheckTx{
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
func (c *Consensus[T]) Info(ctx context.Context, _ *abci.RequestInfo) (*abci.ResponseInfo, error) {
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

	return &abci.ResponseInfo{
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
func (c *Consensus[T]) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	// follow the query path from here
	decodedMsg, err := c.txCodec.Decode(req.Data)
	protoMsg, ok := any(decodedMsg).(transaction.Type)
	if !ok {
		return nil, fmt.Errorf("decoded type T %T must implement core/transaction.Type", decodedMsg)
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

	var resp *abci.ResponseQuery

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
func (c *Consensus[T]) InitChain(ctx context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	c.logger.Info("InitChain", "initialHeight", req.InitialHeight, "chainID", req.ChainId)

	c.chainID = req.ChainId

	// On a new chain, we consider the init chain block height as 0, even though
	// req.InitialHeight is 1 by default.
	// TODO

	var consMessages []transaction.Type
	if req.ConsensusParams != nil {
		consMessages = append(consMessages, &consensustypes.MsgUpdateParams{
			Authority: c.cfg.ConsensusAuthority,
			Block:     req.ConsensusParams.Block,
			Evidence:  req.ConsensusParams.Evidence,
			Validator: req.ConsensusParams.Validator,
			Abci:      req.ConsensusParams.Abci,
		})
	}

	genesisHeaderInfo := header.Info{
		Height:  req.InitialHeight,
		Hash:    nil,
		Time:    req.Time,
		ChainID: req.ChainId,
		AppHash: nil,
	}

	genesisState, err := c.app.InitGenesis(ctx, genesisHeaderInfo, consMessages, req.AppStateBytes)
	if err != nil {
		return nil, fmt.Errorf("genesis state init failure: %w", err)
	}

	println(genesisState) // TODO: this needs to be committed to store as height 0.

	// TODO: populate
	return &abci.ResponseInitChain{
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
		AppHash:         []byte{},
	}, nil
}

// PrepareProposal implements types.Application.
// It is called by cometbft to prepare a proposal block.
func (c *Consensus[T]) PrepareProposal(
	ctx context.Context,
	req *abci.RequestPrepareProposal,
) (resp *abci.ResponsePrepareProposal, err error) {
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

	txs, err := c.prepareProposalHandler(ctx, c.app, decodedTxs, req)
	if err != nil {
		return nil, err
	}

	encodedTxs := make([][]byte, len(txs))
	for i, tx := range txs {
		encodedTxs[i] = tx.Bytes()
	}

	return &abci.ResponsePrepareProposal{
		Txs: encodedTxs,
	}, nil
}

// ProcessProposal implements types.Application.
// It is called by cometbft to process/verify a proposal block.
func (c *Consensus[T]) ProcessProposal(
	ctx context.Context,
	req *abci.RequestProcessProposal,
) (*abci.ResponseProcessProposal, error) {
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

	err := c.processProposalHandler(ctx, c.app, decodedTxs, req)
	if err != nil {
		c.logger.Error("failed to process proposal", "height", req.Height, "time", req.Time, "hash", fmt.Sprintf("%X", req.Hash), "err", err)
		return &abci.ResponseProcessProposal{
			Status: abci.ResponseProcessProposal_REJECT,
		}, nil
	}

	return &abci.ResponseProcessProposal{
		Status: abci.ResponseProcessProposal_ACCEPT,
	}, nil
}

// FinalizeBlock implements types.Application.
// It is called by cometbft to finalize a block.
func (c *Consensus[T]) FinalizeBlock(
	ctx context.Context,
	req *abci.RequestFinalizeBlock,
) (*abci.ResponseFinalizeBlock, error) {
	if err := c.validateFinalizeBlockHeight(req); err != nil {
		return nil, err
	}

	if err := c.checkHalt(req.Height, req.Time); err != nil {
		return nil, err
	}

	// for passing consensus info as a consensus message
	cometInfo := &consensusv1.ConsensusMsgCometInfoRequest{
		Info: &consensusv1.CometInfo{
			Evidence:        ToSDKEvidence(req.Misbehavior),
			ValidatorsHash:  req.NextValidatorsHash,
			ProposerAddress: req.ProposerAddress,
			LastCommit:      ToSDKCommitInfo(req.DecidedLastCommit),
		},
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

	blockReq := &coreappmgr.BlockRequest[T]{
		Height:            uint64(req.Height),
		Time:              req.Time,
		Hash:              req.Hash,
		AppHash:           cid.Hash,
		ChainId:           c.chainID,
		Txs:               decodedTxs,
		ConsensusMessages: []transaction.Type{cometInfo},
	}

	resp, newState, err := c.app.DeliverBlock(ctx, blockReq)
	if err != nil {
		return nil, err
	}

	// after we get the changeset we can produce the commit hash,
	// from the store.
	stateChanges, err := newState.GetStateChanges()
	if err != nil {
		return nil, err
	}
	appHash, err := c.store.Commit(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to commit the changeset: %w", err)
	}

	events := []event.Event{}
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
func (c *Consensus[T]) Commit(ctx context.Context, _ *abci.RequestCommit) (*abci.ResponseCommit, error) {
	lastCommittedBlock := c.lastCommittedBlock.Load()

	c.snapshotManager.SnapshotIfApplicable(lastCommittedBlock.Height)

	cp, err := c.GetConsensusParams(ctx)
	if err != nil {
		return nil, err
	}

	return &abci.ResponseCommit{
		RetainHeight: c.GetBlockRetentionHeight(cp, lastCommittedBlock.Height),
	}, nil
}

// Vote extensions
// VerifyVoteExtension implements types.Application.
func (c *Consensus[T]) VerifyVoteExtension(
	ctx context.Context,
	req *abci.RequestVerifyVoteExtension,
) (*abci.ResponseVerifyVoteExtension, error) {
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
		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
	}

	return resp, err
}

// ExtendVote implements types.Application.
func (c *Consensus[T]) ExtendVote(ctx context.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
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
		return &abci.ResponseExtendVote{}, nil
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
