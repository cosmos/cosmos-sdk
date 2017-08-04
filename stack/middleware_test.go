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
)

const (
	nameSigner = "signer"
)

func TestPermissionSandbox(t *testing.T) {
	require := require.New(t)

	// generic args
	ctx := NewContext("test-chain", 20, log.NewNopLogger())
	store := state.NewMemKVStore()
	raw := NewRawTx([]byte{1, 2, 3, 4})
	rawBytes, err := data.ToWire(raw)
	require.Nil(err)

	// test cases to make sure permissioning is solid
	grantee := basecoin.Actor{App: NameGrant, Address: []byte{1}}
	grantee2 := basecoin.Actor{App: NameGrant, Address: []byte{2}}
	// ibc and grantee are the same, just different chains
	ibc := basecoin.Actor{ChainID: "other", App: NameGrant, Address: []byte{1}}
	ibc2 := basecoin.Actor{ChainID: "other", App: nameSigner, Address: []byte{21}}
	signer := basecoin.Actor{App: nameSigner, Address: []byte{21}}
	cases := []struct {
		asIBC       bool
		grant       basecoin.Actor
		require     basecoin.Actor
		expectedRes data.Bytes
		expected    func(error) bool
	}{
		// grant as normal app middleware
		{false, grantee, grantee, rawBytes, nil},
		{false, grantee, grantee2, nil, errors.IsUnauthorizedErr},
		{false, grantee2, grantee2, rawBytes, nil},
		{false, ibc, grantee, nil, errors.IsInternalErr},
		{false, grantee, ibc, nil, errors.IsUnauthorizedErr},
		{false, grantee, signer, nil, errors.IsUnauthorizedErr},
		{false, signer, signer, nil, errors.IsInternalErr},

		// grant as ibc middleware
		{true, ibc, ibc, rawBytes, nil},   // ibc can set permissions
		{true, ibc2, ibc2, rawBytes, nil}, // for any app
		// the must match, both app and chain
		{true, ibc, ibc2, nil, errors.IsUnauthorizedErr},
		{true, ibc, grantee, nil, errors.IsUnauthorizedErr},
		// cannot set local apps from ibc middleware
		{true, grantee, grantee, nil, errors.IsInternalErr},
	}

	for i, tc := range cases {
		app := New(Recovery{})
		if tc.asIBC {
			app = app.IBC(GrantMiddleware{Auth: tc.grant})
		} else {
			app = app.Apps(GrantMiddleware{Auth: tc.grant})
		}
		app = app.
			Apps(CheckMiddleware{Required: tc.require}).
			Use(EchoHandler{})

		cres, err := app.CheckTx(ctx, store, raw)
		checkPerm(t, i, tc.expectedRes, tc.expected, cres, err)

		dres, err := app.DeliverTx(ctx, store, raw)
		checkPerm(t, i, tc.expectedRes, tc.expected, dres, err)
	}
}

func checkPerm(t *testing.T, idx int, data []byte, check func(error) bool, res basecoin.Result, err error) {
	assert := assert.New(t)

	if len(data) > 0 {
		assert.Nil(err, "%d: %+v", idx, err)
		assert.EqualValues(data, res.GetData())
	} else {
		assert.NotNil(err, "%d", idx)
		// check error code!
		assert.True(check(err), "%d: %+v", idx, err)
	}
}
