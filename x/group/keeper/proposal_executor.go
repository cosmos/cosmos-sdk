package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module/server"
	"github.com/cosmos/cosmos-sdk/x/group"
)

func (s serverImpl) execMsgs(ctx context.Context, derivationKey []byte, proposal group.Proposal) error {
	derivedKey := s.key.Derive(derivationKey)
	msgs := proposal.GetMsgs()

	for _, msg := range msgs {
		var reply interface{}

		// Execute the message using the derived key,
		// this will verify that the message signer is the group account.
		err := derivedKey.Invoke(ctx, server.TypeURL(msg), msg, reply)
		if err != nil {
			return err
		}
	}
	return nil
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
