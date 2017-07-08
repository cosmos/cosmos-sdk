package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type validate interface {
	ValidateBasic() error
}

type holder struct {
	validate
}

func (h holder) Unwrap() validate {
	return h.validate
}

type demoTx struct {
	Age int
}

func (t demoTx) Wrap() holder {
	return holder{t}
}

func (t demoTx) ValidateBasic() error {
	return nil
}

func TestErrorMatches(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		pattern, err error
		match        bool
	}{
		{errDecoding, ErrDecoding(), true},
		{errUnauthorized, ErrUnauthorized(), true},
		{errMissingSignature, ErrUnauthorized(), false},
		{errMissingSignature, ErrMissingSignature(), true},
		{errWrongChain, ErrWrongChain("hakz"), true},
		{errUnknownTxType, ErrUnknownTxType(holder{}), true},
		{errUnknownTxType, ErrUnknownTxType("some text here..."), true},
		{errUnknownTxType, ErrUnknownTxType(demoTx{5}.Wrap()), true},
	}

	for i, tc := range cases {
		same := IsSameError(tc.pattern, tc.err)
		assert.Equal(tc.match, same, "%d: %#v / %#v", i, tc.pattern, tc.err)
	}
}

func TestChecks(t *testing.T) {
	// TODO: make sure the Is and Err methods match
	assert := assert.New(t)

	cases := []struct {
		err   error
		check func(error) bool
		match bool
	}{
		{ErrDecoding(), IsDecodingErr, true},
		{ErrUnauthorized(), IsDecodingErr, false},
		{ErrUnauthorized(), IsUnauthorizedErr, true},
		{ErrInvalidSignature(), IsInvalidSignatureErr, true},
		// unauthorized includes InvalidSignature, but not visa versa
		{ErrInvalidSignature(), IsUnauthorizedErr, true},
		{ErrUnauthorized(), IsInvalidSignatureErr, false},
		// make sure WrongChain works properly
		{ErrWrongChain("fooz"), IsUnauthorizedErr, true},
		{ErrWrongChain("barz"), IsWrongChainErr, true},
		// make sure lots of things match InternalErr, but not everything
		{ErrInternal("bad db connection"), IsInternalErr, true},
		{Wrap(errors.New("wrapped")), IsInternalErr, true},
		{ErrUnauthorized(), IsInternalErr, false},
	}

	for i, tc := range cases {
		match := tc.check(tc.err)
		assert.Equal(tc.match, match, "%d", i)
	}
}
