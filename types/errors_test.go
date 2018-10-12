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
		require.Equal(t, err.Codespace(), CodespaceRoot, "Err function expected to return proper codespace. tc #%d", i)
		require.Equal(t, err.Result().Code, ToABCICode(CodespaceRoot, codeType), "Err function expected to return proper ABCICode. tc #%d")
		require.Equal(t, err.QueryResult().Code, uint32(err.ABCICode()), "Err function expected to return proper ABCICode from QueryResult. tc #%d")
		require.Equal(t, err.QueryResult().Log, err.ABCILog(), "Err function expected to return proper ABCILog from QueryResult. tc #%d")
	}

	require.Equal(t, ABCICodeOK, ToABCICode(CodespaceRoot, CodeOK))
}

func TestErrorFormat(t *testing.T) {
	// default err msg
	err := errFns[0]("")
	msg := err.Stacktrace().Error()
	require.NotPanicsf(t, func() { ErrMustHaveValidFormat(msg) }, "Should have a valid format")

	// custom err msg
	err = errFns[1]("this is a custom error")
	msg = err.Stacktrace().Error()
	require.NotPanicsf(t, func() { ErrMustHaveValidFormat(msg) }, "Should have a valid format")

	// custom err msg with aditional value
	err = errFns[2]("")
	msg = AppendMsgToErr("Error", err.Stacktrace().Error())
	require.NotPanicsf(t, func() { ErrMustHaveValidFormat(msg) }, "Should have a valid format")

	// unexpected err msg
	err = errFns[3]("")
	require.Panicsf(t, func() { ErrMustHaveValidFormat(err.ABCILog()) }, "Shouldn't have a valid format")
}

func TestGetABCILogMsg(t *testing.T) {
	for i, errFn := range errFns {
		err := errFn("")
		msg := MustGetABCILogMsg(err.ABCILog())
		require.Equal(t, err.Stacktrace().Error(), msg, "Err function expected to return the 'message' value from the ABCI Log. tc #%d", i)
	}

	// invalid format
	err := errFns[0]("")
	require.Panicsf(t, func() { MustGetABCILogMsg(err.Stacktrace().Error()) }, "Shouldn't have a valid format")
}
