package context

import (
	"github.com/tendermint/tendermint/crypto"

	types "github.com/cosmos/cosmos-sdk/types/tx"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// TxGenerator defines an interface a client can utilize to generate an
	// application-defined concrete transaction type. The type returned must
	// implement TxBuilder.
	TxGenerator interface {
		NewTxBuilder() TxBuilder
		NewFee() ClientFee
		NewSignature() ClientSignature
		TxEncoder() sdk.TxEncoder
		SignModeHandler() types.SignModeHandler
	}

	ClientFee interface {
		sdk.Fee
		SetGas(uint64)
		SetAmount(sdk.Coins)
	}

	ClientSignature interface {
		sdk.Signature
		SetPubKey(crypto.PubKey) error
		SetSignature([]byte)
	}

	// TxBuilder defines an interface which an application-defined concrete transaction
	// type must implement. Namely, it must be able to set messages, generate
	// signatures, and provide canonical bytes to sign over. The transaction must
	// also know how to encode itself.
	TxBuilder interface {
		GetTx() sdk.Tx

		SetMsgs(...sdk.Msg) error
		GetSignatures() []sdk.Signature
		SetSignatures(...ClientSignature) error
		GetFee() sdk.Fee
		SetFee(ClientFee) error
		GetMemo() string
		SetMemo(string)

		// CanonicalSignBytes returns the canonical sign bytes to sign over, given a
		// chain ID, along with an account and sequence number.
		CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error)
	}
)
