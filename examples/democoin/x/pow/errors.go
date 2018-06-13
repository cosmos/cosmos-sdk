package pow

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO remove, seems hacky
type CodeType = sdk.CodeType

// POW errors reserve 200 ~ 299
const (
	DefaultCodespace          sdk.CodespaceType = 5
	CodeInvalidDifficulty     CodeType          = 201
	CodeNonexistentDifficulty CodeType          = 202
	CodeNonexistentReward     CodeType          = 203
	CodeNonexistentCount      CodeType          = 204
	CodeInvalidProof          CodeType          = 205
	CodeNotBelowTarget        CodeType          = 206
	CodeInvalidCount          CodeType          = 207
	CodeUnknownRequest        CodeType          = sdk.CodeUnknownRequest
)

func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidDifficulty:
		return "insuffient difficulty"
	case CodeNonexistentDifficulty:
		return "nonexistent difficulty"
	case CodeNonexistentReward:
		return "nonexistent reward"
	case CodeNonexistentCount:
		return "nonexistent count"
	case CodeInvalidProof:
		return "invalid proof"
	case CodeNotBelowTarget:
		return "not below target"
	case CodeInvalidCount:
		return "invalid count"
	case CodeUnknownRequest:
		return "unknown request"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

// nolint
func ErrInvalidDifficulty(codespace sdk.CodespaceType, msg string) sdk.Error {
	return newError(codespace, CodeInvalidDifficulty, msg)
}
func ErrNonexistentDifficulty(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeNonexistentDifficulty, "")
}
func ErrNonexistentReward(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeNonexistentReward, "")
}
func ErrNonexistentCount(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeNonexistentCount, "")
}
func ErrInvalidProof(codespace sdk.CodespaceType, msg string) sdk.Error {
	return newError(codespace, CodeInvalidProof, msg)
}
func ErrNotBelowTarget(codespace sdk.CodespaceType, msg string) sdk.Error {
	return newError(codespace, CodeNotBelowTarget, msg)
}
func ErrInvalidCount(codespace sdk.CodespaceType, msg string) sdk.Error {
	return newError(codespace, CodeInvalidCount, msg)
}

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

func newError(codespace sdk.CodespaceType, code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(codespace, code, msg)
}
