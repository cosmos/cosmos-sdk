package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	authmiddleware "github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/group"
	grouperrors "github.com/cosmos/cosmos-sdk/x/group/errors"
)

// doExecuteMsgs routes the messages to the registered handlers. Messages are limited to those that require no authZ or
// by the account of group policy only. Otherwise this gives access to other peoples accounts as the sdk ant handler is bypassed
func (s Keeper) doExecuteMsgs(ctx sdk.Context, router *authmiddleware.MsgServiceRouter, proposal group.Proposal, groupPolicyAcc sdk.AccAddress) ([]sdk.Result, error) {
	// Ensure it's not too late to execute the messages.
	// After https://github.com/cosmos/cosmos-sdk/issues/11245, proposals should
	// be pruned automatically, so this function should not even be called, as
	// the proposal doesn't exist in state. For sanity check, we can still keep
	// this simple and cheap check.
	expiryDate := proposal.VotingPeriodEnd.Add(s.config.MaxExecutionPeriod)
	if expiryDate.Before(ctx.BlockTime()) {
		return nil, grouperrors.ErrExpired.Wrapf("proposal expired on %s", expiryDate)
	}

	msgs := proposal.GetMsgs()

	results := make([]sdk.Result, len(msgs))
	if err := ensureMsgAuthZ(msgs, groupPolicyAcc); err != nil {
		return nil, err
	}
	for i, msg := range msgs {
		handler := s.router.Handler(msg)
		if handler == nil {
			return nil, errors.Wrapf(grouperrors.ErrInvalid, "no message handler found for %q", sdk.MsgTypeURL(msg))
		}
		r, err := handler(ctx, msg)
		if err != nil {
			return nil, errors.Wrapf(err, "message %q at position %d", msg, i)
		}
		if r != nil {
			results[i] = *r
		}
	}
	return results, nil
}

// ensureMsgAuthZ checks that if a message requires signers that all of them are equal to the given account address of group policy.
func ensureMsgAuthZ(msgs []sdk.Msg, groupPolicyAcc sdk.AccAddress) error {
	for i := range msgs {
		for _, acct := range msgs[i].GetSigners() {
			if !groupPolicyAcc.Equals(acct) {
				return errors.Wrap(errors.ErrUnauthorized, "msg does not have group policy authorization")
			}
		}
	}
	return nil
}
