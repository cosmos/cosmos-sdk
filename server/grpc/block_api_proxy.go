package grpc

import (
	"context"
	"fmt"
	"io"

	coregrpc "github.com/cometbft/cometbft/rpc/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ coregrpc.BlockAPIServer = (*blockAPIProxy)(nil)

type blockAPIProxy struct {
	client coregrpc.BlockAPIClient
}

// NewBlockAPIProxy creates a new core block api proxy server using the provided protocol and address string for the upstream client
// e.g. tcp://0.0.0.0:9099
func NewBlockAPIProxy(protoAddr string) (*blockAPIProxy, error) {
	blockAPIClient, err := coregrpc.StartBlockAPIGRPCClient(protoAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &blockAPIProxy{
		client: blockAPIClient,
	}, nil
}

// BlockByHash implements coregrpc.BlockAPIServer.
func (b *blockAPIProxy) BlockByHash(req *coregrpc.BlockByHashRequest, srv coregrpc.BlockAPI_BlockByHashServer) error {
	proxySrv, err := b.client.BlockByHash(srv.Context(), req)
	if err != nil {
		return err
	}

	isLast := false
	for !isLast {
		resp, err := proxySrv.Recv()
		if err != nil {
			return err
		}

		if err := srv.Send(resp); err != nil {
			return err
		}

		isLast = resp.IsLast
	}

	return nil
}

// BlockByHeight implements coregrpc.BlockAPIServer.
func (b *blockAPIProxy) BlockByHeight(req *coregrpc.BlockByHeightRequest, srv coregrpc.BlockAPI_BlockByHeightServer) error {
	proxySrv, err := b.client.BlockByHeight(srv.Context(), req)
	if err != nil {
		return err
	}

	isLast := false
	for !isLast {
		resp, err := proxySrv.Recv()
		if err != nil {
			return err
		}

		if err := srv.Send(resp); err != nil {
			return err
		}

		isLast = resp.IsLast
	}

	return nil
}

// Commit implements coregrpc.BlockAPIServer.
func (b *blockAPIProxy) Commit(ctx context.Context, req *coregrpc.CommitRequest) (*coregrpc.CommitResponse, error) {
	return b.client.Commit(ctx, req)
}

// Status implements coregrpc.BlockAPIServer.
func (b *blockAPIProxy) Status(ctx context.Context, req *coregrpc.StatusRequest) (*coregrpc.StatusResponse, error) {
	return b.client.Status(ctx, req)
}

// SubscribeNewHeights implements coregrpc.BlockAPIServer.
func (b *blockAPIProxy) SubscribeNewHeights(req *coregrpc.SubscribeNewHeightsRequest, srv coregrpc.BlockAPI_SubscribeNewHeightsServer) error {
	proxySrv, err := b.client.SubscribeNewHeights(srv.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to upstream: %w", err)
	}

	for {
		resp, err := proxySrv.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("upstream recv error: %w", err)
		}

		if err := srv.Send(resp); err != nil {
			return fmt.Errorf("downstream send error: %w", err)
		}
	}
}

// ValidatorSet implements coregrpc.BlockAPIServer.
func (b *blockAPIProxy) ValidatorSet(ctx context.Context, req *coregrpc.ValidatorSetRequest) (*coregrpc.ValidatorSetResponse, error) {
	return b.client.ValidatorSet(ctx, req)
}
