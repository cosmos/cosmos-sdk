package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/txs"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"
)

func TestPermissionSandbox(t *testing.T) {
	require := require.New(t)

	// generic args
	ctx := NewContext(log.NewNopLogger())
	store := types.NewMemKVStore()
	raw := txs.NewRaw([]byte{1, 2, 3, 4}).Wrap()
	rawBytes, err := data.ToWire(raw)
	require.Nil(err)

	// test cases to make sure permissioning is solid
	grantee := basecoin.Actor{App: NameGrant, Address: []byte{1}}
	grantee2 := basecoin.Actor{App: NameGrant, Address: []byte{2}}
	signer := basecoin.Actor{App: NameSigs, Address: []byte{1}}
	cases := []struct {
		grant       basecoin.Actor
		require     basecoin.Actor
		expectedRes data.Bytes
		expectedErr error
	}{
		{grantee, grantee, rawBytes, nil},
		{grantee, grantee2, nil, errors.Unauthorized()},
		{grantee, signer, nil, errors.Unauthorized()},
		{signer, signer, nil, errors.InternalError("panic")},
	}

	for i, tc := range cases {
		app := New(
			Recovery{}, // we need this so panics turn to errors
			GrantMiddleware{tc.grant},
			CheckMiddleware{tc.require},
		).Use(EchoHandler{})

		res, err := app.CheckTx(ctx, store, raw)
		checkPerm(t, i, tc.expectedRes, tc.expectedErr, res, err)

		res, err = app.DeliverTx(ctx, store, raw)
		checkPerm(t, i, tc.expectedRes, tc.expectedErr, res, err)
	}
}

func checkPerm(t *testing.T, idx int, data []byte, expected error, res basecoin.Result, err error) {
	assert := assert.New(t)

	if expected == nil {
		assert.Nil(err, "%d: %+v", idx, err)
		assert.EqualValues(data, res.Data)
	} else {
		assert.NotNil(err, "%d", idx)
		// check error code!
		shouldCode := errors.Wrap(expected).ErrorCode()
		isCode := errors.Wrap(err).ErrorCode()
		assert.Equal(shouldCode, isCode, "%d: %+v", idx, err)
	}
}
