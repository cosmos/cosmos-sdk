package implementation

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoiface"
)

// ProtoMsg is a generic interface for protobuf messages.
type ProtoMsg[T any] interface {
	*T
	protoreflect.ProtoMessage
	protoiface.MessageV1
}

// RegisterInitHandler registers an initialisation handler for a smart account that uses protobuf.
func RegisterInitHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *InitBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	reqName := ProtoReq(new(Req)).ProtoReflect().Descriptor().FullName()

	router.handler = func(ctx context.Context, initRequest any) (initResponse any, err error) {
		concrete, ok := initRequest.(ProtoReq)
		if !ok {
			return nil, fmt.Errorf("%w: wanted %s, got %T", errInvalidMessage, reqName, initRequest)
		}
		return handler(ctx, concrete)
	}

	router.schema = HandlerSchema{
		RequestSchema:  *NewProtoMessageSchema[Req, ProtoReq](),
		ResponseSchema: *NewProtoMessageSchema[Resp, ProtoResp](),
	}
}

// RegisterExecuteHandler registers an execution handler for a smart account that uses protobuf.
func RegisterExecuteHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *ExecuteBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	reqName := ProtoReq(new(Req)).ProtoReflect().Descriptor().FullName()
	// check if not registered already
	if _, ok := router.handlers[string(reqName)]; ok {
		router.err = fmt.Errorf("handler already registered for message %s", reqName)
		return
	}

	router.handlers[string(reqName)] = func(ctx context.Context, executeRequest any) (executeResponse any, err error) {
		concrete, ok := executeRequest.(ProtoReq)
		if !ok {
			return nil, fmt.Errorf("%w: wanted %s, got %T", errInvalidMessage, reqName, executeRequest)
		}
		return handler(ctx, concrete)
	}

	router.handlersSchema[string(reqName)] = HandlerSchema{
		RequestSchema:  *NewProtoMessageSchema[Req, ProtoReq](),
		ResponseSchema: *NewProtoMessageSchema[Resp, ProtoResp](),
	}
}

// RegisterQueryHandler registers a query handler for a smart account that uses protobuf.
func RegisterQueryHandler[
	Req any, ProtoReq ProtoMsg[Req], Resp any, ProtoResp ProtoMsg[Resp],
](router *QueryBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	RegisterExecuteHandler(router.er, handler)
}

func NewProtoMessageSchema[T any, PT ProtoMsg[T]]() *MessageSchema {
	msg := PT(new(T))
	marshaler := proto.MarshalOptions{Deterministic: true}
	unmarshaler := proto.UnmarshalOptions{DiscardUnknown: true} // TODO: safe to discard unknown? or should reject?
	jsonMarshaler := protojson.MarshalOptions{
		Multiline:     true,
		Indent:        "	",
		UseProtoNames: true,
	}
	jsonUnmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	return &MessageSchema{
		Name: string(msg.ProtoReflect().Descriptor().FullName()),
		TxDecode: func(bytes []byte) (any, error) {
			obj := PT(new(T))
			err := unmarshaler.Unmarshal(bytes, obj)
			return obj, err
		},
		TxEncode: func(a any) ([]byte, error) {
			concrete, ok := a.(PT)
			if !ok {
				return nil, fmt.Errorf("%w: wanted %s, got %T", errInvalidMessage, msg.ProtoReflect().Descriptor().FullName(), a)
			}
			return marshaler.Marshal(concrete)
		},
		HumanDecode: func(bytes []byte) (any, error) {
			obj := PT(new(T))
			err := jsonUnmarshaler.Unmarshal(bytes, obj)
			return obj, err
		},
		HumanEncode: func(a any) ([]byte, error) {
			concrete, ok := a.(PT)
			if !ok {
				return nil, fmt.Errorf("%w: wanted %s, got %T", errInvalidMessage, msg.ProtoReflect().Descriptor().FullName(), a)
			}
			return jsonMarshaler.Marshal(concrete)
		},
	}
}
