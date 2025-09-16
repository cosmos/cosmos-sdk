package protocompat

import (
	"context"
	"fmt"
	"reflect"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/golang/protobuf/proto" // nolint: staticcheck // needed because gogoproto.Merge does not work consistently. See NOTE: comments.
	"google.golang.org/grpc"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	// gogoType represents the reflection type for gogoproto.Message interface
	gogoType = reflect.TypeOf((*gogoproto.Message)(nil)).Elem()
	// protov2Type represents the reflection type for google.golang.org/protobuf/proto.Message interface
	protov2Type = reflect.TypeOf((*proto2.Message)(nil)).Elem()
	// protov2MarshalOpts contains marshal options for protov2 messages with deterministic output
	protov2MarshalOpts = proto2.MarshalOptions{Deterministic: true}
)

// Handler defines a function type that can handle both gogoproto and protov2 messages.
// It takes a context, request message, and response message, returning an error if any.
type Handler = func(ctx context.Context, request, response protoiface.MessageV1) error

// MakeHybridHandler creates a handler that can process both gogoproto and protov2 messages.
// It automatically detects the handler type and creates appropriate compatibility layer.
// The returned handler can accept either message type and convert between them as needed.
func MakeHybridHandler(cdc codec.BinaryCodec, sd *grpc.ServiceDesc, method grpc.MethodDesc, handler any) (Handler, error) {
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
		return makeProtoV2HybridHandler(methodDesc, cdc, method, handler)
	}
	return makeGogoHybridHandler(methodDesc, cdc, method, handler)
}

// makeProtoV2HybridHandler creates a handler for protov2-based methods that can also handle gogoproto messages.
// It performs automatic conversion between message types when necessary.
func makeProtoV2HybridHandler(prefMethod protoreflect.MethodDescriptor, cdc codec.BinaryCodec, method grpc.MethodDesc, handler any) (Handler, error) {
	// Check if a gogoproto counterpart exists for this protov2 method.
	// If not, we can only handle protov2 messages directly.
	gogoExists := gogoproto.MessageType(string(prefMethod.Output().FullName())) != nil
	if !gogoExists {
		// Handler for protov2-only methods (no gogoproto counterpart available)
		return func(ctx context.Context, inReq, outResp protoiface.MessageV1) error {
			protov2Request, ok := inReq.(proto2.Message)
			if !ok {
				return fmt.Errorf("invalid request type %T, method %s does not accept gogoproto messages", inReq, prefMethod.FullName())
			}
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				// Copy the protov2 request into the handler's message parameter
				proto2.Merge(msg.(proto2.Message), protov2Request)
				return nil
			}, nil)
			if err != nil {
				return err
			}
			// Copy the protov2 response back to the output message
			proto2.Merge(outResp.(proto2.Message), resp.(proto2.Message))
			return nil
		}, nil
	}
	// Handler for protov2 methods that have gogoproto counterparts
	return func(ctx context.Context, inReq, outResp protoiface.MessageV1) error {
		// Handle both protov2 and gogoproto message types
		switch m := inReq.(type) {
		case proto2.Message:
			// Direct protov2 message - no conversion needed
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				// Copy the protov2 request into the handler's message parameter
				proto2.Merge(msg.(proto2.Message), m)
				return nil
			}, nil)
			if err != nil {
				return err
			}
			// Copy the protov2 response back to the output message
			proto2.Merge(outResp.(proto2.Message), resp.(proto2.Message))
			return nil
		case gogoproto.Message:
			// Convert gogoproto message to protov2 for the handler
			requestBytes, err := cdc.Marshal(m)
			if err != nil {
				return err
			}
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				// Unmarshal gogoproto bytes into protov2 message for the handler
				return proto2.Unmarshal(requestBytes, msg.(proto2.Message))
			}, nil)
			if err != nil {
				return err
			}
			// Convert protov2 response back to gogoproto format
			// since the original request was gogoproto, the response should match
			respBytes, err := protov2MarshalOpts.Marshal(resp.(proto2.Message))
			if err != nil {
				return err
			}

			// Unmarshal protov2 response bytes into gogoproto message
			return cdc.Unmarshal(respBytes, outResp.(gogoproto.Message))
		default:
			panic("unreachable")
		}
	}, nil
}

// makeGogoHybridHandler creates a handler for gogoproto-based methods that can also handle protov2 messages.
// It performs automatic conversion between message types when necessary.
func makeGogoHybridHandler(prefMethod protoreflect.MethodDescriptor, cdc codec.BinaryCodec, method grpc.MethodDesc, handler any) (Handler, error) {
	// Check if a protov2 counterpart exists for this gogoproto method.
	_, err := protoregistry.GlobalTypes.FindMessageByName(prefMethod.Output().FullName())
	if err != nil {
		// No protov2 counterpart exists - this method only handles gogoproto messages
		return func(ctx context.Context, inReq, outResp protoiface.MessageV1) error {
			// Reject protov2 messages since no conversion is possible
			_, ok := inReq.(proto2.Message)
			if ok {
				return fmt.Errorf("invalid request type %T, method %s does not accept protov2 messages", inReq, prefMethod.FullName())
			}
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				// Copy the gogoproto request into the handler's message parameter
				// ref: https://github.com/cosmos/cosmos-sdk/issues/18003
				// NOTE: gogoproto.Merge fails for unknown reasons, but proto.Merge works correctly
				// with gogoproto messages, so we use the standard proto.Merge instead.
				proto.Merge(msg.(gogoproto.Message), inReq)
				return nil
			}, nil)
			if err != nil {
				return err
			}
			// Copy the gogoproto response back to the output message
			// ref: https://github.com/cosmos/cosmos-sdk/issues/18003
			// NOTE: gogoproto.Merge fails for unknown reasons, but proto.Merge works correctly
			// with gogoproto messages, so we use the standard proto.Merge instead.
			proto.Merge(outResp.(gogoproto.Message), resp.(gogoproto.Message))
			return nil
		}, nil
	}
	// Handler for gogoproto methods that have protov2 counterparts
	return func(ctx context.Context, inReq, outResp protoiface.MessageV1) error {
		switch m := inReq.(type) {
		case proto2.Message:
			// Convert protov2 message to gogoproto for the handler
			requestBytes, err := protov2MarshalOpts.Marshal(m)
			if err != nil {
				return err
			}
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				// Unmarshal protov2 bytes into gogoproto message for the handler
				return cdc.Unmarshal(requestBytes, msg.(gogoproto.Message))
			}, nil)
			if err != nil {
				return err
			}
			// Convert gogoproto response back to protov2 format
			// since the original request was protov2, the response should match
			respBytes, err := cdc.Marshal(resp.(gogoproto.Message))
			if err != nil {
				return err
			}
			// Unmarshal gogoproto response bytes into protov2 message
			return proto2.Unmarshal(respBytes, outResp.(proto2.Message))
		case gogoproto.Message:
			// Direct gogoproto message - no conversion needed
			resp, err := method.Handler(handler, ctx, func(msg any) error {
				// Copy the gogoproto request into the handler's message parameter
				// ref: https://github.com/cosmos/cosmos-sdk/issues/18003
				asGogoProto := msg.(gogoproto.Message)
				// NOTE: gogoproto.Merge fails for unknown reasons, but proto.Merge works correctly
				// with gogoproto messages, so we use the standard proto.Merge instead.
				proto.Merge(asGogoProto, m)
				return nil
			}, nil)
			if err != nil {
				return err
			}
			// Copy the gogoproto response back to the output message
			// ref: https://github.com/cosmos/cosmos-sdk/issues/18003
			// NOTE: gogoproto.Merge fails for unknown reasons, but proto.Merge works correctly
			// with gogoproto messages, so we use the standard proto.Merge instead.
			proto.Merge(outResp.(gogoproto.Message), resp.(gogoproto.Message))
			return nil
		default:
			panic("unreachable")
		}
	}, nil
}

// isProtov2 determines whether the given method handler expects protov2 messages.
// It returns true if the handler accepts protov2 messages, false if it expects gogoproto messages.
// The function uses the decoder function passed to the method handler to determine the expected
// message type. Since the decoder function is provided by the concrete implementer and specifies
// the message type where bytes should be unmarshaled, we can use reflection to determine the type.
func isProtov2(md grpc.MethodDesc) (isV2Type bool, err error) {
	pullRequestType := func(msg any) error {
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
	// doNotExecute is a dummy handler that immediately returns without processing the request.
	// This allows us to inspect the message type without actually executing the handler logic.
	doNotExecute := func(_ context.Context, _ any, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (any, error) {
		return nil, nil
	}
	// We can safely pass nil context and nil request since we are not actually executing the request.
	// The doNotExecute function immediately returns without calling any other handlers or processing logic.
	_, _ = md.Handler(nil, nil, pullRequestType, doNotExecute)
	return
}

// RequestFullNameFromMethodDesc returns the fully-qualified name of the request message type
// for the specified gRPC service method. This is useful for identifying message types
// in the protobuf registry.
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
