package accounts

import (
	"context"
	"fmt"
	"reflect"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type MsgRouter interface {
	Handler(msg sdk.Msg) func(ctx sdk.Context, req sdk.Msg) (*sdk.Result, error)
}

func gogoToProto(cdc codec.Codec, msg proto.Message) (gogoproto.Message, error) {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	reflectType := gogoproto.MessageType(string(msg.ProtoReflect().Descriptor().FullName()))
	if reflectType == nil {
		return nil, fmt.Errorf("no proto type registered for %s", msg.ProtoReflect().Descriptor().FullName())
	}
	gogoMsg := reflect.New(reflectType).Interface().(gogoproto.Message)
	err = cdc.Unmarshal(msgBytes, gogoMsg)
	if err != nil {
		return nil, err
	}
	return gogoMsg, nil
}

func newExecFunc(cdc codec.Codec, router MsgRouter) func(ctx context.Context, msg proto.Message) (proto.Message, error) {
	// we need to convert messages back and forth.
	return func(ctx context.Context, msg proto.Message) (proto.Message, error) {
		gogoMsg, err := gogoToProto(cdc, msg)
		if err != nil {
			return nil, err
		}

		handler := router.Handler(gogoMsg)
		if handler == nil {
			return nil, fmt.Errorf("no handler for message %s", msg.ProtoReflect().Descriptor().FullName())
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		res, err := handler(sdkCtx, gogoMsg)
		if err != nil {
			return nil, err
		}

		if len(res.MsgResponses) != 1 {
			return nil, fmt.Errorf("unexpected number of responses: %d", len(res.MsgResponses))
		}

		// unpack any
		return anypb.UnmarshalNew(&anypb.Any{
			TypeUrl: res.MsgResponses[0].TypeUrl,
			Value:   res.MsgResponses[0].Value,
		}, proto.UnmarshalOptions{})
	}
}

type QueryRouter interface {
	Route(path string) func(ctx sdk.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error)
}

// newQueryFunc returns a function that can be used to query the chain.
// This accepts a concrete protobuf message, and it returns bytes. This is because
// we need the concrete type to marshal it to bytes as an ABCI query, but we don't
// have knowledge of the concrete returned type.
func newQueryFunc(qr QueryRouter) func(ctx context.Context, method string, msg proto.Message) ([]byte, error) {
	return func(ctx context.Context, method string, msg proto.Message) ([]byte, error) {
		handler := qr.Route(method)
		if handler == nil {
			return nil, fmt.Errorf("no handler for method %s", method)
		}

		requestBytes, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}

		queryReq := abci.RequestQuery{
			Data: requestBytes,
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		queryResp, err := handler(sdkCtx, &queryReq)
		if err != nil {
			return nil, err
		}
		return queryResp.Value, nil
	}
}
