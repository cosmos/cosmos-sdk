package abci

import (
	"context"

	"github.com/tendermint/tendermint/abci/types"
)

type HandlerBase struct{}

func (h HandlerBase) Info(context.Context, types.RequestInfo) types.ResponseInfo {
	return types.ResponseInfo{}
}

func (h HandlerBase) SetOption(context.Context, types.RequestSetOption) types.ResponseSetOption {
	return types.ResponseSetOption{}
}

func (h HandlerBase) Query(context.Context, types.RequestQuery) types.ResponseQuery {
	return types.ResponseQuery{}
}

func (h HandlerBase) CheckTx(context.Context, types.RequestCheckTx) types.ResponseCheckTx {
	return types.ResponseCheckTx{}
}

func (h HandlerBase) InitChain(context.Context, types.RequestInitChain) types.ResponseInitChain {
	return types.ResponseInitChain{}
}

func (h HandlerBase) BeginBlock(context.Context, types.RequestBeginBlock) types.ResponseBeginBlock {
	return types.ResponseBeginBlock{}
}

func (h HandlerBase) DeliverTx(context.Context, types.RequestDeliverTx) types.ResponseDeliverTx {
	return types.ResponseDeliverTx{}
}

func (h HandlerBase) EndBlock(context.Context, types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{}
}

func (h HandlerBase) Commit(context.Context) types.ResponseCommit {
	return types.ResponseCommit{}
}

func (h HandlerBase) ListSnapshots(context.Context, types.RequestListSnapshots) types.ResponseListSnapshots {
	return types.ResponseListSnapshots{}
}

func (h HandlerBase) OfferSnapshot(context.Context, types.RequestOfferSnapshot) types.ResponseOfferSnapshot {
	return types.ResponseOfferSnapshot{}
}

func (h HandlerBase) LoadSnapshotChunk(context.Context, types.RequestLoadSnapshotChunk) types.ResponseLoadSnapshotChunk {
	return types.ResponseLoadSnapshotChunk{}
}

func (h HandlerBase) ApplySnapshotChunk(context.Context, types.RequestApplySnapshotChunk) types.ResponseApplySnapshotChunk {
	return types.ResponseApplySnapshotChunk{}
}

var _ Handler = HandlerBase{}
