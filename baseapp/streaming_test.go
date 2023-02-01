package baseapp

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type mockABCIListener struct {
	name      string
	ChangeSet []*storetypes.StoreKVPair
}

func NewMockABCIListener(name string) mockABCIListener {
	return mockABCIListener{
		name:      name,
		ChangeSet: make([]*storetypes.StoreKVPair, 0),
	}
}

func (m mockABCIListener) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	return nil
}

func (m mockABCIListener) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	return nil
}

func (m mockABCIListener) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	return nil
}

func (m *mockABCIListener) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	m.ChangeSet = changeSet
	return nil
}
