package tx

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// Generator defines an interface a client can utilize to generate an
	// application-defined concrete transaction type. The type returned must
	// implement ClientTx.
	Generator interface {
		NewTx() ClientTx
	}

	// ClientTx defines an interface which an application-defined concrete transaction
	// type must implement. Namely, it must be able to set messages, generate
	// signatures, and provide canonical bytes to sign over. The transaction must
	// also know how to encode itself.
	ClientTx interface {
		sdk.Tx
		codec.ProtoMarshaler

		SetMsgs(...sdk.Msg) error
		GetSignatures() []sdk.Signature
		SetSignatures(...sdk.Signature)
		GetFee() sdk.Fee
		SetFee(sdk.Fee)
		GetMemo() string
		SetMemo(string)

		// CanonicalSignBytes returns the canonical JSON bytes to sign over, given a
		// chain ID, along with an account and sequence number. The JSON encoding
		// ensures all field names adhere to their proto definition, default values
		// are omitted, and follows the JSON Canonical Form.
		CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error)
	}
)
