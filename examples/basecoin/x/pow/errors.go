package pow

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	CodeInvalidDifficulty     CodeType = 201
	CodeNonexistentDifficulty CodeType = 202
	CodeNonexistentReward     CodeType = 203
	CodeNonexistentCount      CodeType = 204
	CodeInvalidProof          CodeType = 205
	CodeNotBelowTarget        CodeType = 206
	CodeInvalidCount          CodeType = 207
	CodeUnknownRequest        CodeType = sdk.CodeUnknownRequest
)

func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidDifficulty:
		return "Insuffient difficulty"
	case CodeNonexistentDifficulty:
		return "Nonexistent difficulty"
	case CodeNonexistentReward:
		return "Nonexistent reward"
	case CodeNonexistentCount:
		return "Nonexistent count"
	case CodeInvalidProof:
		return "Invalid proof"
	case CodeNotBelowTarget:
		return "Not below target"
	case CodeInvalidCount:
		return "Invalid count"
	case CodeUnknownRequest:
		return "Unknown request"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

func ErrInvalidDifficulty(msg string) sdk.Error {
	return newError(CodeInvalidDifficulty, msg)
}

func ErrNonexistentDifficulty() sdk.Error {
	return newError(CodeNonexistentDifficulty, "")
}

func ErrNonexistentReward() sdk.Error {
	return newError(CodeNonexistentReward, "")
}

func ErrNonexistentCount() sdk.Error {
	return newError(CodeNonexistentCount, "")
}

func ErrInvalidProof(msg string) sdk.Error {
	return newError(CodeInvalidProof, msg)
}

func ErrNotBelowTarget(msg string) sdk.Error {
	return newError(CodeNotBelowTarget, msg)
}

func ErrInvalidCount(msg string) sdk.Error {
	return newError(CodeInvalidCount, msg)
}

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	} else {
		return codeToDefaultMsg(code)
	}
}

func newError(code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(code, msg)
}
