package grpc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/streaming/file/server/config"
	pb "github.com/cosmos/cosmos-sdk/store/streaming/file/server/v1beta"
	"github.com/tendermint/tendermint/libs/log"
)

type Handler = pb.StateFileServer

// handler is the interface which exposes the StateFile Server methods
type handler struct {
	pb.UnimplementedStateFileServer
	backend *StateFileBackend
	logger  log.Logger
}

// New returns the object for the RPC handler
func New(conf config.StateServerConfig, codec *codec.ProtoCodec, logger log.Logger) (Handler, error) {
	return &handler{
		backend: NewStateFileBackend(conf, codec, logger),
		logger:  logger,
	}, nil
}

// StreamData streams the requested state file data
// this streams new data as it is written to disk
func (h *handler) StreamData(req *pb.StreamRequest, srv pb.StateFile_StreamDataServer) error {
	resChan := make(chan *pb.StreamResponse)
	stopped := make(chan struct{})
	if err := h.backend.StreamData(req, resChan, stopped); err != nil {
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
	if err := h.backend.BackFillData(req, resChan, stopped); err != nil {
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

// BeginBlockDataAt returns a BeginBlockPayload for the provided BeginBlockRequest
func (h *handler) BeginBlockDataAt(ctx context.Context, req *pb.BeginBlockRequest) (*pb.BeginBlockPayload, error) {
	return h.backend.BeginBlockDataAt(ctx, req)
}

// DeliverTxDataAt returns a DeliverTxPayload for the provided BeginBlockRequest
func (h *handler) DeliverTxDataAt(ctx context.Context, req *pb.DeliverTxRequest) (*pb.DeliverTxPayload, error) {
	return h.backend.DeliverTxDataAt(ctx, req)
}

// EndBlockDataAt returns a EndBlockPayload for the provided EndBlockRequest
func (h *handler) EndBlockDataAt(ctx context.Context, req *pb.EndBlockRequest) (*pb.EndBlockPayload, error) {
	return h.backend.EndBlockDataAt(ctx, req)
}
