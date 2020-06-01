package client

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
		TxEncoder() sdk.TxEncoder
		SignModeHandler() types.SignModeHandler
	}

	// TxBuilder defines an interface which an application-defined concrete transaction
	// type must implement. Namely, it must be able to set messages, generate
	// signatures, and provide canonical bytes to sign over. The transaction must
	// also know how to encode itself.
	TxBuilder interface {
		GetTx() types.SigTx

		SetMsgs(msgs ...sdk.Msg) error
		SetSignatures(signatures ...SignatureBuilder) error
		SetMemo(memo string)
		SetFee(amount sdk.Coins)
		SetGasLimit(limit uint64)
	}

	SignatureBuilder struct {
		PubKey crypto.PubKey
		Data   types.SignatureData
	}
)
