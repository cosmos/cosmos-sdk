package msgservice

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterMsgServiceDesc registers all type_urls from Msg services described
// in `sd` into the registry.
func RegisterMsgServiceDesc(registry codectypes.InterfaceRegistry, sd *grpc.ServiceDesc) {
	// Adds a top-level type_url based on the Msg service name.
	for _, method := range sd.Methods {
		fqMethod := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
		methodHandler := method.Handler

		// NOTE: This is how we pull the concrete request type for each handler for registering in the InterfaceRegistry.
		// This approach is maybe a bit hacky, but less hacky than reflecting on the handler object itself.
		// We use a no-op interceptor to avoid actually calling into the handler itself.
		_, _ = methodHandler(nil, context.Background(), func(i interface{}) error {
			msg, ok := i.(proto.Message)
			if !ok {
				// We panic here because there is no other alternative and the app cannot be initialized correctly
				// this should only happen if there is a problem with code generation in which case the app won't
				// work correctly anyway.
				panic(fmt.Errorf("can't register request type %T for service method %s", i, fqMethod))
			}

			registry.RegisterCustomTypeURL((*sdk.MsgRequest)(nil), fqMethod, msg)
			return nil
		}, noopInterceptor)

	}
}

// gRPC NOOP interceptor
func noopInterceptor(_ context.Context, _ interface{}, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (interface{}, error) {
	return nil, nil
}
