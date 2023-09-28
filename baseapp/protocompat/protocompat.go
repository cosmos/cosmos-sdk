package protocompat

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var (
	gogoType    = reflect.TypeOf((*gogoproto.Message)(nil)).Elem()
	protov2Type = reflect.TypeOf((*proto2.Message)(nil)).Elem()
)

type Handler = func(ctx context.Context, message gogoproto.Message) (gogoproto.Message, error)

func MakeHybridHandler(cdc codec.BinaryCodec, sd *grpc.ServiceDesc, method grpc.MethodDesc, handler interface{}) (Handler, error) {
	methodFullName := protoreflect.FullName(fmt.Sprintf("%s.%s", sd.ServiceName, method.MethodName))
	desc, err := gogoproto.HybridResolver.FindDescriptorByName(methodFullName)
	if err != nil {
		return nil, err
	}
	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return nil, fmt.Errorf("invalid method descriptor %s", methodFullName)
	}

	funcType, err := getFuncType(handler, method.MethodName)
	if err != nil {
		return nil, err
	}
	// the handler function takes two arguments: context.Context and proto.Message
	// since it is a method, the first argument is the receiver, which we don't care about.
	// the second argument is the proto.Message, which we need to know the type of.
	inputType := funcType.In(2)
	switch {
	case inputType.Implements(protov2Type):
		return makeProtoV2HybridHandler(methodDesc, cdc, method, handler)
	case inputType.Implements(gogoType):
		return makeGogoHybridHandler(methodDesc, cdc, method, handler)
	default:
		return nil, fmt.Errorf("invalid method handler type %T, input does not implement", method.Handler)
	}
}

// makeProtoV2HybridHandler returns a handler that can handle both gogo and protov2 messages.
func makeProtoV2HybridHandler(prefMethod protoreflect.MethodDescriptor, cdc codec.BinaryCodec, method grpc.MethodDesc, handler any) (Handler, error) {
	// it's a protov2 handler, if a gogo counterparty is not found we cannot handle gogo messages.
	gogoRespType := gogoproto.MessageType(string(prefMethod.Output().FullName()))
	if gogoRespType == nil {
		return func(ctx context.Context, request gogoproto.Message) (gogoproto.Message, error) {
			protov2Request, ok := request.(proto2.Message)
			if !ok {
				return nil, fmt.Errorf("invalid request type %T, method %s does not accept gogoproto messages", request, prefMethod.FullName())
			}
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				proto2.Merge(msg.(proto2.Message), protov2Request)
				return nil
			}, nil)
			if err != nil {
				return nil, err
			}
			return resp.(gogoproto.Message), nil
		}, nil
	}
	gogoRespType = gogoRespType.Elem() // we need the non pointer type
	return func(ctx context.Context, request gogoproto.Message) (gogoproto.Message, error) {
		// we check if the request is a protov2 message.
		switch m := request.(type) {
		case proto2.Message:
			// we can just call the handler after making a copy of the message, for safety reasons.
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				proto2.Merge(msg.(proto2.Message), m)
				return nil
			}, nil)
			if err != nil {
				return nil, err
			}
			return resp.(gogoproto.Message), nil
		case gogoproto.Message:
			// we need to marshal and unmarshal the request.
			requestBytes, err := cdc.Marshal(m)
			if err != nil {
				return nil, err
			}
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				// unmarshal request into the message.
				return proto2.Unmarshal(requestBytes, msg.(proto2.Message))
			}, nil)
			if err != nil {
				return nil, err
			}
			// the response is a protov2 message, so we cannot just return it.
			// since the request came as gogoproto, we expect the response
			// to also be gogoproto.
			respBytes, err := proto2.Marshal(resp.(proto2.Message))
			if err != nil {
				return nil, err
			}

			// unmarshal response into a gogo message.
			gogoResp := reflect.New(gogoRespType).Interface().(gogoproto.Message)
			return gogoResp, cdc.Unmarshal(respBytes, gogoResp)
		default:
			panic("unreachable")
		}
	}, nil
}

func makeGogoHybridHandler(prefMethod protoreflect.MethodDescriptor, cdc codec.BinaryCodec, method grpc.MethodDesc, handler any) (Handler, error) {
	// it's a gogo handler, we check if the existing protov2 counterparty exists.
	protov2RespType, err := protoregistry.GlobalTypes.FindMessageByName(prefMethod.Output().FullName())
	if err != nil {
		// this can only be a gogo message.
		return func(ctx context.Context, request gogoproto.Message) (gogoproto.Message, error) {
			_, ok := request.(proto2.Message)
			if ok {
				return nil, fmt.Errorf("invalid request type %T, method %s does not accept protov2 messages", request, prefMethod.FullName())
			}
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				// merge!
				gogoproto.Merge(msg.(gogoproto.Message), request)
				return nil
			}, nil)
			if err != nil {
				return nil, err
			}
			return resp.(gogoproto.Message), nil
		}, nil
	}
	// this is a gogo handler, and we have a protov2 counterparty.
	return func(ctx context.Context, request gogoproto.Message) (gogoproto.Message, error) {
		switch m := request.(type) {
		case proto2.Message:
			// we need to marshal and unmarshal the request.
			requestBytes, err := proto2.Marshal(m)
			if err != nil {
				return nil, err
			}
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				// unmarshal request into the message.
				return cdc.Unmarshal(requestBytes, msg.(gogoproto.Message))
			}, nil)
			if err != nil {
				return nil, err
			}
			// the response is a gogo message, so we cannot just return it.
			// since the request came as protov2, we expect the response
			// to also be protov2.
			respBytes, err := cdc.Marshal(resp.(gogoproto.Message))
			if err != nil {
				return nil, err
			}
			// now we unmarshal back into a protov2 message.
			protov2Resp := protov2RespType.New().Interface()
			return protov2Resp.(gogoproto.Message), proto2.Unmarshal(respBytes, protov2Resp)
		case gogoproto.Message:
			// we can just call the handler after making a copy of the message, for safety reasons.
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				gogoproto.Merge(msg.(gogoproto.Message), m)
				return nil
			}, nil)
			if err != nil {
				return nil, err
			}
			return resp.(gogoproto.Message), nil
		default:
			panic("unreachable")
		}
	}, nil
}

// getFuncType returns the handler type for the given method.
// this is unfortunately hideous because we need to discover
// from the gRPC server implementer (handler) the method  that
// matches the protobuf service method. This depends on the
// codegen semantics.
func getFuncType(handler any, method string) (reflect.Type, error) {
	handlerType := reflect.TypeOf(handler)
	methodType, exists := handlerType.MethodByName(method)
	if !exists {
		return nil, fmt.Errorf("method %s not found on handler", method)
	}
	return methodType.Type, nil
}
