package abci

import (
	"context"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci/proto"
)

// GRPCClient is an implementation of the Listener interface that talks over RPC.
type GRPCClient struct {
	client proto.ABCIListenerServiceClient
}

var _ baseapp.ABCIListener = (*GRPCClient)(nil)

func (m *GRPCClient) ListenBeginBlock(blockHeight int64, req []byte, res []byte) error {
	_, err := m.client.ListenBeginBlock(context.Background(), &proto.PutRequest{
		BlockHeight: blockHeight,
		Req:         req,
		Res:         res,
	})
	return err
}

func (m *GRPCClient) ListenEndBlock(blockHeight int64, req []byte, res []byte) error {
	_, err := m.client.ListenEndBlock(context.Background(), &proto.PutRequest{
		BlockHeight: blockHeight,
		Req:         req,
		Res:         res,
	})
	return err
}

func (m *GRPCClient) ListenDeliverTx(blockHeight int64, req []byte, res []byte) error {
	_, err := m.client.ListenDeliverTx(context.Background(), &proto.PutRequest{
		BlockHeight: blockHeight,
		Req:         req,
		Res:         res,
	})
	return err
}

func (m *GRPCClient) ListenStoreKVPair(blockHeight int64, data []byte) error {
	_, err := m.client.ListenStoreKVPair(context.Background(), &proto.PutRequest{
		BlockHeight: blockHeight,
		StoreKvPair: data,
	})
	return err
}

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl baseapp.ABCIListener
}

func (m *GRPCServer) ListenBeginBlock(_ context.Context, req *proto.PutRequest) (*proto.Empty, error) {
	return &proto.Empty{}, m.Impl.ListenBeginBlock(req.BlockHeight, req.Req, req.Res)
}

func (m *GRPCServer) ListenEndBlock(_ context.Context, req *proto.PutRequest) (*proto.Empty, error) {
	return &proto.Empty{}, m.Impl.ListenEndBlock(req.BlockHeight, req.Req, req.Res)
}

func (m *GRPCServer) ListenDeliverTx(_ context.Context, req *proto.PutRequest) (*proto.Empty, error) {
	return &proto.Empty{}, m.Impl.ListenDeliverTx(req.BlockHeight, req.Req, req.Res)
}

func (m *GRPCServer) ListenStoreKVPair(_ context.Context, req *proto.PutRequest) (*proto.Empty, error) {
	return &proto.Empty{}, m.Impl.ListenStoreKVPair(req.BlockHeight, req.StoreKvPair)
}
