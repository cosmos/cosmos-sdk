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

func createRole(app basecoin.Handler, store state.SimpleDB,
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
	// shows how we can build larger constructs, eg. (A and B) OR C
	deploy := data.Bytes("deploy")
	pdeploy, err := createRole(app, store, deploy, 1, a, pdev)
	require.Nil(err)

	// now, let's test the roles are set properly
	cases := []struct {
		valid bool
		// which roles we try to assume (can be multiple!)
		// note: that wrapping is FILO, so tries to assume last role first
		roles    []data.Bytes
		signers  []basecoin.Actor // which people sign the  tx
		required []basecoin.Actor // which permission we require to succeed
	}{
		// basic checks to see logic works
		{true, nil, nil, nil},
		{true, nil, ba{b, c}, ba{b}},
		{false, nil, ba{b}, ba{b, c}},

		// simple role check
		{false, []data.Bytes{devs}, ba{a, b}, ba{pdev}},        // not enough sigs
		{false, nil, ba{b, c}, ba{pdev}},                       // must explicitly request group status
		{true, []data.Bytes{devs}, ba{b, c}, ba{pdev}},         // ahh... better
		{true, []data.Bytes{deploy}, ba{a, b}, ba{b, pdeploy}}, // deploy also works

		// multiple levels of roles - must be in correct order - assume dev, then deploy
		{false, []data.Bytes{devs, deploy}, ba{c, d}, ba{pdeploy}},
		{true, []data.Bytes{deploy, devs}, ba{c, d}, ba{pdev, pdeploy}},
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
		cres, err := app.CheckTx(myCtx, store, tx)
		_, err2 := app.DeliverTx(myCtx, store, tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
			assert.Nil(err2, "%d: %+v", i, err2)
			// make sure we charge for each role
			assert.Equal(roles.CostAssume*uint(len(tc.roles)), cres.GasAllocated)
			assert.Equal(uint(0), cres.GasPayment)
		} else {
			assert.NotNil(err, "%d", i)
			assert.NotNil(err2, "%d", i)
		}
	}
}
