package errors

/**
*    Copyright (C) 2017 Ethan Frey
**/

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
)

const (
	msgDecoding          = "Error decoding input"
	msgUnauthorized      = "Unauthorized"
	msgInvalidAddress    = "Invalid Address"
	msgInvalidCoins      = "Invalid Coins"
	msgInvalidFormat     = "Invalid Format"
	msgInvalidSequence   = "Invalid Sequence"
	msgInvalidSignature  = "Invalid Signature"
	msgInsufficientFees  = "Insufficient Fees"
	msgNoInputs          = "No Input Coins"
	msgNoOutputs         = "No Output Coins"
	msgTooLarge          = "Input size too large"
	msgMissingSignature  = "Signature missing"
	msgTooManySignatures = "Too many signatures"
	msgNoChain           = "No chain id provided"
	msgWrongChain        = "Tx belongs to different chain - %s"
	msgUnknownTxType     = "We cannot handle this tx - %v"
)

func UnknownTxType(tx basecoin.Tx) TMError {
	msg := fmt.Sprintf(msgUnknownTxType, tx)
	return New(msg, abci.CodeType_UnknownRequest)
}

func InternalError(msg string) TMError {
	return New(msg, abci.CodeType_InternalError)
}

func DecodingError() TMError {
	return New(msgDecoding, abci.CodeType_EncodingError)
}

func Unauthorized() TMError {
	return New(msgUnauthorized, abci.CodeType_Unauthorized)
}

func MissingSignature() TMError {
	return New(msgMissingSignature, abci.CodeType_Unauthorized)
}

func TooManySignatures() TMError {
	return New(msgTooManySignatures, abci.CodeType_Unauthorized)
}

func InvalidSignature() TMError {
	return New(msgInvalidSignature, abci.CodeType_Unauthorized)
}

func NoChain() TMError {
	return New(msgNoChain, abci.CodeType_Unauthorized)
}

func WrongChain(chain string) TMError {
	msg := fmt.Sprintf(msgWrongChain, chain)
	return New(msg, abci.CodeType_Unauthorized)
}

func InvalidAddress() TMError {
	return New(msgInvalidAddress, abci.CodeType_BaseInvalidInput)
}

func InvalidCoins() TMError {
	return New(msgInvalidCoins, abci.CodeType_BaseInvalidInput)
}

func InvalidFormat() TMError {
	return New(msgInvalidFormat, abci.CodeType_BaseInvalidInput)
}

func InvalidSequence() TMError {
	return New(msgInvalidSequence, abci.CodeType_BaseInvalidInput)
}

func InsufficientFees() TMError {
	return New(msgInsufficientFees, abci.CodeType_BaseInvalidInput)
}

func NoInputs() TMError {
	return New(msgNoInputs, abci.CodeType_BaseInvalidInput)
}

func NoOutputs() TMError {
	return New(msgNoOutputs, abci.CodeType_BaseInvalidOutput)
}

func TooLarge() TMError {
	return New(msgTooLarge, abci.CodeType_EncodingError)
}
