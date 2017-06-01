package errors

/**
*    Copyright (C) 2017 Ethan Frey
**/

import abci "github.com/tendermint/abci/types"

const (
	msgDecoding          = "Error decoding input"
	msgUnauthorized      = "Unauthorized"
	msgInvalidAddress    = "Invalid Address"
	msgInvalidCoins      = "Invalid Coins"
	msgInvalidSequence   = "Invalid Sequence"
	msgInvalidSignature  = "Invalid Signature"
	msgNoInputs          = "No Input Coins"
	msgNoOutputs         = "No Output Coins"
	msgTooLarge          = "Input size too large"
	msgMissingSignature  = "Signature missing"
	msgTooManySignatures = "Too many signatures"
)

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

func InvalidAddress() TMError {
	return New(msgInvalidAddress, abci.CodeType_BaseInvalidInput)
}

func InvalidCoins() TMError {
	return New(msgInvalidCoins, abci.CodeType_BaseInvalidInput)
}

func InvalidSequence() TMError {
	return New(msgInvalidSequence, abci.CodeType_BaseInvalidInput)
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
