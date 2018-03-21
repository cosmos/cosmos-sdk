package cool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Cool module reserves error 300-399 lawl
	CodeIncorrectCoolAnswer sdk.CodeType = 300
)

// ErrIncorrectCoolAnswer - Error returned upon an incorrect guess
func ErrIncorrectCoolAnswer(answer string) sdk.Error {
	return sdk.NewError(CodeIncorrectCoolAnswer, "Incorrect Cool answer - `"+answer+"'")
}
