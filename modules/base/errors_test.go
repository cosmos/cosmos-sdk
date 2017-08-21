package base

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/errors"
)

func TestErrorMatches(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		pattern, err error
		match        bool
	}{
		{errWrongChain, ErrWrongChain("hakz"), true},
	}

	for i, tc := range cases {
		same := errors.IsSameError(tc.pattern, tc.err)
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
		// make sure WrongChain works properly
		{ErrWrongChain("fooz"), errors.IsUnauthorizedErr, true},
		{ErrWrongChain("barz"), IsWrongChainErr, true},
	}

	for i, tc := range cases {
		match := tc.check(tc.err)
		assert.Equal(tc.match, match, "%d", i)
	}
}
