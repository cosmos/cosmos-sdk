package protocompat

import (
	"context"
	"fmt"
	"reflect"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoiface"
)

var (
	gogoType    = reflect.TypeOf((*gogoproto.Message)(nil)).Elem()
	protov2Type = reflect.TypeOf((*proto2.Message)(nil)).Elem()
)

type Handler = func(ctx context.Context, request, response protoiface.MessageV1) error

// MakeHandler returns a handler that can handle both gogo and protov2 messages, no matter
// if the handler is a gogo or protov2 handler.
func MakeHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler interface{}) (Handler, error) {
	methodFullName := protoreflect.FullName(fmt.Sprintf("%s.%s", sd.ServiceName, method.MethodName))
	desc, err := gogoproto.HybridResolver.FindDescriptorByName(methodFullName)
	if err != nil {
		return nil, err
	}
	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return nil, fmt.Errorf("invalid method descriptor %s", methodFullName)
	}

	isProtov2Handler, err := isProtov2(method)
	if err != nil {
		return nil, err
	}
	if isProtov2Handler {
		return nil, fmt.Errorf("protov2 handlers are not allowed %s", methodFullName)
	}
	return makeGogoHandler(methodDesc, method, handler)
}

func makeGogoHandler(prefMethod protoreflect.MethodDescriptor, method grpc.MethodDesc, handler any) (Handler, error) {
	return func(ctx context.Context, inReq, outResp protoiface.MessageV1) error {
		// we do not handle protov2
		_, ok := inReq.(proto2.Message)
		if ok {
			return fmt.Errorf("invalid request type %T, method %s does not accept protov2 messages", inReq, prefMethod.FullName())
		}

		resp, err := method.Handler(handler, ctx, func(msg any) error {
			// reflection to copy from inReq to msg
			dstVal := reflect.ValueOf(msg).Elem()
			srcVal := reflect.ValueOf(inReq).Elem()
			dstVal.Set(srcVal)
			return nil
		}, nil)
		if err != nil {
			return err
		}

		// reflection to copy from resp to outResp
		dstVal := reflect.ValueOf(outResp).Elem()
		srcVal := reflect.ValueOf(resp).Elem()
		dstVal.Set(srcVal)

		return nil
	}, nil
}

// isProtov2 returns true if the given method accepts protov2 messages.
// Returns false if it does not.
// It uses the decoder function passed to the method handler to determine
// the type. Since the decoder function is passed in by the concrete implementer the expected
// message where bytes are unmarshaled to, we can use that to determine the type.
func isProtov2(md grpc.MethodDesc) (isV2Type bool, err error) {
	pullRequestType := func(msg interface{}) error {
		typ := reflect.TypeOf(msg)
		switch {
		case typ.Implements(protov2Type):
			isV2Type = true
			return nil
		case typ.Implements(gogoType):
			isV2Type = false
			return nil
		default:
			err = fmt.Errorf("invalid request type %T, expected protov2 or gogo message", msg)
			return nil
		}
	}
	// doNotExecute is a dummy handler that stops the request execution.
	doNotExecute := func(_ context.Context, _ any, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (any, error) {
		return nil, nil
	}
	// we are allowed to pass in a nil context and nil request, since we are not actually executing the request.
	// this is made possible by the doNotExecute function which immediately returns without calling other handlers.
	_, _ = md.Handler(nil, nil, pullRequestType, doNotExecute)
	return
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
