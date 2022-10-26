package grpc_abci_v1

import (
	"context"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"io"
)

var (
	_ baseapp.ABCIListener = (*GRPCClient)(nil)
	_ ABCIListenerPlugin   = (*GRPCClient)(nil)
)

// GRPCClient is an implementation of the ABCIListener and ABCIListenerPlugin interfaces that talks over RPC.
type GRPCClient struct {
	client        ABCIListenerServiceClient
	stream        ABCIListenerService_StreamClient
	stopNodeOnErr bool
}

func (m *GRPCClient) Listen(ctx context.Context, blockHeight int64, eventType string, data []byte) error {
	_, err := m.client.Listen(ctx, &ListenRequest{
		BlockHeight: blockHeight,
		EventType:   eventType,
		Data:        data,
	})
	return err
}

func (m *GRPCClient) ListenBeginBlock(ctx types.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	reqbz, err := req.Marshal()
	if err != nil {
		return err
	}
	resbz, err := res.Marshal()
	if err != nil {
		return err
	}
	if err := m.Listen(ctx, ctx.BlockHeight(), "BEGIN_BLOCK_REQ", reqbz); err != nil {
		return err
	}
	if err := m.Listen(ctx, ctx.BlockHeight(), "BEGIN_BLOCK_RES", resbz); err != nil {
		return err
	}
	return nil
}

func (m *GRPCClient) ListenEndBlock(ctx types.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	reqbz, err := req.Marshal()
	if err != nil {
		return err
	}
	resbz, err := res.Marshal()
	if err != nil {
		return err
	}
	if err := m.Listen(ctx, ctx.BlockHeight(), "END_BLOCK_REQ", reqbz); err != nil {
		return err
	}
	if err := m.Listen(ctx, ctx.BlockHeight(), "END_BLOCK_RES", resbz); err != nil {
		return err
	}
	return nil
}

func (m *GRPCClient) ListenDeliverTx(ctx types.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	reqbz, err := req.Marshal()
	if err != nil {
		return err
	}
	resbz, err := res.Marshal()
	if err != nil {
		return err
	}
	if err := m.Listen(ctx, ctx.BlockHeight(), "DELIVER_TX_REQ", reqbz); err != nil {
		return err
	}
	if err := m.Listen(ctx, ctx.BlockHeight(), "DELIVER_TX_RES", resbz); err != nil {
		return err
	}
	return nil
}

func (m *GRPCClient) OnStoreCommit(ctx types.Context, changeSet [][]byte) error {
	stream, err := m.client.Stream(ctx)
	if err != nil {
		return err
	}
	for _, data := range changeSet {
		if err = stream.Send(&ListenRequest{
			BlockHeight: ctx.BlockHeight(),
			EventType:   "STATE_CHANGE",
			Data:        data,
		}); err != nil {
			return err
		}
	}
	if _, err := stream.CloseAndRecv(); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl ABCIListenerPlugin
}

func (m *GRPCServer) Listen(ctx context.Context, req *ListenRequest) (*Empty, error) {
	return &Empty{}, m.Impl.Listen(ctx, req.BlockHeight, req.EventType, req.Data)
}

func (m *GRPCServer) Stream(stream ABCIListenerService_StreamServer) error {
	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err = m.Impl.Listen(context.Background(), recv.BlockHeight, recv.EventType, recv.Data); err != nil {
			return err
		}
	}
	return nil
}
