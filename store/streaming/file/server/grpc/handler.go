package grpc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/streaming/file/server/config"
	pb "github.com/cosmos/cosmos-sdk/store/streaming/file/server/v1beta"
	"github.com/tendermint/tendermint/libs/log"
)

// Handler wraps the StateFileServer interface with an additional Stop() method
type Handler interface {
	pb.StateFileServer
	Stop()
}

// handler is the struct which implements the Handler methods
type handler struct {
	pb.UnimplementedStateFileServer
	backend  *StateFileBackend
	logger   log.Logger
	quitChan chan struct{}
}

// NewHandler returns the object for the gRPC handler
func NewHandler(conf *config.StateFileServerConfig, codec *codec.ProtoCodec, logger log.Logger) (Handler, error) {
	quitChan := make(chan struct{})
	return &handler{
		backend:  NewStateFileBackend(conf, codec, logger, quitChan),
		logger:   logger,
		quitChan: quitChan,
	}, nil
}

// StreamData implements StateFileServer
// StreamData streams the requested state file data
// this streams new data as it is written to disk
func (h *handler) StreamData(req *pb.StreamRequest, srv pb.StateFile_StreamDataServer) error {
	resChan := make(chan *pb.StreamResponse)
	err, done := h.backend.StreamData(req, resChan)
	if err != nil {
		return err
	}
	for {
		select {
		case res := <-resChan:
			if err := srv.Send(res); err != nil {
				h.logger.Error("StreamData send error", "err", err)
			}
		// if Close() is called the backend process will quit and close this channel
		// so we don't need a select case for h.quitChan here
		// this way we wait for the backend to finish sending before shutting down the handler
		case <-done:
			h.logger.Info("quiting handler StreamData process")
			return nil
		}
	}
}

// BackFillData implements StateFileServer
// BackFillData streams the requested state file data
// this streams data that is already written to disk
func (h *handler) BackFillData(req *pb.StreamRequest, srv pb.StateFile_BackFillDataServer) error {
	resChan := make(chan *pb.StreamResponse)
	err, done := h.backend.BackFillData(req, resChan)
	if err != nil {
		return err
	}
	for {
		select {
		case res := <-resChan:
			if err := srv.Send(res); err != nil {
				h.logger.Error("BackFillData send error", "err", err)
			}
		// if Close() is called the backend process will quit and close this channel
		// so we don't need a select case for h.quitChan here
		// this way we wait for the backend to finish sending before shutting down the handler
		case <-done:
			h.logger.Info("quiting handler BackFillData process")
			return nil
		}
	}
}

// BeginBlockDataAt implements StateFileServer
// BeginBlockDataAt returns a BeginBlockPayload for the provided BeginBlockRequest
func (h *handler) BeginBlockDataAt(ctx context.Context, req *pb.BeginBlockRequest) (*pb.BeginBlockPayload, error) {
	return h.backend.BeginBlockDataAt(ctx, req)
}

// DeliverTxDataAt implements StateFileServer
// DeliverTxDataAt returns a DeliverTxPayload for the provided BeginBlockRequest
func (h *handler) DeliverTxDataAt(ctx context.Context, req *pb.DeliverTxRequest) (*pb.DeliverTxPayload, error) {
	return h.backend.DeliverTxDataAt(ctx, req)
}

// EndBlockDataAt implements StateFileServer
// EndBlockDataAt returns a EndBlockPayload for the provided EndBlockRequest
func (h *handler) EndBlockDataAt(ctx context.Context, req *pb.EndBlockRequest) (*pb.EndBlockPayload, error) {
	return h.backend.EndBlockDataAt(ctx, req)
}

// Stop stops the handler processes and its backend processes
func (h *handler) Stop() {
	close(h.quitChan)
}
