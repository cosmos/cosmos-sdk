package grpc

import (
	"context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/state_file_server/config"
	pb "github.com/cosmos/cosmos-sdk/state_file_server/grpc/v1beta"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler = pb.StateFileServer

// handler is the interface which exposes the StateFile Server methods
type handler struct {
	pb.UnimplementedStateFileServer
	backend *StateFileBackend
	logger log.Logger
}

// New returns the object for the RPC handler
func New(conf config.StateServerBackendConfig, codec *codec.ProtoCodec, logger log.Logger) (Handler, error) {
	return &handler{
		backend: NewStateFileBackend(conf, codec, logger),
		logger: logger,
	}, nil
}

// StreamData streams the requested state file data
// this streams new data as it is written to disk
func (h *handler) StreamData(req *pb.StreamRequest, srv pb.StateFile_StreamDataServer) error {
	resChan := make(chan *pb.StreamResponse)
	stopped := make(chan struct{})
	if err := h.backend.Stream(req, resChan, stopped); err != nil {
		return err
	}
	for {
		select {
		case res := <-resChan:
			if err := srv.Send(res); err != nil {
				h.logger.Error("StreamData send error", "err", err)
			}
		case <-stopped:
			return nil
		}
	}
}

// BackFillData stream the requested state file data
// this stream data that is already written to disk
func (h *handler) BackFillData(req *pb.StreamRequest, srv pb.StateFile_BackFillDataServer) error {
	resChan := make(chan *pb.StreamResponse)
	stopped := make(chan struct{})
	if err := h.backend.BackFill(req, resChan, stopped); err != nil {
		return err
	}
	for {
		select {
		case res := <-resChan:
			if err := srv.Send(res); err != nil {
				h.logger.Error("BackFillData send error", "err", err)
			}
		case <-stopped:
			return nil
		}
	}
}

func (h *handler) BeginBlockDataAt(ctx context.Context, req *pb.BeginBlockRequest) (*pb.BeginBlockPayload, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BeginBlockDataAt not implemented")
}
func (h *handler) DeliverTxDataAt(ctx context.Context, req *pb.DeliverTxRequest) (*pb.DeliverTxPayload, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeliverTxDataAt not implemented")
}
func (h *handler) EndBlockDataAt(ctx context.Context, req *pb.EndBlockRequest) (*pb.EndBlockPayload, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EndBlockDataAt not implemented")
}