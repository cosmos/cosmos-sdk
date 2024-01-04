package cometbft

import (
	"context"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	coreappmgr "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/stf/mock"
	abci "github.com/cometbft/cometbft/abci/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/protobuf/proto"
)

var _ abci.Application = (*Consensus[mock.Tx])(nil)

func NewConsensus[T transaction.Tx](app appmanager.AppManager[T]) *Consensus[T] {
	return &Consensus[T]{
		app: app,
	}
}

type Consensus[T transaction.Tx] struct {
	app    appmanager.AppManager[T]
	logger log.Logger

	name string

	currentChangeSet []store.ChangeSet
	currentHeight    int64
	appHashes        map[int64][]byte
}

// Comet specific methods
func (*Consensus[T]) GetBlockRetentionHeight(commitHeight int64) int64 {
	// TODO
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
		LastBlockAppHash: []byte{},                       // TODO: missing c.app.AppHash(), should we store a map height->hash?
	}, nil
}

var QueryPathBroadcastTx = "/cosmos.tx.v1beta1.Service/BroadcastTx"

// Query implements types.Application.
func (c *Consensus[T]) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	// reject special cases
	if req.Path == QueryPathBroadcastTx {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "can't route a broadcast tx message")
	}

	// TODO: make a func that turns RequestQuery into the proper query struct
	var req2 proto.Message
	res, err := c.app.Query(ctx, req2)
	if err != nil {
		return nil, err
	}

	return parseQueryResponse(res)
}

// InitChain implements types.Application.
func (c *Consensus[T]) InitChain(ctx context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
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

	return &abci.ResponseInitChain{
		// ConsensusParams: w.app.ConsensusParams(), // TODO: add consensus params
		Validators: nil,
		AppHash:    nil,
	}, nil
}

// PrepareProposal implements types.Application.
func (c *Consensus[T]) PrepareProposal(ctx context.Context, req *abci.RequestPrepareProposal) (resp *abci.ResponsePrepareProposal, err error) {
	if req.Height < 1 {
		return nil, errors.New("PrepareProposal called with invalid height")
	}

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

	maxTotalBlockSize := uint32(512) // TODO: make this configurable
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
		// TODO: log this error
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
	blockReq := coreappmgr.BlockRequest{
		Height:            uint64(req.Height),
		Time:              req.Time,
		Hash:              req.Hash,
		Txs:               req.Txs,
		ConsensusMessages: nil, // TODO: add misbehaviors and other relevant stuff
	}
	resp, changeSet, err := c.app.DeliverBlock(ctx, blockReq)

	// TODO: make these concurrent safe, maybe with a height -> changeset map
	c.currentHeight = int64(req.Height)
	c.currentChangeSet = changeSet
	c.appHashes[req.Height] = resp.Apphash

	return parseFinalizeBlockResponse(resp, err)
}

// Commit implements types.Application.
func (c *Consensus[T]) Commit(ctx context.Context, _ *abci.RequestCommit) (*abci.ResponseCommit, error) {
	// TODO: add a way to get the last block height here + the changes
	_, err := c.app.CommitBlock(ctx, c.app.LastBlockHeight(), nil)
	return &abci.ResponseCommit{
		RetainHeight: c.GetBlockRetentionHeight(0),
	}, err
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

// snapshots

// ApplySnapshotChunk implements types.Application.
func (*Consensus[T]) ApplySnapshotChunk(context.Context, *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	panic("unimplemented")
}

// ListSnapshots implements types.Application.
func (*Consensus[T]) ListSnapshots(context.Context, *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	panic("unimplemented")
}

// LoadSnapshotChunk implements types.Application.
func (*Consensus[T]) LoadSnapshotChunk(context.Context, *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	panic("unimplemented")
}

// OfferSnapshot implements types.Application.
func (*Consensus[T]) OfferSnapshot(context.Context, *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	panic("unimplemented")
}
