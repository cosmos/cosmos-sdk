package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var codeTypes = []CodeType{
	CodeInternal,
	CodeTxParse,
	CodeInvalidSequence,
	CodeUnauthorized,
	CodeInsufficientFunds,
	CodeUnknownRequest,
	CodeUnknownAddress,
	CodeInvalidPubKey,
	CodeGenesisParse,
}

type errFn func(msg string) Error

var errFns = []errFn{
	ErrInternal,
	ErrTxParse,
	ErrInvalidSequence,
	ErrUnauthorized,
	ErrInsufficientFunds,
	ErrUnknownRequest,
	ErrUnknownAddress,
	ErrInvalidPubKey,
	ErrGenesisParse,
}

func TestCodeType(t *testing.T) {
	assert.True(t, CodeOK.IsOK())

	for _, c := range codeTypes {
		assert.False(t, c.IsOK())
		msg := CodeToDefaultMsg(c)
		assert.False(t, strings.HasPrefix(msg, "Unknown code"))
	}
}

func TestErrFn(t *testing.T) {
	for i, errFn := range errFns {
		err := errFn("")
		codeType := codeTypes[i]
		assert.Equal(t, err.ABCICode(), codeType)
		assert.Equal(t, err.Result().Code, codeType)
	}
}
