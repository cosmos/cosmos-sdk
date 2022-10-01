package grpc_abci_v1

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// GRPCClient is an implementation of the Listener interface that talks over RPC.
type GRPCClient struct {
	client         ABCIListenerServiceClient
	blockHeight    int64
	txIdx          int64
	storeKVPairIdx int64
}

var _ baseapp.ABCIListener = (*GRPCClient)(nil)

func (m *GRPCClient) ListenBeginBlock(blockHeight int64, req []byte, res []byte) error {
	_, err := m.client.ListenBeginBlock(context.Background(), &PutRequest{
		BlockHeight: blockHeight,
		Req:         req,
		Res:         res,
	})
	return err
}

func (m *GRPCClient) ListenEndBlock(blockHeight int64, req []byte, res []byte) error {
	_, err := m.client.ListenEndBlock(context.Background(), &PutRequest{
		BlockHeight: blockHeight,
		Req:         req,
		Res:         res,
	})
	return err
}

func (m *GRPCClient) ListenDeliverTx(blockHeight int64, req []byte, res []byte) error {
	m.updateTxIdx(blockHeight)
	_, err := m.client.ListenDeliverTx(context.Background(), &PutRequest{
		BlockHeight: blockHeight,
		Req:         req,
		Res:         res,
		TxIdx:       m.txIdx,
	})
	return err
}

func (m *GRPCClient) ListenStoreKVPair(blockHeight int64, data []byte) error {
	m.updateStoreKVPairIdx(blockHeight)
	_, err := m.client.ListenStoreKVPair(context.Background(), &PutRequest{
		BlockHeight:    blockHeight,
		StoreKvPair:    data,
		StoreKvPairIdx: m.storeKVPairIdx,
	})
	return err
}

func (m *GRPCClient) updateTxIdx(currBlockHeight int64) {
	if m.blockHeight < currBlockHeight {
		m.blockHeight = currBlockHeight
		m.txIdx = 0
	} else {
		m.txIdx++
	}
}

func (m *GRPCClient) updateStoreKVPairIdx(currBlockHeight int64) {
	if m.blockHeight < currBlockHeight {
		m.blockHeight = currBlockHeight
		m.storeKVPairIdx = 0
	} else {
		m.storeKVPairIdx++
	}
}

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl baseapp.ABCIListener
}

func (m *GRPCServer) ListenBeginBlock(_ context.Context, req *PutRequest) (*Empty, error) {
	return &Empty{}, m.Impl.ListenBeginBlock(req.BlockHeight, req.Req, req.Res)
}

func (m *GRPCServer) ListenEndBlock(_ context.Context, req *PutRequest) (*Empty, error) {
	return &Empty{}, m.Impl.ListenEndBlock(req.BlockHeight, req.Req, req.Res)
}

func (m *GRPCServer) ListenDeliverTx(_ context.Context, req *PutRequest) (*Empty, error) {
	return &Empty{}, m.Impl.ListenDeliverTx(req.BlockHeight, req.Req, req.Res)
}

func (m *GRPCServer) ListenStoreKVPair(_ context.Context, req *PutRequest) (*Empty, error) {
	return &Empty{}, m.Impl.ListenStoreKVPair(req.BlockHeight, req.StoreKvPair)
}
