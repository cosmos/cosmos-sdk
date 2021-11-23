package keeper

import (
	app "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
)

// func (s Keeper) execMsgs(ctx context.Context, derivationKey []byte, proposal group.Proposal) error {
// 	derivedKey := s.key.Derive(derivationKey)
// 	msgs := proposal.GetMsgs()

// 	for _, msg := range msgs {
// 		var reply interface{}

// 		// Execute the message using the derived key,
// 		// this will verify that the message signer is the group account.
// 		err := derivedKey.Invoke(ctx, server.TypeURL(msg), msg, reply)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// doExecuteMsgs routes the messages to the registered handlers. Messages are limited to those that require no authZ or
// by the group account only. Otherwise this gives access to other peoples accounts as the sdk ant handler is bypassed
func doExecuteMsgs(ctx sdk.Context, router app.MsgServiceRouter, proposal group.Proposal, groupAccount sdk.AccAddress) ([]sdk.Result, error) {
	msgs := proposal.GetMsgs()

	results := make([]sdk.Result, len(msgs))
	if err := ensureMsgAuthZ(msgs, groupAccount); err != nil {
		return nil, err
	}
	for i, msg := range msgs {
		handler := router.Route(ctx, msg.Route())
		if handler == nil {
			return nil, errors.Wrapf(group.ErrInvalid, "no message handler found for %q", msg.Route())
		}
		r, err := handler(ctx, msg)
		if err != nil {
			return nil, errors.Wrapf(err, "message %q at position %d", msg.Type(), i)
		}
		if r != nil {
			results[i] = *r
		}
	}
	return results, nil
}

// ensureMsgAuthZ checks that if a message requires signers that all of them are equal to the given group account.
func ensureMsgAuthZ(msgs []sdk.Msg, groupAccount sdk.AccAddress) error {
	for i := range msgs {
		for _, acct := range msgs[i].GetSigners() {
			if !groupAccount.Equals(acct) {
				return errors.Wrap(errors.ErrUnauthorized, "msg does not have group account authorization")
			}
		}
	}
	return nil
}
