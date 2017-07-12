package roles_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

// shortcut for the lazy
type ba []basecoin.Actor

func createRole(app basecoin.Handler, store state.KVStore,
	name []byte, min uint32, sigs ...basecoin.Actor) (basecoin.Actor, error) {
	tx := roles.NewCreateRoleTx(name, min, sigs)
	ctx := stack.MockContext("foo", 1)
	_, err := app.DeliverTx(ctx, store, tx)
	return roles.NewPerm(name), err
}

func TestAssumeRole(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// one handle to add a role, another to check permissions
	disp := stack.NewDispatcher(
		stack.WrapHandler(roles.NewHandler()),
		stack.WrapHandler(stack.CheckHandler{}),
	)
	// and wrap with the roles middleware
	app := stack.New(roles.NewMiddleware()).Use(disp)

	// basic state for the app
	ctx := stack.MockContext("role-chain", 123)
	store := state.NewMemKVStore()

	// potential actors
	a := basecoin.Actor{App: "sig", Address: []byte("jae")}
	b := basecoin.Actor{App: "sig", Address: []byte("bucky")}
	c := basecoin.Actor{App: "sig", Address: []byte("ethan")}
	d := basecoin.Actor{App: "tracko", Address: []byte("rigel")}

	// devs is a 2-of-3 multisig
	devs := data.Bytes{0, 1, 0, 1}
	pdev, err := createRole(app, store, devs, 2, b, c, d)
	require.Nil(err)

	// deploy requires a dev role, or supreme authority
	deploy := data.Bytes("deploy")
	_, err = createRole(app, store, deploy, 1, a, pdev)
	require.Nil(err)

	// now, let's test the roles are set properly
	cases := []struct {
		valid    bool
		roles    []data.Bytes     // which roles we try to assume (can be multiple!)
		signers  []basecoin.Actor // which people sign the  tx
		required []basecoin.Actor // which permission we require to succeed
	}{
		// basic checks to see logic works
		{true, nil, nil, nil},
		{true, nil, ba{b, c}, ba{b}},
		{false, nil, ba{b}, ba{b, c}},

		// simple role check
	}

	for i, tc := range cases {
		// set the signers, the required check
		myCtx := ctx.WithPermissions(tc.signers...)
		tx := stack.NewCheckTx(tc.required)
		// and the roles we attempt to assume
		for _, r := range tc.roles {
			tx = roles.NewAssumeRoleTx(r, tx)
		}

		// try CheckTx and DeliverTx and make sure they both assert permissions
		_, err := app.CheckTx(myCtx, store, tx)
		_, err2 := app.DeliverTx(myCtx, store, tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
			assert.Nil(err2, "%d: %+v", i, err2)
		} else {
			assert.NotNil(err, "%d", i)
			assert.NotNil(err2, "%d", i)
		}
	}
}
