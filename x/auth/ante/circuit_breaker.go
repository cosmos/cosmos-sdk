package ante

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MaxNestedMsgDepth bounds how deep CircuitBreakerDecorator descends into msgs
// that wrap other msgs. A top-level msg is depth 0.
const MaxNestedMsgDepth = 6

// msgsWithGetMessages is implemented by wrappers such as authz MsgExec.
type msgsWithGetMessages interface {
	GetMessages() ([]sdk.Msg, error)
}

// msgsWithGetMsgs is implemented by wrappers such as gov v1 and group
// MsgSubmitProposal.
type msgsWithGetMsgs interface {
	GetMsgs() ([]sdk.Msg, error)
}

// CircuitBreakerDecorator rejects a tx when the circuit breaker disallows any
// of its msgs, or any msg nested inside them. It opens every msg that exposes
// GetMessages or GetMsgs, including wrappers the breaker always allows, such
// as gov MsgSubmitProposal under x/safety, so a proposal containing a
// disabled msg fails CheckTx. It lets CheckTx turn away tripped txs before
// they enter the mempool or pay fees, so it runs before fee deduction. The
// router check in baseapp remains the authoritative gate.
//
// The decorator also runs in ReCheckTx and, with an app-side mempool, in
// PrepareProposal, so a tx that entered the mempool before a trip is evicted
// there. Like a tx rejected in CheckTx, it pays no fee and keeps its
// sequence, and its signed bytes execute once the trip is reset unless the
// sender signs another tx at that sequence or the tx times out. With an
// app-side mempool it runs in ProcessProposal too, where a tripped msg makes
// the SDK's default handler reject the whole proposal.
//
// In FinalizeBlock the decorator passes every tx on and leaves the check to
// the router, since a tx rejected before fee deduction would keep its
// sequence. Rejected at the router, the tx pays its fee and uses its sequence
// like any failed tx.
//
// A breaker error rejects the tx with that error, so x/safety's refusal keeps
// its codespace and names the blocking pattern. A refusal without an error
// returns sdkerrors.ErrCircuitBreakerTripped, as the router does.
type CircuitBreakerDecorator struct {
	breaker baseapp.CircuitBreaker
}

// NewCircuitBreakerDecorator returns a CircuitBreakerDecorator backed by the
// given breaker, normally the one the app passes to SetCircuitBreaker.
func NewCircuitBreakerDecorator(breaker baseapp.CircuitBreaker) CircuitBreakerDecorator {
	return CircuitBreakerDecorator{breaker: breaker}
}

// AnteHandle implements sdk.AnteDecorator.
func (d CircuitBreakerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if ctx.ExecMode() == sdk.ExecModeFinalize {
		return next(ctx, tx, simulate)
	}
	if err := d.checkMsgs(ctx, tx.GetMsgs(), 0); err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}

func (d CircuitBreakerDecorator) checkMsgs(ctx sdk.Context, msgs []sdk.Msg, depth int) error {
	if depth > MaxNestedMsgDepth {
		return sdkerrors.ErrInvalidRequest.Wrapf("nested msg depth exceeds limit %d", MaxNestedMsgDepth)
	}
	for _, msg := range msgs {
		typeURL := sdk.MsgTypeURL(msg)
		allowed, err := d.breaker.IsAllowed(ctx, typeURL)
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
		if err := d.checkMsgs(ctx, inner, depth+1); err != nil {
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
