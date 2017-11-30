package errors

import (
	stderr "errors"
	"strconv"
	"testing"

	pkerr "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
)

func TestCreateResult(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		err  error
		msg  string
		code abci.CodeType
	}{
		{stderr.New("base"), "base", defaultErrCode},
		{pkerr.New("dave"), "dave", defaultErrCode},
		{New("nonce", abci.CodeType_BadNonce), "nonce", abci.CodeType_BadNonce},
		{Wrap(stderr.New("wrap")), "wrap", defaultErrCode},
		{WithCode(stderr.New("coded"), abci.CodeType_BaseInvalidInput), "coded", abci.CodeType_BaseInvalidInput},
		{ErrDecoding(), errDecoding.Error(), abci.CodeType_EncodingError},
		{ErrUnauthorized(), errUnauthorized.Error(), abci.CodeType_Unauthorized},
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
