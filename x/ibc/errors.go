package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeInvalidSequence sdk.CodeType = 201
)

func ErrInvalidSequence() sdk.Error {
	return sdk.NewError(CodeInvalidSequence, "")
}
