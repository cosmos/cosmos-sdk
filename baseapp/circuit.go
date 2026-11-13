package baseapp

import (
	"context"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CircuitBreaker is an interface that defines the methods for a circuit breaker.
type CircuitBreaker interface {
	IsAllowed(ctx context.Context, typeURL string) (bool, error)
}

// msgsWithGetMessages is implemented by wrappers such as authz MsgExec.
type msgsWithGetMessages interface {
	GetMessages() ([]sdk.Msg, error)
}

// msgsWithGetMsgs is implemented by wrappers such as gov v1 and group
// MsgSubmitProposal.
type msgsWithGetMsgs interface {
	GetMsgs() ([]sdk.Msg, error)
}

// CheckMsgsAllowed returns an error when cb disallows any of msgs, or any msg
// nested inside them. It opens every msg that exposes GetMessages or GetMsgs,
// including wrappers the breaker always allows, such as gov MsgSubmitProposal
// under x/safety.
//
// A breaker error is returned as is, so x/safety's refusal keeps its
// codespace and names the blocking pattern. A refusal without an error
// returns sdkerrors.ErrCircuitBreakerTripped, as the router does.
//
// It descends as deep as the tx decoder unpacks Any values,
// codectypes.MaxUnpackAnyRecursionDepth, and adds no nesting rule of its own.
// Msgs built in process that nest deeper fail with sdkerrors.ErrInvalidRequest
// rather than pass unchecked.
func CheckMsgsAllowed(ctx context.Context, cb CircuitBreaker, msgs []sdk.Msg) error {
	return checkMsgsAllowed(ctx, cb, msgs, 0)
}

func checkMsgsAllowed(ctx context.Context, cb CircuitBreaker, msgs []sdk.Msg, depth int) error {
	if depth > codectypes.MaxUnpackAnyRecursionDepth {
		return sdkerrors.ErrInvalidRequest.Wrapf("nested msg depth exceeds limit %d", codectypes.MaxUnpackAnyRecursionDepth)
	}
	for _, msg := range msgs {
		typeURL := sdk.MsgTypeURL(msg)
		allowed, err := cb.IsAllowed(ctx, typeURL)
		if err != nil {
			return err
		}
		if !allowed {
			return sdkerrors.ErrCircuitBreakerTripped.Wrap(typeURL)
		}

		inner, err := innerMsgs(msg)
		if err != nil {
			return err
		}
		if len(inner) == 0 {
			continue
		}
		if err := checkMsgsAllowed(ctx, cb, inner, depth+1); err != nil {
			return err
		}
	}
	return nil
}

// innerMsgs returns the msgs a wrapper exposes through either GetMessages or
// GetMsgs, or nil when msg wraps nothing.
func innerMsgs(msg sdk.Msg) ([]sdk.Msg, error) {
	var inner []sdk.Msg
	if w, ok := msg.(msgsWithGetMessages); ok {
		msgs, err := w.GetMessages()
		if err != nil {
			return nil, err
		}
		inner = append(inner, msgs...)
	}
	if w, ok := msg.(msgsWithGetMsgs); ok {
		msgs, err := w.GetMsgs()
		if err != nil {
			return nil, err
		}
		inner = append(inner, msgs...)
	}
	return inner, nil
}
