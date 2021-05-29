package abci

import (
	"context"

	"github.com/tendermint/tendermint/abci/types"
)

type AppBase struct {
	handler    Handler
	checkCtx   context.Context
	deliverCtx context.Context
}

var _ types.Application = AppBase{}

func (a AppBase) Info(info types.RequestInfo) types.ResponseInfo {
	return a.handler.Info(a.checkCtx, info)
}

func (a AppBase) SetOption(option types.RequestSetOption) types.ResponseSetOption {
	return a.handler.SetOption(a.checkCtx, option)
}

func (a AppBase) Query(query types.RequestQuery) types.ResponseQuery {
	return a.handler.Query(a.checkCtx, query)
}

func (a AppBase) CheckTx(tx types.RequestCheckTx) types.ResponseCheckTx {
	return a.handler.CheckTx(a.checkCtx, tx)
}

func (a AppBase) InitChain(chain types.RequestInitChain) types.ResponseInitChain {
	return a.handler.InitChain(a.deliverCtx, chain)
}

func (a AppBase) BeginBlock(block types.RequestBeginBlock) types.ResponseBeginBlock {
	return a.handler.BeginBlock(a.deliverCtx, block)
}

func (a AppBase) DeliverTx(tx types.RequestDeliverTx) types.ResponseDeliverTx {
	return a.handler.DeliverTx(a.deliverCtx, tx)
}

func (a AppBase) EndBlock(block types.RequestEndBlock) types.ResponseEndBlock {
	return a.handler.EndBlock(a.deliverCtx, block)
}

func (a AppBase) Commit() types.ResponseCommit {
	return a.handler.Commit(a.deliverCtx)
}

func (a AppBase) ListSnapshots(snapshots types.RequestListSnapshots) types.ResponseListSnapshots {
	return a.handler.ListSnapshots(a.checkCtx, snapshots)
}

func (a AppBase) OfferSnapshot(snapshot types.RequestOfferSnapshot) types.ResponseOfferSnapshot {
	return a.handler.OfferSnapshot(a.checkCtx, snapshot)
}

func (a AppBase) LoadSnapshotChunk(chunk types.RequestLoadSnapshotChunk) types.ResponseLoadSnapshotChunk {
	return a.handler.LoadSnapshotChunk(a.checkCtx, chunk)
}

func (a AppBase) ApplySnapshotChunk(chunk types.RequestApplySnapshotChunk) types.ResponseApplySnapshotChunk {
	return a.handler.ApplySnapshotChunk(a.checkCtx, chunk)
}
