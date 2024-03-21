package protocompat

import (
	"context"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/golang/protobuf/proto" // nolint: staticcheck // needed because gogoproto.Merge does not work consistently. See NOTE: comments.
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/core/transaction"
)

type Handler = func(ctx context.Context, request, response transaction.Type) error

// MakeHandler returns a handler given a message.
func MakeHandler(method grpc.MethodDesc, handler interface{}) (Handler, error) {
	return makeGogoHybridHandler(method, handler)
}

func makeGogoHybridHandler(method grpc.MethodDesc, handler any) (Handler, error) {
	return func(ctx context.Context, inReq, outResp transaction.Type) error {
		resp, err := method.Handler(handler, ctx, func(msg any) error {
			// merge! ref: https://github.com/cosmos/cosmos-sdk/issues/18003
			// NOTE: using gogoproto.Merge will fail for some reason unknown to me, but
			// using proto.Merge with gogo messages seems to work fine.
			proto.Merge(msg.(gogoproto.Message), inReq)
			return nil
		}, nil)
		if err != nil {
			return err
		}
		// merge resp, ref: https://github.com/cosmos/cosmos-sdk/issues/18003
		// NOTE: using gogoproto.Merge will fail for some reason unknown to me, but
		// using proto.Merge with gogo messages seems to work fine.
		proto.Merge(outResp.(gogoproto.Message), resp.(gogoproto.Message))
		return nil
	}, nil
}

// RequestFullNameFromMethodDesc returns the fully-qualified name of the request message of the provided service's method.
func RequestFullNameFromMethodDesc(sd *grpc.ServiceDesc, method grpc.MethodDesc) (protoreflect.FullName, error) {
	methodFullName := protoreflect.FullName(fmt.Sprintf("%s.%s", sd.ServiceName, method.MethodName))
	desc, err := gogoproto.HybridResolver.FindDescriptorByName(methodFullName)
	if err != nil {
		return "", fmt.Errorf("cannot find method descriptor %s", methodFullName)
	}
	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return "", fmt.Errorf("invalid method descriptor %s", methodFullName)
	}
	return methodDesc.Input().FullName(), nil
}

// ResponseFullNameFromMethodDesc returns the fully-qualified name of the response message of the provided service's method.
func ResponseFullNameFromMethodDesc(sd *grpc.ServiceDesc, method grpc.MethodDesc) (protoreflect.FullName, error) {
	methodFullName := protoreflect.FullName(fmt.Sprintf("%s.%s", sd.ServiceName, method.MethodName))
	desc, err := gogoproto.HybridResolver.FindDescriptorByName(methodFullName)
	if err != nil {
		return "", fmt.Errorf("cannot find method descriptor %s", methodFullName)
	}
	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return "", fmt.Errorf("invalid method descriptor %s", methodFullName)
	}
	return methodDesc.Output().FullName(), nil
}
