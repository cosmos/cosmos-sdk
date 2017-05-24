package errors

/**
*    Copyright (C) 2017 Ethan Frey
**/

import abci "github.com/tendermint/abci/types"

const (
	msgDecoding     = "Error decoding input"
	msgUnauthorized = "Unauthorized"
)

func DecodingError() TMError {
	return New(msgDecoding, abci.CodeType_EncodingError)
}

func Unauthorized() TMError {
	return New(msgUnauthorized, abci.CodeType_Unauthorized)
}
