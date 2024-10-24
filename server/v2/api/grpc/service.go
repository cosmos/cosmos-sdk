package grpc

import (
	"context"
	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"github.com/cosmos/gogoproto/proto"
	gogoproto "github.com/cosmos/gogoproto/types/any"
)

type MyServer struct {
	app serverv2.AppI[transaction.Tx]
}

func (m MyServer) Query(ctx context.Context, request *QueryRequest) (*QueryResponse, error) {
	msgName := request.Request.TypeUrl

	handlers := m.app.QueryHandlers()

	protoMsg := handlers[msgName].MakeMsg()
	err := proto.Unmarshal(request.Request.Value, protoMsg)
	if err != nil {
		return nil, err
	}

	queryResp, err := m.app.Query(ctx, 1, protoMsg)
	if err != nil {
		return nil, err
	}

	respBytes, err := proto.Marshal(queryResp)
	if err != nil {
		return nil, err
	}

	anyResp := &gogoproto.Any{
		TypeUrl: "/" + proto.MessageName(queryResp),
		Value:   respBytes,
	}

	return &QueryResponse{Response: anyResp}, nil
}

func (m MyServer) ListQueryHandlers(ctx context.Context, request *ListQueryHandlersRequest) (*ListQueryHandlersResponse, error) {
	//TODO implement me
	panic("implement me")
}
