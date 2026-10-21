package msgservice

import (
	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgRouter returns the handler registered for a msg, or nil if there is
// none. baseapp.MsgServiceRouter implements it.
type MsgRouter interface {
	Handler(msg sdk.Msg) func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error)
}

// Dispatch runs msg through the handler router registered for it and returns
// the msg server's response as R. The router handler runs ValidateBasic, the
// app's circuit breaker check and the msg server method, and Dispatch emits
// the handler's events on ctx.
//
// Code outside a module that would call the module's msg server directly,
// such as an EVM precompile or another module's keeper, calls Dispatch
// instead so a circuit breaker trip of the msg stops it. Dispatch does not
// check the msg's signers: the caller authorizes them, as it would before a
// direct call.
func Dispatch[R proto.Message](ctx sdk.Context, router MsgRouter, msg sdk.Msg) (R, error) {
	var resp R
	handler := router.Handler(msg)
	if handler == nil {
		return resp, sdkerrors.ErrUnknownRequest.Wrapf("no msg service handler for %s", sdk.MsgTypeURL(msg))
	}

	res, err := handler(ctx, msg)
	if err != nil {
		return resp, err
	}

	events := make(sdk.Events, len(res.Events))
	for i, e := range res.Events {
		events[i] = sdk.Event(e)
	}
	ctx.EventManager().EmitEvents(events)

	if len(res.MsgResponses) != 1 {
		return resp, sdkerrors.ErrInvalidType.Wrapf("expected one response for %s, got %d", sdk.MsgTypeURL(msg), len(res.MsgResponses))
	}
	resp, ok := res.MsgResponses[0].GetCachedValue().(R)
	if !ok {
		return resp, sdkerrors.ErrInvalidType.Wrapf("expected %T response for %s, got %T", resp, sdk.MsgTypeURL(msg), res.MsgResponses[0].GetCachedValue())
	}
	return resp, nil
}
