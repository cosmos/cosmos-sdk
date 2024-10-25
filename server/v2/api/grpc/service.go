package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"github.com/cosmos/gogoproto/proto"
	gogoproto "github.com/cosmos/gogoproto/types/any"
)

// RegisterV2Service registers the V2 gRPC service implementation with the given server.
// It takes a generic type T that implements the transaction.Tx interface.
func RegisterV2Service[T transaction.Tx](server *grpc.Server, app serverv2.AppI[T]) {
	if server == nil || app == nil {
		panic("server and app must not be nil")
	}
	RegisterServiceServer(server, NewV2Service(app))
}

// V2Service implements the gRPC service interface for handling queries and listing handlers.
type V2Service[T transaction.Tx] struct {
	app serverv2.AppI[T]
}

// NewV2Service creates a new instance of V2Service with the provided app.
func NewV2Service[T transaction.Tx](app serverv2.AppI[T]) *V2Service[T] {
	return &V2Service[T]{app: app}
}

// Query handles incoming query requests by unmarshaling the request, processing it,
// and returning the response in an Any protobuf message.
func (s V2Service[T]) Query(ctx context.Context, request *QueryRequest) (*QueryResponse, error) {
	if request == nil || request.Request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	msgName := request.Request.TypeUrl
	handlers := s.app.QueryHandlers()

	handler, exists := handlers[msgName]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "handler not found for %s", msgName)
	}

	protoMsg := handler.MakeMsg()
	if err := proto.Unmarshal(request.Request.Value, protoMsg); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal request: %v", err)
	}

	queryResp, err := s.app.Query(ctx, 0, protoMsg)
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

func (s V2Service[T]) ListQueryHandlers(_ context.Context, _ *ListQueryHandlersRequest) (*ListQueryHandlersResponse, error) {
	handlersMap := s.app.QueryHandlers()

	var handlerDescriptors []*Handler
	for handlerName := range handlersMap {
		msg := handlersMap[handlerName].MakeMsg()
		resp := handlersMap[handlerName].MakeMsgResp()

		handlerDescriptors = append(handlerDescriptors, &Handler{
			RequestName:  proto.MessageName(msg),
			ResponseName: proto.MessageName(resp),
		})
	}

	return &ListQueryHandlersResponse{Handlers: handlerDescriptors}, nil
}
