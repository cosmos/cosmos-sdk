package baseapp

import (
	"context"

	"github.com/tendermint/tendermint/abci/types"
)

type baseApp struct {
	ctx context.Context
}

var _ types.Application = &baseApp{}

func (a baseApp) Info(info types.RequestInfo) types.ResponseInfo {
	panic("implement me")
}

func (a baseApp) SetOption(option types.RequestSetOption) types.ResponseSetOption {
	return types.ResponseSetOption{}
}

func (a baseApp) Query(query types.RequestQuery) types.ResponseQuery {
	panic("implement me")
}

func (a baseApp) CheckTx(tx types.RequestCheckTx) types.ResponseCheckTx {
	panic("implement me")
}

func (a baseApp) InitChain(chain types.RequestInitChain) types.ResponseInitChain {
	panic("implement me")
}

func (a baseApp) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	panic("TODO")
}

func (a baseApp) DeliverTx(tx types.RequestDeliverTx) types.ResponseDeliverTx {
	panic("implement me")
}

func (a baseApp) EndBlock(block types.RequestEndBlock) types.ResponseEndBlock {
	panic("implement me")
}

func (a baseApp) Commit() types.ResponseCommit {
	panic("implement me")
}

func (a baseApp) ListSnapshots(snapshots types.RequestListSnapshots) types.ResponseListSnapshots {
	panic("implement me")
}

func (a baseApp) OfferSnapshot(snapshot types.RequestOfferSnapshot) types.ResponseOfferSnapshot {
	panic("implement me")
}

func (a baseApp) LoadSnapshotChunk(chunk types.RequestLoadSnapshotChunk) types.ResponseLoadSnapshotChunk {
	panic("implement me")
}

func (a baseApp) ApplySnapshotChunk(chunk types.RequestApplySnapshotChunk) types.ResponseApplySnapshotChunk {
	panic("implement me")
}
