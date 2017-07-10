package roles_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
)

func TestRole(t *testing.T) {
	assert := assert.New(t)

	// prepare some actors...
	a := basecoin.Actor{App: "foo", Address: []byte("bar")}
	b := basecoin.Actor{ChainID: "eth", App: "foo", Address: []byte("bar")}
	c := basecoin.Actor{App: "foo", Address: []byte("baz")}
	d := basecoin.Actor{App: "si-ly", Address: []byte("bar")}
	e := basecoin.Actor{App: "si-ly", Address: []byte("big")}
	f := basecoin.Actor{App: "sig", Address: []byte{1}}
	g := basecoin.Actor{App: "sig", Address: []byte{2, 3, 4}}

	cases := []struct {
		sigs    uint32
		allowed []basecoin.Actor
		signers []basecoin.Actor
		valid   bool
	}{
		// make sure simple compare is correct
		{1, []basecoin.Actor{a}, []basecoin.Actor{a}, true},
		{1, []basecoin.Actor{a}, []basecoin.Actor{b}, false},
		{1, []basecoin.Actor{a}, []basecoin.Actor{c}, false},
		{1, []basecoin.Actor{a}, []basecoin.Actor{d}, false},
		// make sure multi-sig counts to 1
		{1, []basecoin.Actor{a, b, c}, []basecoin.Actor{d, e, a, f}, true},
		{1, []basecoin.Actor{a, b, c}, []basecoin.Actor{a, b, c, d}, true},
		{1, []basecoin.Actor{a, b, c}, []basecoin.Actor{d, e, f}, false},
		// make sure multi-sig counts higher
		{2, []basecoin.Actor{b, e, g}, []basecoin.Actor{g, c, a, d, b}, true},
		{2, []basecoin.Actor{b, e, g}, []basecoin.Actor{c, a, d, b}, false},
		{3, []basecoin.Actor{a, b, c}, []basecoin.Actor{g}, false},
	}

	for idx, tc := range cases {
		i := strconv.Itoa(idx)
		// make sure IsSigner works
		role := roles.NewRole(tc.sigs, tc.allowed)
		for _, a := range tc.allowed {
			assert.True(role.IsSigner(a), i)
		}
		// make sure IsAuthorized works
		ctx := stack.MockContext("chain-id", 100).WithPermissions(tc.signers...)
		allowed := role.IsAuthorized(ctx)
		assert.Equal(tc.valid, allowed, i)
	}

}
