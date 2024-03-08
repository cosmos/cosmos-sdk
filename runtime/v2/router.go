package runtime

import (
	"context"
	"fmt"

	"cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/runtime/v2/protocompat"
	"github.com/cosmos/cosmos-sdk/codec"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"
)

var _ appmodule.MsgRouter = (*handlerRouter)(nil)

type v1OrV2Message = protoiface.MessageV1
type v1Message = interface {
	gogoproto.Message
	v1OrV2Message
}
type v2Message = interface {
	proto.Message
	v1OrV2Message
}

type hybridHandler struct {
	cdc codec.BinaryCodec

	makeReqV1  func() v1Message
	makeRespV1 func() v1Message

	makeReqV2  func() v2Message
	makeRespV2 func() v2Message

	preHandlerV1 func(ctx context.Context, msg v1Message) error
	preHandlerV2 func(ctx context.Context, msg v2Message) error

	handlerV1 func(ctx context.Context, msg, resp v1Message) error // either one of this is set, not both
	handlerV2 func(ctx context.Context, msg, resp v2Message) error

	postHandlerV1 func(ctx context.Context, msg v1Message, resp v1Message) error
	postHandlerV2 func(ctx context.Context, msg v2Message, resp v2Message) error
}

type doLazy[T any] struct {
	f      func() (T, error)
	result *T
}

func newDoLazy[T any](f func() (T, error)) doLazy[T] {
	return doLazy[T]{f: f}
}

func (d doLazy[T]) get() (T, error) {
	if d.result != nil {
		return *d.result, nil
	}
	result, err := d.f()
	if err != nil {
		var t T
		return t, err
	}
	d.result = &result
	return result, nil
}

func (h hybridHandler) handle(ctx context.Context, msg v1OrV2Message, resp v1OrV2Message) error {
	switch msg := msg.(type) {
	case v2Message:
		return handleHybridMsg[v2Message, v1Message](
			ctx, msg, resp.(v2Message),
			h.marshalV2,
			h.unmarshalV2,
			h.marshalV1,
			h.unmarshalV1,
			h.makeRespV2,
			h.makeReqV1,
			h.makeRespV1,
			h.preHandlerV2,
			h.preHandlerV1,
			h.handlerV2,
			h.handlerV1,
			h.postHandlerV2,
			h.postHandlerV1,
		)
	case v1Message:
		return handleHybridMsg[v1Message, v2Message](
			ctx, msg, resp,
			h.marshalV1,
			h.unmarshalV1,
			h.marshalV2,
			h.unmarshalV2,
			h.makeRespV1,
			h.makeReqV2,
			h.makeRespV2,
			h.preHandlerV1,
			h.preHandlerV2,
			h.handlerV1,
			h.handlerV2,
			h.postHandlerV1,
			h.postHandlerV2,
		)
	default:
		return fmt.Errorf("unsupported message type: %T", msg)
	}
}

func (h hybridHandler) marshalV1(msg v1Message) ([]byte, error) {
	return h.cdc.Marshal(msg)
}

func (h hybridHandler) unmarshalV1(bytes []byte, msg v1Message) error {
	return h.cdc.Unmarshal(bytes, msg)
}

func (h hybridHandler) marshalV2(msg v2Message) ([]byte, error) {
	return proto.MarshalOptions{Deterministic: true}.Marshal(msg)
}

func (h hybridHandler) unmarshalV2(bytes []byte, msg v2Message) error {
	return proto.UnmarshalOptions{}.Unmarshal(bytes, msg)
}

// A and B represent two different interfaces for the same protobuf message.
func handleHybridMsg[A v1OrV2Message, B v1OrV2Message](
	ctx context.Context,
	reqA A,
	respA A,
	marshalA func(A) ([]byte, error),
	unmarshalA func([]byte, A) error,
	marshalB func(B) ([]byte, error),
	unmarshalB func([]byte, B) error,
	makeRespA func() A,
	makeReqB func() B,
	makeBResp func() B,
	preHandlerA func(ctx context.Context, msg A) error,
	preHandlerB func(ctx context.Context, msg B) error,
	handlerA func(ctx context.Context, msg A, resp A) error,
	handlerB func(ctx context.Context, msg B, resp B) error,
	postHandlerA func(ctx context.Context, msg A, resp A) error,
	postHandlerB func(ctx context.Context, msg B, resp B) error,
) error {
	getReqB := newDoLazy(func() (B, error) {
		msgBytes, err := marshalA(reqA)
		if err != nil {
			var b B
			return b, err
		}
		b := makeReqB()
		err = unmarshalB(msgBytes, b)
		return b, err
	})

	// before executing run pre handlers
	if preHandlerA != nil {
		err := preHandlerA(ctx, reqA)
		if err != nil {
			return err
		}
	}

	if preHandlerB != nil {
		vBReq, err := getReqB.get()
		if err != nil {
			return err
		}
		err = preHandlerB(ctx, vBReq)
		if err != nil {
			return err
		}
	}

	// this means msg and handler match
	if handlerA == nil {
		err := handlerA(ctx, reqA, respA)
		if err != nil {
			return err
		}

		if postHandlerA != nil {
			err = postHandlerA(ctx, reqA, respA)
			if err != nil {
				return err
			}
		}

		if postHandlerB != nil {
			// get reqB
			reqB, err := getReqB.get()
			if err != nil {
				return err
			}
			// we need to marshal the resp into vb
			msgBytes, err := marshalA(respA)
			respB := makeBResp()
			err = unmarshalB(msgBytes, respB)
			if err != nil {
				return err
			}
			err = postHandlerB(ctx, reqB, respB)
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		// this means handler and msg do not match, so we need to convert req into B
		// and then send it.
		reqB, err := getReqB.get()
		if err != nil {
			return err
		}

		respB := makeBResp()

		err = handlerB(ctx, reqB, respB)
		if err != nil {
			return err
		}

		if postHandlerA != nil {
			respA := makeRespA()
			respBytes, err := marshalB(respB)
			if err != nil {
				return err
			}
			err = unmarshalA(respBytes, respA)
			if err != nil {
				return err
			}
			err = postHandlerA(ctx, reqA, respA)
		}

		if postHandlerB != nil {
			err = postHandlerB(ctx, reqB, respB)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func newRouters() (*handlerRouter, *preHandlerRouter, *postHandlerRouter, *handlerRouter) {
	return &handlerRouter{
			handlers: map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error){},
		},
		&preHandlerRouter{
			specificPreHandler: map[string][]func(ctx context.Context, msg proto.Message) error{},
		},
		&postHandlerRouter{
			specificPostHandler: map[string][]func(ctx context.Context, msg proto.Message, resp proto.Message) error{},
		},
		&handlerRouter{
			handlers: map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error){},
		}
}

type handlerRouter struct {
	err      error
	handlers map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error)
}

func (h *handlerRouter) Register(msgName string, handler appmodule.Handler) {
	if h.handlers != nil {
		return
	}
	if _, exist := h.handlers[msgName]; exist {
		h.err = fmt.Errorf("conflicting msg handlers: %s", msgName)
	}
	h.handlers[msgName] = handler
}

func (h *handlerRouter) registerLegacyGRPC(cdc codec.BinaryCodec, sd *grpc.ServiceDesc, md grpc.MethodDesc, ss any) error {
	requestName, err := protocompat.RequestFullNameFromMethodDesc(sd, md)
	if err != nil {
		return err
	}

	responseName, err := protocompat.ResponseFullNameFromMethodDesc(sd, md)
	if err != nil {
		return err
	}

	// now we create the hybrid handler
	hybridHandler, err := protocompat.MakeHybridHandler(cdc, sd, md, ss)
	if err != nil {
		return err
	}

	responseV2Type, err := protoregistry.GlobalTypes.FindMessageByName(responseName)
	if err != nil {
		return err
	}

	h.handlers[string(requestName)] = func(ctx context.Context, msg transaction.Type) (resp transaction.Type, err error) {
		resp = responseV2Type.New().Interface()
		return resp, hybridHandler(ctx, msg.(protoiface.MessageV1), resp.(protoiface.MessageV1))
	}

	return nil
}

func (h *handlerRouter) build(pre *preHandlerRouter, post *postHandlerRouter) (
	func(ctx context.Context, msg proto.Message) (proto.Message, error),
	error,
) {
	handlers := make(map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error))

	globalPreHandler := func(ctx context.Context, msg transaction.Type) error {
		for _, h := range pre.globalPreHandler {
			err := h(ctx, msg)
			if err != nil {
				return err
			}
		}
		return nil
	}

	globalPostHandler := func(ctx context.Context, msg, msgResp transaction.Type) error {
		for _, h := range post.globalPostHandler {
			err := h(ctx, msg, msgResp)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for msgType, handler := range h.handlers {
		// find pre handler
		specificPreHandlers := pre.specificPreHandler[msgType]
		// find post handler
		specificPostHandlers := post.specificPostHandler[msgType]
		// build the handler
		handlers[msgType] = buildHandler(handler, specificPreHandlers, globalPreHandler, specificPostHandlers, globalPostHandler)
	}
	// TODO: add checks for when a pre handler/post handler is registered but there is no matching handler.

	// return handler as function
	return func(ctx context.Context, msg transaction.Type) (transaction.Type, error) {
		typeName := proto.MessageName(msg)
		handler, exists := handlers[string(typeName)]
		if !exists {
			return nil, fmt.Errorf("%w: %s", appmodule.ErrNoHandler, typeName)
		}
		return handler(ctx, msg)
	}, nil
}

func buildHandler(
	handler func(ctx context.Context, msg proto.Message) (proto.Message, error),
	preHandlers []func(ctx context.Context, msg proto.Message) error, globalPreHandler func(ctx context.Context, msg proto.Message) error,
	postHandlers []func(ctx context.Context, msg proto.Message, resp proto.Message) error, globalPostHandler func(ctx context.Context, msg proto.Message, resp proto.Message) error) func(ctx context.Context, msg proto.Message) (proto.Message, error) {
	return func(ctx context.Context, msg proto.Message) (msgResp proto.Message, err error) {
		for _, preHandler := range preHandlers {
			if err := preHandler(ctx, msg); err != nil {
				return nil, err
			}
		}

		err = globalPreHandler(ctx, msg)
		if err != nil {
			return nil, err
		}

		msgResp, err = handler(ctx, msg)
		if err != nil {
			return nil, err
		}

		for _, postHandler := range postHandlers {
			if err := postHandler(ctx, msg, msgResp); err != nil {
				return nil, err
			}
		}

		err = globalPostHandler(ctx, msg, msgResp)
		return msgResp, err
	}
}

var _ appmodule.PreMsgRouter = (*preHandlerRouter)(nil)

type preHandlerRouter struct {
	err error

	specificPreHandler map[string][]func(ctx context.Context, msg proto.Message) error
	globalPreHandler   []func(ctx context.Context, msg proto.Message) error
}

func (p *preHandlerRouter) Register(msgName string, handler appmodule.PreMsgHandler) {
	p.specificPreHandler[msgName] = append(p.specificPreHandler[msgName], handler)
}

func (p *preHandlerRouter) RegisterGlobal(handler appmodule.PreMsgHandler) {
	p.globalPreHandler = append(p.globalPreHandler, handler)
}

var _ appmodule.PostMsgRouter = (*postHandlerRouter)(nil)

type postHandlerRouter struct {
	err error

	specificPostHandler map[string][]func(ctx context.Context, msg proto.Message, resp proto.Message) error
	globalPostHandler   []func(ctx context.Context, msg proto.Message, resp proto.Message) error
}

func (p postHandlerRouter) Register(msgName string, handler appmodule.PostMsgHandler) {
	p.specificPostHandler[msgName] = append(p.specificPostHandler[msgName], handler)
}

func (p postHandlerRouter) RegisterGlobal(handler appmodule.PostMsgHandler) {
	p.globalPostHandler = append(p.globalPostHandler, handler)
}
