package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/txs"
)

func TestPermissionSandbox(t *testing.T) {
	require := require.New(t)

	// generic args
	ctx := NewContext("test-chain", log.NewNopLogger())
	store := state.NewMemKVStore()
	raw := txs.NewRaw([]byte{1, 2, 3, 4})
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
		expected    func(error) bool
	}{
		{grantee, grantee, rawBytes, nil},
		{grantee, grantee2, nil, errors.IsUnauthorizedErr},
		{grantee, signer, nil, errors.IsUnauthorizedErr},
		{signer, signer, nil, errors.IsInternalErr},
	}

	for i, tc := range cases {
		app := New(
			Recovery{}, // we need this so panics turn to errors
			GrantMiddleware{Auth: tc.grant},
			CheckMiddleware{Required: tc.require},
		).Use(EchoHandler{})

		res, err := app.CheckTx(ctx, store, raw)
		checkPerm(t, i, tc.expectedRes, tc.expected, res, err)

		res, err = app.DeliverTx(ctx, store, raw)
		checkPerm(t, i, tc.expectedRes, tc.expected, res, err)
	}
}

func checkPerm(t *testing.T, idx int, data []byte, check func(error) bool, res basecoin.Result, err error) {
	assert := assert.New(t)

	if len(data) > 0 {
		assert.Nil(err, "%d: %+v", idx, err)
		assert.EqualValues(data, res.Data)
	} else {
		assert.NotNil(err, "%d", idx)
		// check error code!
		assert.True(check(err), "%d: %+v", idx, err)
	}
}
