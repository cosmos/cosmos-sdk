package cometbft

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	corecomet "cosmossdk.io/core/comet"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft/types"
	coreappmgr "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/stf/mock"

	// "cosmossdk.io/server/v2/streaming"

	cometerrors "cosmossdk.io/server/v2/cometbft/types/errors"
	snapshottypes "cosmossdk.io/store/v2/snapshots/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"
)

// Supported ABCI Query prefixes and paths
const (
	QueryPathApp    = "app"
	QueryPathCustom = "custom"
	QueryPathP2P    = "p2p"
	QueryPathStore  = "store"

	QueryPathBroadcastTx = "/cosmos.tx.v1beta1.Service/BroadcastTx"
)

var _ abci.Application = (*Consensus[mock.Tx])(nil)

func NewConsensus[T transaction.Tx](app appmanager.AppManager[T], cfg Config) *Consensus[T] {
	return &Consensus[T]{
		app: app,
		cfg: cfg,
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
	app     appmanager.AppManager[T]
	cfg     Config
	store   store.Store
	logger  log.Logger
	txCodec transaction.Codec[T]

	// this is only available after this node has committed a block (in FinalizeBlock),
	// otherwise it will be empty and we will need to query the app for the last
	// committed block.
	lastCommittedBlock atomic.Pointer[BlockData]
}

// CheckTx implements types.Application.
func (c *Consensus[T]) CheckTx(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
	resp, err := c.app.Simulate(ctx, req.Tx) // TODO: is it ok to use Simulate here?
	if err != nil {
		return nil, err
	}

	cometResp := &abci.CheckTxResponse{
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
func (c *Consensus[T]) Info(context.Context, *abci.InfoRequest) (*abci.InfoResponse, error) {
	version, _, err := c.store.StateLatest()
	if err != nil {
		return nil, err
	}

	cp, err := c.GetConsensusParams()
	if err != nil {
		return nil, err
	}

	return &abci.InfoResponse{
		Data:            c.cfg.Name,
		Version:         c.cfg.Version,
		AppVersion:      cp.GetVersion().App,
		LastBlockHeight: int64(version),
		// LastBlockAppHash: c.app.LastCommittedBlockHash(), // TODO: implement this on store. It's required by CometBFT
	}, nil
}

// Query implements types.Application.
func (c *Consensus[T]) Query(ctx context.Context, req *abci.QueryRequest) (*abci.QueryResponse, error) {
	// reject special cases
	if req.Path == QueryPathBroadcastTx {
		return QueryResult(errorsmod.Wrap(cometerrors.ErrInvalidRequest, "can't route a broadcast tx message"), c.cfg.Trace), nil
	}

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
	// TODO: won't work for now
	return &abci.InitChainResponse{
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
func (c *Consensus[T]) PrepareProposal(ctx context.Context, req *abci.PrepareProposalRequest) (resp *abci.PrepareProposalResponse, err error) {
	if req.Height < 1 {
		return nil, errors.New("PrepareProposal called with invalid height")
	}

	cp, err := c.GetConsensusParams()
	if err != nil {
		return nil, err
	}

	txs, err := c.app.BuildBlock(ctx, uint64(req.Height), uint32(cp.Block.MaxBytes))
	if err != nil {
		return nil, err
	}

	encodedTxs := make([][]byte, len(txs))
	for i, tx := range txs {
		encodedTxs[i] = tx.Encode()
	}

	return &abci.PrepareProposalResponse{
		Txs: encodedTxs,
	}, nil
}

// ProcessProposal implements types.Application.
func (c *Consensus[T]) ProcessProposal(ctx context.Context, req *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error) {
	decodedTxs := []T{}
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
		return &abci.ProcessProposalResponse{
			Status: abci.PROCESS_PROPOSAL_STATUS_REJECT,
		}, nil
	}

	return &abci.ProcessProposalResponse{
		Status: abci.PROCESS_PROPOSAL_STATUS_ACCEPT,
	}, nil
}

// FinalizeBlock implements types.Application.
func (c *Consensus[T]) FinalizeBlock(ctx context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
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

	blockReq := coreappmgr.BlockRequest{
		Height:            uint64(req.Height),
		Time:              req.Time,
		Hash:              req.Hash,
		Txs:               req.Txs,
		ConsensusMessages: []proto.Message{cometInfo},
	}

	resp, changeSet, err := c.app.DeliverBlock(ctx, blockReq)
	if err != nil {
		return nil, err
	}

	appHash, err := c.app.CommitBlock(ctx, blockReq.Height, changeSet)
	if err != nil {
		return nil, err
	}

	c.lastCommittedBlock.Store(&BlockData{
		Height:    int64(req.Height),
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
func (c *Consensus[T]) Commit(ctx context.Context, _ *abci.CommitRequest) (*abci.CommitResponse, error) {
	lastCommittedBlock := c.lastCommittedBlock.Load()

	c.cfg.SnapshotManager.SnapshotIfApplicable(lastCommittedBlock.Height)

	cp, err := c.GetConsensusParams()
	if err != nil {
		return nil, err
	}

	return &abci.CommitResponse{
		RetainHeight: c.GetBlockRetentionHeight(cp, lastCommittedBlock.Height),
	}, nil
}

// Vote extensions
// VerifyVoteExtension implements types.Application.
func (*Consensus[T]) VerifyVoteExtension(context.Context, *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	panic("unimplemented")
}

// ExtendVote implements types.Application.
func (*Consensus[T]) ExtendVote(context.Context, *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	panic("unimplemented")
}

// ApplySnapshotChunk implements types.Application.
func (c *Consensus[T]) ApplySnapshotChunk(_ context.Context, req *abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error) {
	if c.cfg.SnapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.ApplySnapshotChunkResponse{Result: abci.APPLY_SNAPSHOT_CHUNK_RESULT_ABORT}, nil
	}

	_, err := c.cfg.SnapshotManager.RestoreChunk(req.Chunk)
	switch {
	case err == nil:
		return &abci.ApplySnapshotChunkResponse{Result: abci.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrChunkHashMismatch):
		c.logger.Error(
			"chunk checksum mismatch; rejecting sender and requesting refetch",
			"chunk", req.Index,
			"sender", req.Sender,
			"err", err,
		)
		return &abci.ApplySnapshotChunkResponse{
			Result:        abci.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY,
			RefetchChunks: []uint32{req.Index},
			RejectSenders: []string{req.Sender},
		}, nil

	default:
		c.logger.Error("failed to restore snapshot", "err", err)
		return &abci.ApplySnapshotChunkResponse{Result: abci.APPLY_SNAPSHOT_CHUNK_RESULT_ABORT}, nil
	}
}

// ListSnapshots implements types.Application.
func (c *Consensus[T]) ListSnapshots(_ context.Context, ctx *abci.ListSnapshotsRequest) (resp *abci.ListSnapshotsResponse, err error) {
	if c.cfg.SnapshotManager == nil {
		return resp, nil
	}

	snapshots, err := c.cfg.SnapshotManager.List()
	if err != nil {
		c.logger.Error("failed to list snapshots", "err", err)
		return nil, err
	}

	for _, snapshot := range snapshots {
		abciSnapshot, err := snapshot.ToABCI()
		if err != nil {
			c.logger.Error("failed to convert ABCI snapshots", "err", err)
			return nil, err
		}

		resp.Snapshots = append(resp.Snapshots, &abciSnapshot)
	}

	return resp, nil
}

// LoadSnapshotChunk implements types.Application.
func (c *Consensus[T]) LoadSnapshotChunk(_ context.Context, req *abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error) {
	if c.cfg.SnapshotManager == nil {
		return &abci.LoadSnapshotChunkResponse{}, nil
	}

	chunk, err := c.cfg.SnapshotManager.LoadChunk(req.Height, req.Format, req.Chunk)
	if err != nil {
		c.logger.Error(
			"failed to load snapshot chunk",
			"height", req.Height,
			"format", req.Format,
			"chunk", req.Chunk,
			"err", err,
		)
		return nil, err
	}

	return &abci.LoadSnapshotChunkResponse{Chunk: chunk}, nil
}

// OfferSnapshot implements types.Application.
func (c *Consensus[T]) OfferSnapshot(_ context.Context, req *abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error) {
	if c.cfg.SnapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_ABORT}, nil
	}

	if req.Snapshot == nil {
		c.logger.Error("received nil snapshot")
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT}, nil
	}

	snapshot, err := snapshottypes.SnapshotFromABCI(req.Snapshot)
	if err != nil {
		c.logger.Error("failed to decode snapshot metadata", "err", err)
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT}, nil
	}

	err = c.cfg.SnapshotManager.Restore(snapshot)
	switch {
	case err == nil:
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrUnknownFormat):
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT_FORMAT}, nil

	case errors.Is(err, snapshottypes.ErrInvalidMetadata):
		c.logger.Error(
			"rejecting invalid snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_REJECT}, nil

	default:
		c.logger.Error(
			"failed to restore snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)

		// We currently don't support resetting the IAVL stores and retrying a
		// different snapshot, so we ask CometBFT to abort all snapshot restoration.
		return &abci.OfferSnapshotResponse{Result: abci.OFFER_SNAPSHOT_RESULT_ABORT}, nil
	}
}
