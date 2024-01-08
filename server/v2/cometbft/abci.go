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
	coreappmgr "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/stf/mock"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ abci.Application = (*Consensus[mock.Tx])(nil)

func NewConsensus[T transaction.Tx](app appmanager.AppManager[T]) *Consensus[T] {
	return &Consensus[T]{
		app: app,
	}
}

type CurrentBlock struct {
	Height    int64
	Hash      []byte
	ChangeSet []store.ChangeSet
}

type Consensus[T transaction.Tx] struct {
	app    appmanager.AppManager[T]
	logger log.Logger

	name string

	current atomic.Pointer[CurrentBlock]

	snapshotManager *snapshots.Manager
}

// TODO
func (*Consensus[T]) GetBlockRetentionHeight(commitHeight int64) int64 {
	return 0
}

// CheckTx implements types.Application.
func (c *Consensus[T]) CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	resp, err := c.app.Validate(ctx, req.Tx)
	if err != nil {
		return nil, err
	}
	cometResp := &abci.ResponseCheckTx{
		Code:      0,
		GasWanted: 0, // TODO: maybe appmanager.TxResult should include this
		GasUsed:   int64(resp.GasUsed),
		Events:    intoABCIEvents(resp.Events),
	}
	if resp.Error != nil {
		cometResp.Code = 1
		cometResp.Log = resp.Error.Error()
	}
	return cometResp, nil
}

// Info implements types.Application.
func (c *Consensus[T]) Info(context.Context, *abci.RequestInfo) (*abci.ResponseInfo, error) {
	// TODO: big TODO here

	return &abci.ResponseInfo{
		Data: c.name,
		// Version:          versionStr,
		// AppVersion:       appVersion,
		LastBlockHeight:  int64(c.app.LastBlockHeight()), // last committed block height
		LastBlockAppHash: []byte{},                       // TODO: missing apphash of the last committed block
	}, nil
}

var QueryPathBroadcastTx = "/cosmos.tx.v1beta1.Service/BroadcastTx"

// Query implements types.Application.
func (c *Consensus[T]) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	// reject special cases
	if req.Path == QueryPathBroadcastTx {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "can't route a broadcast tx message")
	}

	// TODO: Here we have to use the path to look for a grpc method through the proto v2 registry
	// (using FindDescriptorByName), then with the method descriptor we get the proper req and resp.
	appreq, err := parseQueryRequest(req)
	if err != nil {
		return nil, err
	}

	res, err := c.app.Query(ctx, appreq)
	if err != nil {
		return nil, err
	}

	return parseQueryResponse(req, res)
}

// InitChain implements types.Application.
func (c *Consensus[T]) InitChain(ctx context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	// TODO: won't work now
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

	// TODO: maybe we don't need this here and we handle panics in AppManager
	defer func() {
		if err := recover(); err != nil {
			c.logger.Error(
				"panic recovered in PrepareProposal",
				"height", req.Height,
				"time", req.Time,
				"panic", err,
			)

			resp = &abci.ResponsePrepareProposal{Txs: req.Txs}
		}
	}()

	maxTotalBlockSize := uint32(4096) // TODO: make this configurable
	txs, err := c.app.BuildBlock(ctx, uint64(req.Height), maxTotalBlockSize)
	if err != nil {
		return nil, err
	}

	fmt.Println("txs", txs)
	// TODO: convert []tx into [][]byte
	return &abci.ResponsePrepareProposal{}, nil
}

// ProcessProposal implements types.Application.
func (c *Consensus[T]) ProcessProposal(ctx context.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	err := c.app.VerifyBlock(ctx, uint64(req.Height), nil) // TODO: convert [][]byte into []tx
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

// check that ConsenusInfo (mock) implements proto.Message
var _ = proto.Message(&ConsensusInfo{})

type ConsensusInfo struct { // TODO: this is a mock, we need a proper proto.Message
	corecomet.Info
}

func (*ConsensusInfo) ProtoReflect() protoreflect.Message {
	panic("unimplemented")
}

// FinalizeBlock implements types.Application.
func (c *Consensus[T]) FinalizeBlock(ctx context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	// TODO: add validation over block height (validateFinalizeBlockHeight)

	cometInfo := &ConsensusInfo{
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

	// keep these values in memory so we can commit them later
	c.current.Store(&CurrentBlock{
		Height:    int64(req.Height),
		Hash:      req.Hash,
		ChangeSet: changeSet,
	})

	return parseFinalizeBlockResponse(resp, err)
}

// Commit implements types.Application.
func (c *Consensus[T]) Commit(ctx context.Context, _ *abci.RequestCommit) (*abci.ResponseCommit, error) {
	// get the block processed in FinalizeBlock
	currentState := c.current.Load()

	_, err := c.app.CommitBlock(ctx, uint64(currentState.Height), currentState.ChangeSet)
	if err != nil {
		return nil, err
	}

	c.current.Store(nil) // reset current block

	// TODO: add abci listener here and snapshotting
	c.snapshotManager.SnapshotIfApplicable(currentState.Height)

	return &abci.ResponseCommit{
		RetainHeight: c.GetBlockRetentionHeight(currentState.Height),
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

// snapshots (unchanged from baseapp's implementation)

// ApplySnapshotChunk implements types.Application.
func (c *Consensus[T]) ApplySnapshotChunk(_ context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	if c.snapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
	}

	_, err := c.snapshotManager.RestoreChunk(req.Chunk)
	switch {
	case err == nil:
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrChunkHashMismatch):
		c.logger.Error(
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
		c.logger.Error("failed to restore snapshot", "err", err)
		return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
	}
}

// ListSnapshots implements types.Application.
func (c *Consensus[T]) ListSnapshots(_ context.Context, ctx *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	resp := &abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{}}
	if c.snapshotManager == nil {
		return resp, nil
	}

	snapshots, err := c.snapshotManager.List()
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
func (c *Consensus[T]) LoadSnapshotChunk(_ context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	if c.snapshotManager == nil {
		return &abci.ResponseLoadSnapshotChunk{}, nil
	}

	chunk, err := c.snapshotManager.LoadChunk(req.Height, req.Format, req.Chunk)
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

	return &abci.ResponseLoadSnapshotChunk{Chunk: chunk}, nil
}

// OfferSnapshot implements types.Application.
func (c *Consensus[T]) OfferSnapshot(_ context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	if c.snapshotManager == nil {
		c.logger.Error("snapshot manager not configured")
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}, nil
	}

	if req.Snapshot == nil {
		c.logger.Error("received nil snapshot")
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil
	}

	// TODO: SnapshotFromABCI should be moved to this package or out of the SDK
	snapshot, err := snapshottypes.SnapshotFromABCI(req.Snapshot)
	if err != nil {
		c.logger.Error("failed to decode snapshot metadata", "err", err)
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil
	}

	err = c.snapshotManager.Restore(snapshot)
	switch {
	case err == nil:
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ACCEPT}, nil

	case errors.Is(err, snapshottypes.ErrUnknownFormat):
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT_FORMAT}, nil

	case errors.Is(err, snapshottypes.ErrInvalidMetadata):
		c.logger.Error(
			"rejecting invalid snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)
		return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}, nil

	default:
		c.logger.Error(
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
