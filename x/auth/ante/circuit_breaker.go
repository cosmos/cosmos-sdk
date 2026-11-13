package ante

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
// like any failed tx. Gov and group run the same check when a proposal is
// submitted, through MsgServiceRouter.CheckMsgsAllowed, so a proposer cannot
// store a proposal holding a disabled msg by skipping CheckTx.
//
// The decorator runs baseapp.CheckMsgsAllowed, which sets the error a refused
// tx returns and how deep the check descends into nested msgs.
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
	if err := baseapp.CheckMsgsAllowed(ctx, d.breaker, tx.GetMsgs()); err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}
