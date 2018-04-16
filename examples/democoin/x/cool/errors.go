package cool

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Cool module reserves error 400-499 lawl
	CodeIncorrectCoolAnswer sdk.CodeType = 400
)

// ErrIncorrectCoolAnswer - Error returned upon an incorrect guess
func ErrIncorrectCoolAnswer(codespace sdk.CodespaceType, answer string) sdk.Error {
	return sdk.NewError(codespace, CodeIncorrectCoolAnswer, fmt.Sprintf("Incorrect cool answer: %v", answer))
}
