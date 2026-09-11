package node

import (
	"context"
	"errors"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/server/blocklog"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// RegisterLogsService registers the LogsService directly on a gRPC server.
// It is deliberately not routed through the GRPCQueryRouter: the query reads
// no application state and must never be reachable over ABCI Query, where a
// slow scan would block FinalizeBlock. A nil store registers a service that
// reports capture as disabled.
func RegisterLogsService(server gogogrpc.Server, store *blocklog.Store) {
	RegisterLogsServiceServer(server, NewLogsServer(store))
}

// NewLogsServer returns a LogsServiceServer backed by store, which may be nil
// when capture is disabled.
func NewLogsServer(store *blocklog.Store) LogsServiceServer {
	return logsServer{store: store}
}

type logsServer struct{ store *blocklog.Store }

var _ LogsServiceServer = logsServer{}

func (s logsServer) Logs(_ context.Context, req *LogsRequest) (*LogsResponse, error) {
	if s.store == nil {
		return nil, status.Error(codes.FailedPrecondition, "block log capture is disabled: set [block-logs] retain-blocks > 0 in app.toml")
	}
	blReq := blocklog.Request{Start: req.Start, End: req.End, Level: req.Level, Module: req.Module}
	if p := req.Pagination; p != nil {
		switch {
		case p.Offset != 0:
			return nil, status.Error(codes.InvalidArgument, "pagination.offset is not supported, use key")
		case p.CountTotal:
			return nil, status.Error(codes.InvalidArgument, "pagination.count_total is not supported")
		case p.Reverse:
			return nil, status.Error(codes.InvalidArgument, "pagination.reverse is not supported")
		}
		blReq.Key, blReq.Limit = p.Key, p.Limit
	}

	res, err := s.store.Query(blReq)
	if errors.Is(err, blocklog.ErrInvalidArgument) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	out := &LogsResponse{Entries: make([]*LogEntry, len(res.Entries)), Pagination: &query.PageResponse{NextKey: res.NextKey}}
	for i, e := range res.Entries {
		out.Entries[i] = &LogEntry{Height: e.Height, Time: e.Time, Level: e.Level, Module: e.Module, Msg: e.Msg, Fields: e.Fields}
	}
	return out, nil
}
