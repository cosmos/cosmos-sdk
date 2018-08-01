package utils

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
)

// SendTx implements a auxiliary handler that facilitates sending a series of
// messages in a signed transaction given a TxContext and a QueryContext. It
// ensures that the account exists, has a proper number and sequence set. In
// addition, it builds and signs a transaction with the supplied messages.
// Finally, it broadcasts the signed transaction to a node.
func SendTx(txCtx authctx.TxContext, queryCtx context.QueryContext, msgs []sdk.Msg) error {
	if err := queryCtx.EnsureAccountExists(); err != nil {
		return err
	}

	from, err := queryCtx.GetFromAddress()
	if err != nil {
		return err
	}

	if txCtx.AccountNumber == 0 {
		accNum, err := queryCtx.GetAccountNumber(from)
		if err != nil {
			return err
		}

		txCtx = txCtx.WithAccountNumber(accNum)
	}

	if txCtx.Sequence == 0 {
		accSeq, err := queryCtx.GetAccountSequence(from)
		if err != nil {
			return err
		}

		txCtx = txCtx.WithSequence(accSeq)
	}

	passphrase, err := keys.GetPassphrase(queryCtx.FromAddressName)
	if err != nil {
		return err
	}

	// build and sign the transaction
	txBytes, err := txCtx.BuildAndSign(queryCtx.FromAddressName, passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	return queryCtx.EnsureBroadcastTx(txBytes)
}
