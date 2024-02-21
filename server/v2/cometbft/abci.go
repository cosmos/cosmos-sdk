package cometbft

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	abci "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	corecomet "cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/handlers"
	"cosmossdk.io/server/v2/cometbft/types"
	cometerrors "cosmossdk.io/server/v2/cometbft/types/errors"
	coreappmgr "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/streaming"
	"cosmossdk.io/store/v2/snapshots"
)

const (
	QueryPathApp   = "app"
	QueryPathP2P   = "p2p"
	QueryPathStore = "store"
)

var _ abci.Application = (*Consensus[transaction.Tx])(nil)

type (
	VerifyVoteExtensionFunc func(context.Context, store.ReaderMap, *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error)
	ExtendVoteFunc          func(context.Context, store.ReaderMap, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error)
)

type Consensus[T transaction.Tx] struct {
	app             appmanager.AppManager[T]
	cfg             Config
	store           types.Store
	logger          log.Logger
	txCodec         transaction.Codec[T]
	streaming       streaming.Manager
	snapshotManager *snapshots.Manager
	mempool         mempool.Mempool[T]

	verifyVoteExt VerifyVoteExtensionFunc
	extendVote    ExtendVoteFunc

	// this is only available after this node has committed a block (in FinalizeBlock),
	// otherwise it will be empty and we will need to query the app for the last
	// committed block. TODO(tip): check if concurrency is really needed
	lastCommittedBlock atomic.Pointer[BlockData]

	prepareProposalHandler handlers.PrepareHandler[T]
	processProposalHandler handlers.ProcessHandler[T]
}

func NewConsensus[T transaction.Tx](
	app appmanager.AppManager[T],
	mp mempool.Mempool[T],
	store types.Store,
	cfg Config,
) *Consensus[T] {
	return &Consensus[T]{
		mempool: mp,
		store:   store,
		app:     app,
		cfg:     cfg,
	}
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
	// TODO: evaluate here if to return error, or CheckTxResponse.error.
	decodedTx, err := c.txCodec.Decode(req.Tx)
	if err != nil {
		return nil, err
	}

	resp, err := c.app.ValidateTx(ctx, decodedTx)
	if err != nil {
		return nil, err
	}

	/* TODO insertion into the mempool, insertion should a cache tx,
	type CacheTx struct {
		// Tx is the transaction.
		Tx transaction.Tx
		// Encoded
		EncodedTx []byte
	}

	either do this in x/tx or here, but we need to avoid re-encoding the tx due to maliability
	*/

	cometResp := &abci.ResponseCheckTx{
		// Code:      resp.Code, //TODO: extract error code from resp.Error
		GasWanted: uint64ToInt64(resp.GasWanted),
		GasUsed:   uint64ToInt64(resp.GasUsed),
		Events:    intoABCIEvents(resp.Events, c.cfg.IndexEvents),
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

	cp, err := c.GetConsensusParams(ctx)
	if err != nil {
		return nil, err
	}

	// TODO use rootstore interface and that has LastCommitID

	return &abci.ResponseInfo{
		Data:            c.cfg.Name,
		Version:         c.cfg.Version,
		AppVersion:      cp.GetVersion().App,
		LastBlockHeight: int64(version),
		// LastBlockAppHash: c.store.LastCommittedID().Hash(), // TODO: implement this on store. It's required by CometBFT
	}, nil
}

// Query implements types.Application.
// It is called by cometbft to query application state.
func (c *Consensus[T]) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	appreq, err := parseQueryRequest(req)
	// if no error is returned then we can handle the query with the appmanager
	// otherwise it is a KV store query
	if err == nil {
		res, err := c.app.Query(ctx, uint64(req.Height), appreq)
		if err != nil {
			return nil, err
		}

		return queryResponse(req, res)
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

	// TODO: won't work for now
	return &abci.ResponseInitChain{
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
		AppHash:         []byte{},
	}, nil

	// valUpdates := []validator.Update{}
	// for _, v := range req.Validators {
	// 	pubkey, err := cryptocdc.FromCmtProtoPublicKey(v.PubKey)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	valUpdates = append(valUpdates, validator.Update{
	// 		PubKey: pubkey.Bytes(),
	// 		Power:  v.Power,
	// 	})
	// }

	// rr := appmanager.RequestInitChain{
	// 	Time:          req.Time,
	// 	ChainId:       req.ChainId,
	// 	AppStateBytes: req.AppStateBytes,
	// 	InitialHeight: req.InitialHeight,
	// 	Validators:    valUpdates,
	// }

	// res, err := c.app.InitChain(ctx, rr)
	// if err != nil {
	// 	return nil, err
	// }

	// abciVals := make(abci.ValidatorUpdates, len(res.Validators))
	// for i, update := range res.Validators {
	// 	abciVals[i] = abci.ValidatorUpdate{
	// 		PubKey: cmtprotocrypto.PublicKey{
	// 			Sum: &cmtprotocrypto.PublicKey_Ed25519{
	// 				Ed25519: update.PubKey,
	// 			},
	// 		},
	// 		Power: update.Power,
	// 	}
	// }

	// if len(req.Validators) > 0 {
	// 	if len(req.Validators) != len(abciVals) {
	// 		return nil, fmt.Errorf(
	// 			"len(RequestInitChain.Validators) != len(GenesisValidators) (%d != %d)",
	// 			len(req.Validators), len(abciVals),
	// 		)
	// 	}

	// 	sort.Sort(abci.ValidatorUpdates(req.Validators))
	// 	sort.Sort(abciVals)

	// 	for i := range abciVals {
	// 		if !proto.Equal(&abciVals[i], &req.Validators[i]) {
	// 			return nil, fmt.Errorf("genesisValidators[%d] != req.Validators[%d] ", i, i)
	// 		}
	// 	}
	// }
}

// PrepareProposal implements types.Application.
// It is called by cometbft to prepare a proposal block.
func (c *Consensus[T]) PrepareProposal(ctx context.Context, req *abci.RequestPrepareProposal) (resp *abci.ResponsePrepareProposal, err error) {
	if req.Height < 1 {
		return nil, errors.New("PrepareProposal called with invalid height")
	}

	txs, err := c.prepareProposalHandler(ctx, c.app, req)
	if err != nil {
		return nil, err
	}

	// TODO add bytes method in x/tx or cachetx
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
func (c *Consensus[T]) ProcessProposal(ctx context.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
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

	err = c.app.VerifyBlock(ctx, uint64(req.Height), decodedTxs)
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
func (c *Consensus[T]) FinalizeBlock(ctx context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	if err := c.validateFinalizeBlockHeight(req); err != nil {
		return nil, err
	}

	if err := c.checkHalt(req.Height, req.Time); err != nil {
		return nil, err
	}

	// for passing consensus info as a consensus message
	cometInfo := &types.ConsensusInfo{
		Info: corecomet.Info{
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

	blockReq := &coreappmgr.BlockRequest[T]{
		Height:            uint64(req.Height),
		Time:              req.Time,
		Hash:              req.Hash,
		Txs:               decodedTxs,
		ConsensusMessages: []proto.Message{cometInfo},
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
	appHash, err := c.store.StateCommit(stateChanges)
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
		return nil, fmt.Errorf("unable to remove txs: %w", err) // TODO: evaluate what erroring means here, and if we should even error.
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
func (c *Consensus[T]) VerifyVoteExtension(ctx context.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
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
