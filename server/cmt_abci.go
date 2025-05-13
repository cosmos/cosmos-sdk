package server

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

type cometABCIWrapper struct {
	app servertypes.ABCI
}

func NewCometABCIWrapper(app servertypes.ABCI) abci.Application {
	return cometABCIWrapper{app: app}
}

func (w cometABCIWrapper) Info(_ context.Context, req *abci.InfoRequest) (*abci.InfoResponse, error) {
	return w.app.Info(req)
}

func (w cometABCIWrapper) Query(ctx context.Context, req *abci.QueryRequest) (*abci.QueryResponse, error) {
	return w.app.Query(ctx, req)
}

func (w cometABCIWrapper) CheckTx(_ context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
	return w.app.CheckTx(req)
}

func (w cometABCIWrapper) InitChain(_ context.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	return w.app.InitChain(req)
}

func (w cometABCIWrapper) PrepareProposal(_ context.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	return w.app.PrepareProposal(req)
}

func (w cometABCIWrapper) ProcessProposal(_ context.Context, req *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error) {
	return w.app.ProcessProposal(req)
}

func (w cometABCIWrapper) FinalizeBlock(_ context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
	return w.app.FinalizeBlock(req)
}

func (w cometABCIWrapper) ExtendVote(ctx context.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	return w.app.ExtendVote(ctx, req)
}

func (w cometABCIWrapper) VerifyVoteExtension(_ context.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	return w.app.VerifyVoteExtension(req)
}

func (w cometABCIWrapper) Commit(_ context.Context, _ *abci.CommitRequest) (*abci.CommitResponse, error) {
	return w.app.Commit()
}

func (w cometABCIWrapper) ListSnapshots(_ context.Context, req *abci.ListSnapshotsRequest) (*abci.ListSnapshotsResponse, error) {
	return w.app.ListSnapshots(req)
}

func (w cometABCIWrapper) OfferSnapshot(_ context.Context, req *abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error) {
	return w.app.OfferSnapshot(req)
}

func (w cometABCIWrapper) LoadSnapshotChunk(_ context.Context, req *abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error) {
	return w.app.LoadSnapshotChunk(req)
}

func (w cometABCIWrapper) ApplySnapshotChunk(_ context.Context, req *abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error) {
	return w.app.ApplySnapshotChunk(req)
}
