package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/errors"
)

func TestChecks(t *testing.T) {
	// TODO: make sure the Is and Err methods match
	assert := assert.New(t)

	cases := []struct {
		err   error
		check func(error) bool
		match bool
	}{
		// unauthorized includes InvalidSignature, but not visa versa
		{ErrInvalidSignature(), IsInvalidSignatureErr, true},
		{ErrInvalidSignature(), errors.IsUnauthorizedErr, true},
	}

	for i, tc := range cases {
		match := tc.check(tc.err)
		assert.Equal(tc.match, match, "%d", i)
	}
}
