package grpc

import (
	"context"

	"github.com/cosmos/gogoproto/proto"
	gogoproto "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/appmanager"
)

// v2Service implements the gRPC service interface for handling queries and listing handlers.
type v2Service[T transaction.Tx] struct {
	queryHandlers map[string]appmodulev2.Handler
	queryable     queryFunc
	stf           appmanager.StateTransitionFunction[T]
	store         serverv2.Store
	gaslimit      uint64
}

// Query handles incoming query requests by unmarshaling the request, processing it,
// and returning the response in an Any protobuf message.
func (s v2Service[T]) Query(ctx context.Context, request *QueryRequest) (*QueryResponse, error) {
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

	height, err := getHeightFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid get height from context: %v", err)
	}

	var state corestore.ReaderMap
	if height == 0 {
		_, state, err = s.store.StateLatest()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get latest state: %v", err)
		}
	} else {
		state, err = s.store.StateAt(height)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get state at height %d: %v", height, err)
		}
	}

	queryResp, err := s.queryable.Query(ctx, state, s.gaslimit, protoMsg)
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

func (s v2Service[T]) ListQueryHandlers(_ context.Context, _ *ListQueryHandlersRequest) (*ListQueryHandlersResponse, error) {
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
