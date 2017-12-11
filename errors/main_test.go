package errors

import (
	stderr "errors"
	"strconv"
	"testing"

	pkerr "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCreateResult(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		err  error
		msg  string
		code uint32
	}{
		{stderr.New("base"), "base", CodeTypeInternalErr},
		{pkerr.New("dave"), "dave", CodeTypeInternalErr},
		{New("nonce", CodeTypeBadNonce), "nonce", CodeTypeBadNonce},
		{Wrap(stderr.New("wrap")), "wrap", CodeTypeInternalErr},
		{WithCode(stderr.New("coded"), CodeTypeBaseInvalidInput), "coded", CodeTypeBaseInvalidInput},
		{ErrDecoding(), errDecoding.Error(), CodeTypeEncodingErr},
		{ErrUnauthorized(), errUnauthorized.Error(), CodeTypeUnauthorized},
	}

	for idx, tc := range cases {
		i := strconv.Itoa(idx)

		dres := DeliverResult(tc.err)
		assert.True(dres.IsErr(), i)
		assert.Equal(tc.msg, dres.Log, i)
		assert.Equal(tc.code, dres.Code, i)

		cres := CheckResult(tc.err)
		assert.True(cres.IsErr(), i)
		assert.Equal(tc.msg, cres.Log, i)
		assert.Equal(tc.code, cres.Code, i)
	}
}
