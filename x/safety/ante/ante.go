package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

// MaxNestedMsgDepth bounds how deep SafetyDecorator descends into msgs that
// wrap other msgs. A top-level msg is depth 0.
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

// SafetyDecorator rejects a tx when any of its msgs, or any msg nested inside
// them, is disabled. It opens every msg that exposes GetMessages or GetMsgs,
// including protected wrappers such as gov MsgSubmitProposal, so a proposal
// containing a disabled msg fails CheckTx. It also opens wrappers that
// only an exact disabled entry blocks, such as authz MsgExec, so a "*" or
// prefix trip that lets the wrapper through still rejects the msgs it
// wraps. It lets CheckTx turn away tripped txs before they enter the
// mempool or pay fees, so place it before fee deduction. The router check in
// baseapp remains the authoritative gate.
//
// The decorator also runs in ReCheckTx and, with an app-side mempool, in
// PrepareProposal, so a tx that entered the mempool before a trip is evicted
// there. Like a tx rejected in
// CheckTx, it pays no fee and keeps its sequence, and its signed bytes
// execute once the trip is reset unless the sender signs another tx at that
// sequence or the tx times out.
//
// In FinalizeBlock the decorator passes every tx on and leaves the check to
// the router, since a tx rejected before fee deduction would keep its
// sequence. Rejected at the router, the tx pays its fee and uses its sequence
// like any failed tx. A proposal that reaches FinalizeBlock after a trip is
// created, and the router blocks its disabled msg when the proposal executes.
type SafetyDecorator struct {
	breaker sdk.CircuitBreaker
}

// NewSafetyDecorator returns a SafetyDecorator backed by the given breaker,
// normally the x/safety keeper.
func NewSafetyDecorator(breaker sdk.CircuitBreaker) SafetyDecorator {
	return SafetyDecorator{breaker: breaker}
}

// AnteHandle implements sdk.AnteDecorator.
func (d SafetyDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if ctx.ExecMode() == sdk.ExecModeFinalize {
		return next(ctx, tx, simulate)
	}
	if err := d.checkMsgs(ctx, tx.GetMsgs(), 0); err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}

func (d SafetyDecorator) checkMsgs(ctx sdk.Context, msgs []sdk.Msg, depth int) error {
	if depth > MaxNestedMsgDepth {
		return types.ErrMaxDepth.Wrapf("limit %d", MaxNestedMsgDepth)
	}
	for _, msg := range msgs {
		if err := msgservice.CheckCircuitBreaker(ctx, d.breaker, sdk.MsgTypeURL(msg)); err != nil {
			return err
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
