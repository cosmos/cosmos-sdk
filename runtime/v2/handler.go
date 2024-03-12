package runtime

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
)

type (
	v1OrV2Message = protoiface.MessageV1

	v1Message = interface {
		gogoproto.Message
		v1OrV2Message
	}

	v2Message = interface {
		proto.Message
		v1OrV2Message
	}

	v1Handler = func(ctx context.Context, msg v1Message, resp v1Message) error
	v2Handler = func(ctx context.Context, msg v2Message, resp v2Message) error

	v1PreHandler = func(ctx context.Context, msg v1Message) error
	v2PreHandler = func(ctx context.Context, msg v2Message) error

	v1PostHandler = func(ctx context.Context, msg v1Message, resp v1Message) error
	v2PostHandler = func(ctx context.Context, msg v2Message, resp v2Message) error
)

func newHybridHandler(
	cdc codec.BinaryCodec,
	makeReqV1 func() v1Message,
	makeRespV1 func() v1Message,
	makeReqV2 func() v2Message,
	makeRespV2 func() v2Message,
	preHandlerV1 v1PreHandler,
	preHandlerV2 v2PreHandler,
	handlerV1 v1Handler,
	handlerV2 v2Handler,
	postHandlerV1 v1PostHandler,
	postHandlerV2 v2PostHandler,
) *hybridHandler {
	return &hybridHandler{
		cdc:           cdc,
		makeReqV1:     makeReqV1,
		makeRespV1:    makeRespV1,
		makeReqV2:     makeReqV2,
		makeRespV2:    makeRespV2,
		preHandlerV1:  preHandlerV1,
		preHandlerV2:  preHandlerV2,
		handlerV1:     handlerV1,
		handlerV2:     handlerV2,
		postHandlerV1: postHandlerV1,
		postHandlerV2: postHandlerV2,
	}
}

type hybridHandler struct {
	cdc codec.BinaryCodec

	makeReqV1  func() v1Message
	makeRespV1 func() v1Message

	makeReqV2  func() v2Message
	makeRespV2 func() v2Message

	preHandlerV1 v1PreHandler
	preHandlerV2 v2PreHandler

	handlerV1 v1Handler // either one of this is set, not both
	handlerV2 v2Handler

	postHandlerV1 v1PostHandler
	postHandlerV2 v2PostHandler
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
