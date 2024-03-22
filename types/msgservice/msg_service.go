package msgservice

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"

	"cosmossdk.io/core/registry"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterMsgServiceDesc registers all type_urls from Msg services described
// in `sd` into the registry.
func RegisterMsgServiceDesc(registrar registry.InterfaceRegistrar, sd *grpc.ServiceDesc) {
	for _, method := range sd.Methods {
		// we call each handler with methods that introspect the request type and
		// register it as a concrete implementation of sdk.Msg
		_, _ = method.Handler(nil, context.Background(), func(req any) error {
			msg, ok := req.(proto.Message)
			if !ok {
				panic(fmt.Errorf("expected proto.Message, got %T", req))
			}
			registrar.RegisterImplementations((*sdk.Msg)(nil), msg)
			return nil
		}, func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			// this is needed to avoid the nil handler from being called
			return nil, nil
		})
	}
}
