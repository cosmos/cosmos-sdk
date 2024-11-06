package grpc

import (
	"context"

	"github.com/cosmos/gogoproto/proto"
	gogoproto "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
)

// v2Service implements the gRPC service interface for handling queries and listing handlers.
type v2Service struct {
	queryHandlers map[string]appmodulev2.Handler
	queryable     func(ctx context.Context, version uint64, msg transaction.Msg) (transaction.Msg, error)
}

// Query handles incoming query requests by unmarshaling the request, processing it,
// and returning the response in an Any protobuf message.
func (s v2Service) Query(ctx context.Context, request *QueryRequest) (*QueryResponse, error) {
	if request == nil || request.Request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	msgName := request.Request.TypeUrl

	handler, exists := s.queryHandlers[msgName]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "handler not found for %s", msgName)
	}

	protoMsg := handler.MakeMsg()
	if err := proto.Unmarshal(request.Request.Value, protoMsg); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal request: %v", err)
	}

	queryResp, err := s.queryable(ctx, 0, protoMsg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	respBytes, err := proto.Marshal(queryResp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal response: %v", err)
	}

	anyResp := &gogoproto.Any{
		TypeUrl: "/" + proto.MessageName(queryResp),
		Value:   respBytes,
	}

	return &QueryResponse{Response: anyResp}, nil
}

func (s v2Service) ListQueryHandlers(_ context.Context, _ *ListQueryHandlersRequest) (*ListQueryHandlersResponse, error) {
	var handlerDescriptors []*Handler
	for handlerName := range s.queryHandlers {
		msg := s.queryHandlers[handlerName].MakeMsg()
		resp := s.queryHandlers[handlerName].MakeMsgResp()

		handlerDescriptors = append(handlerDescriptors, &Handler{
			RequestName:  proto.MessageName(msg),
			ResponseName: proto.MessageName(resp),
		})
	}

	return &ListQueryHandlersResponse{Handlers: handlerDescriptors}, nil
}
