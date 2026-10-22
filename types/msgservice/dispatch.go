package msgservice

import (
	"reflect"

	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgRouter returns the handler registered for a msg, or nil if there is
// none. baseapp.MsgServiceRouter implements it.
type MsgRouter interface {
	Handler(msg sdk.Msg) func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error)
}

// Dispatch runs msg through router's handler for it and returns the msg
// server's response as R. msg must be the gogoproto type registered for its
// type URL, not the pulsar type from cosmossdk.io/api. The router handler
// runs ValidateBasic, the app's circuit breaker check and the msg server
// method on a branch of ctx.
// Dispatch writes the branch and emits the handler's events on ctx only when
// the handler succeeds and its response is an R, so a failed Dispatch leaves
// ctx unchanged. Dispatch emits only the msg server's events: the "message"
// event with action, sender and module, and the msg_index attribute, come
// from baseapp's tx path and are absent, as they are for msgs authz runs.
//
// Code outside a module, such as an EVM precompile or another module's
// keeper, runs the module's msgs with Dispatch. The SDK modules do not
// export their msg servers, so the router, with its circuit breaker check,
// is the one way to run a msg. Dispatch does not check the msg's signers:
// the caller authorizes them.
func Dispatch[R proto.Message](ctx sdk.Context, router MsgRouter, msg sdk.Msg) (R, error) {
	var resp R
	if router == nil {
		return resp, sdkerrors.ErrLogic.Wrap("nil msg router")
	}
	handler := router.Handler(msg)
	if handler == nil {
		return resp, sdkerrors.ErrUnknownRequest.Wrapf("no msg service handler for %s", sdk.MsgTypeURL(msg))
	}
	// A pulsar msg has the type URL of its gogo type, but the handler for
	// that URL takes only the gogo type.
	if want := proto.MessageType(proto.MessageName(msg)); reflect.TypeOf(msg) != want {
		return resp, sdkerrors.ErrInvalidType.Wrapf("expected %v for %s, got %T", want, sdk.MsgTypeURL(msg), msg)
	}

	cacheCtx, write := ctx.CacheContext()
	res, err := handler(cacheCtx, msg)
	if err != nil {
		return resp, err
	}

	if len(res.MsgResponses) != 1 {
		return resp, sdkerrors.ErrInvalidType.Wrapf("expected one response for %s, got %d", sdk.MsgTypeURL(msg), len(res.MsgResponses))
	}
	resp, ok := res.MsgResponses[0].GetCachedValue().(R)
	if !ok {
		return resp, sdkerrors.ErrInvalidType.Wrapf("expected %T response for %s, got %T", resp, sdk.MsgTypeURL(msg), res.MsgResponses[0].GetCachedValue())
	}

	events := make(sdk.Events, len(res.Events))
	for i, e := range res.Events {
		events[i] = sdk.Event(e)
	}
	cacheCtx.EventManager().EmitEvents(events)
	write()
	return resp, nil
}
