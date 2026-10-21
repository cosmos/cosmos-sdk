package msgservice

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CheckMsgAllowed returns an error when the circuit breaker attached to ctx
// blocks msg. Msg server methods call it first, so a msg that reaches its msg
// server without the MsgServiceRouter, such as a direct call from another
// module or an EVM precompile, gets the same check as a routed msg.
//
// It returns nil when ctx has no circuit breaker or has the circuit breaker
// bypass set, and skips the lookup when the MsgServiceRouter is dispatching
// this msg type in ctx, since the router has already checked it.
func CheckMsgAllowed(ctx context.Context, msg sdk.Msg) error {
	sdkCtx, ok := ctx.Value(sdk.SdkContextKey).(sdk.Context)
	if !ok {
		return nil
	}
	cb := sdkCtx.CircuitBreaker()
	if cb == nil {
		return nil
	}
	typeURL := sdk.MsgTypeURL(msg)
	if sdkCtx.CheckedMsgTypeURL() == typeURL {
		return nil
	}
	return CheckCircuitBreaker(sdkCtx, cb, typeURL)
}

// CheckCircuitBreaker returns an error when cb blocks typeURL. An error from
// cb is returned as is, so the msg fails closed. A plain refusal returns
// sdkerrors.ErrUnauthorized. It returns nil without asking cb when ctx is, or
// wraps, an sdk.Context with the circuit breaker bypass set.
func CheckCircuitBreaker(ctx context.Context, cb sdk.CircuitBreaker, typeURL string) error {
	if sdkCtx, ok := ctx.Value(sdk.SdkContextKey).(sdk.Context); ok && sdkCtx.CircuitBreakerBypass() {
		return nil
	}
	allowed, err := cb.IsAllowed(ctx, typeURL)
	if err != nil {
		return err
	}
	if !allowed {
		return sdkerrors.ErrUnauthorized.Wrapf("circuit breaker disables execution of this message: %s", typeURL)
	}
	return nil
}
