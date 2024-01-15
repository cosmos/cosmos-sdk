package cometbft

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	abci "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	corecomet "cosmossdk.io/core/comet"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/types"
	cometerrors "cosmossdk.io/server/v2/cometbft/types/errors"
	coreappmgr "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/event"
	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/streaming"
)

const (
	QueryPathApp   = "app"
	QueryPathP2P   = "p2p"
	QueryPathStore = "store"
)

var _ abci.Application = (*Consensus[transaction.Tx])(nil)

func NewConsensus[T transaction.Tx](
	app appmanager.AppManager[T],
	mp mempool.Mempool[T],
	store store.Store,
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
	Height    int64
	Hash      []byte
	ChangeSet []store.ChangeSet
}

type Consensus[T transaction.Tx] struct {
	app       appmanager.AppManager[T]
	cfg       Config
	store     store.Store
	logger    log.Logger
	txCodec   transaction.Codec[T]
	streaming streaming.Manager
	mempool   mempool.Mempool[T]

	// this is only available after this node has committed a block (in FinalizeBlock),
	// otherwise it will be empty and we will need to query the app for the last
	// committed block.
	lastCommittedBlock atomic.Pointer[BlockData]
}

// CheckTx implements types.Application.
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

	cometResp := &abci.ResponseCheckTx{
		Code:      0,
		GasWanted: int64(resp.GasWanted),
		GasUsed:   int64(resp.GasUsed),
		Events:    intoABCIEvents(resp.Events, c.cfg.IndexEvents),
	}
	if resp.Error != nil {
		cometResp.Code = 1
		cometResp.Log = resp.Error.Error()
	}
	return cometResp, nil
}

// Info implements types.Application.
func (c *Consensus[T]) Info(context.Context, *abci.RequestInfo) (*abci.ResponseInfo, error) {
	version, _, err := c.store.StateLatest()
	if err != nil {
		return nil, err
	}

	cp, err := c.GetConsensusParams()
	if err != nil {
		return nil, err
	}

	return &abci.ResponseInfo{
		Data:            c.cfg.Name,
		Version:         c.cfg.Version,
		AppVersion:      cp.GetVersion().App,
		LastBlockHeight: int64(version),
		// LastBlockAppHash: c.app.LastCommittedBlockHash(), // TODO: implement this on store. It's required by CometBFT
	}, nil
}

// Query implements types.Application.
func (c *Consensus[T]) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	appreq, err := parseQueryRequest(req)
	if err == nil { // if no error is returned then we can handle the query with the appmanager
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
func (c *Consensus[T]) PrepareProposal(ctx context.Context, req *abci.RequestPrepareProposal) (resp *abci.ResponsePrepareProposal, err error) {
	if req.Height < 1 {
		return nil, errors.New("PrepareProposal called with invalid height")
	}

	cp, err := c.GetConsensusParams()
	if err != nil {
		return nil, err
	}

	txs, err := c.mempool.Get(ctx, int(cp.Block.MaxBytes))
	if err != nil {
		return nil, err
	}

	txs, err = c.app.BuildBlock(ctx, uint64(req.Height), txs)
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
func (c *Consensus[T]) ProcessProposal(ctx context.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	decodedTxs := make([]T, len(req.Txs))
	for _, tx := range req.Txs {
		decTx, err := c.txCodec.Decode(tx)
		if err != nil {
			// continue even if tx decoding fails
			c.logger.Error("failed to decode tx", "err", err)
		}
		decodedTxs = append(decodedTxs, decTx)
	}

	err := c.app.VerifyBlock(ctx, uint64(req.Height), decodedTxs)
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

	resp, changeSet, err := c.app.DeliverBlock(ctx, blockReq)
	if err != nil {
		return nil, err
	}

	// after we get the changeset we can produce the commit hash,
	// from the store.
	appHash, err := c.store.StateCommit(changeSet)
	if err != nil {
		return nil, fmt.Errorf("unable to commit the changeset: %w", err)
	}

	events := []event.Event{}
	events = append(events, resp.BeginBlockEvents...)
	events = append(events, resp.UpgradeBlockEvents...)
	for _, tx := range resp.TxResults {
		events = append(events, tx.Events...)
	}
	events = append(events, resp.EndBlockEvents...)

	// listen to state streaming changes in accordance with the block
	for _, streamingListener := range c.streaming.Listeners {
		if err := streamingListener.ListenDeliverBlock(ctx, streaming.ListenDeliverBlockRequest{
			BlockHeight: req.Height,
			// Txs:         req.Txs, TODO: see how to map txs
			Events: streaming.IntoStreamingEvents(events),
		}); err != nil {
			c.logger.Error("ListenDeliverBlock listening hook failed", "height", req.Height, "err", err)
		}

		strChangeSet := make([]*streaming.StoreKVPair, len(changeSet))
		for i, cs := range changeSet {
			strChangeSet[i] = &streaming.StoreKVPair{
				Key:    cs.Key,
				Value:  cs.Value,
				Delete: cs.Remove,
			}
		}

		if err := streamingListener.ListenStateChanges(ctx, strChangeSet); err != nil {
			c.logger.Error("ListenStateChanges listening hook failed", "height", req.Height, "err", err)
		}
	}

	// remove txs from the mempool
	err = c.mempool.Remove(decodedTxs)
	if err != nil {
		return nil, fmt.Errorf("unable to remove txs: %w", err) // TODO: evaluate what erroring means here, and if we should even error.
	}

	c.lastCommittedBlock.Store(&BlockData{
		Height:    req.Height,
		Hash:      appHash,
		ChangeSet: changeSet,
	})

	cp, err := c.GetConsensusParams()
	if err != nil {
		return nil, err
	}

	return finalizeBlockResponse(resp, cp, appHash, c.cfg.IndexEvents)
}

// Commit implements types.Application.
func (c *Consensus[T]) Commit(ctx context.Context, _ *abci.RequestCommit) (*abci.ResponseCommit, error) {
	lastCommittedBlock := c.lastCommittedBlock.Load()

	c.cfg.SnapshotManager.SnapshotIfApplicable(lastCommittedBlock.Height)

	cp, err := c.GetConsensusParams()
	if err != nil {
		return nil, err
	}

	return &abci.ResponseCommit{
		RetainHeight: c.GetBlockRetentionHeight(cp, lastCommittedBlock.Height),
	}, nil
}

// Vote extensions
// VerifyVoteExtension implements types.Application.
func (*Consensus[T]) VerifyVoteExtension(context.Context, *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	panic("unimplemented")
}

// ExtendVote implements types.Application.
func (*Consensus[T]) ExtendVote(context.Context, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	panic("unimplemented")
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
