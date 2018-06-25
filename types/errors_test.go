package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var codeTypes = []CodeType{
	CodeInternal,
	CodeTxDecode,
	CodeInvalidSequence,
	CodeUnauthorized,
	CodeInsufficientFunds,
	CodeUnknownRequest,
	CodeUnknownAddress,
	CodeInvalidPubKey,
}

type errFn func(msg string) Error

var errFns = []errFn{
	ErrInternal,
	ErrTxDecode,
	ErrInvalidSequence,
	ErrUnauthorized,
	ErrInsufficientFunds,
	ErrUnknownRequest,
	ErrUnknownAddress,
	ErrInvalidPubKey,
}

func TestCodeType(t *testing.T) {
	assert.True(t, ABCICodeOK.IsOK())

	for _, c := range codeTypes {
		msg := CodeToDefaultMsg(c)
		assert.False(t, strings.HasPrefix(msg, "Unknown code"))
	}
}

func TestCodeConverter(t *testing.T) {
	space := CodespaceType(53)
	code := CodeType(32)

	abciCode := ToABCICode(space, code)

	localSpace, localCode := ToLocalCode(abciCode)

	assert.Equal(t, space, localSpace, "ToLocalCode did not convert to local codespace correctly")
	assert.Equal(t, code, localCode, "ToLocalCode did not convert to local code correctly")
}

func TestErrFn(t *testing.T) {
	for i, errFn := range errFns {
		err := errFn("")
		codeType := codeTypes[i]
		assert.Equal(t, err.Code(), codeType)
		assert.Equal(t, err.Result().Code, ToABCICode(CodespaceRoot, codeType))
	}
}
