package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var codeTypes = []CodeType{
	CodeInternal,
	CodeTxDecode,
	CodeInvalidSequence,
	CodeUnauthorized,
	CodeInsufficientFunds,
	CodeUnknownRequest,
	CodeInvalidAddress,
	CodeInvalidPubKey,
	CodeUnknownAddress,
	CodeInsufficientCoins,
	CodeInvalidCoins,
	CodeOutOfGas,
	CodeMemoTooLarge,
}

type errFn func(msg string) Error

var errFns = []errFn{
	ErrInternal,
	ErrTxDecode,
	ErrInvalidSequence,
	ErrUnauthorized,
	ErrInsufficientFunds,
	ErrUnknownRequest,
	ErrInvalidAddress,
	ErrInvalidPubKey,
	ErrUnknownAddress,
	ErrInsufficientCoins,
	ErrInvalidCoins,
	ErrOutOfGas,
	ErrMemoTooLarge,
}

func TestCodeType(t *testing.T) {
	require.True(t, ABCICodeOK.IsOK())

	for tcnum, c := range codeTypes {
		msg := CodeToDefaultMsg(c)
		require.NotEqual(t, unknownCodeMsg(c), msg, "Code expected to be known. tc #%d, code %d, msg %s", tcnum, c, msg)
	}

	msg := CodeToDefaultMsg(CodeOK)
	require.Equal(t, unknownCodeMsg(CodeOK), msg)
}

func TestErrFn(t *testing.T) {
	for i, errFn := range errFns {
		err := errFn("")
		codeType := codeTypes[i]
		require.Equal(t, err.Code(), codeType, "Err function expected to return proper code. tc #%d", i)
		require.Equal(t, err.Result().Code, ToABCICode(CodespaceRoot, codeType), "Err function expected to return proper ABCICode. tc #%d")
	}

	require.Equal(t, ABCICodeOK, ToABCICode(CodespaceRoot, CodeOK))
}
