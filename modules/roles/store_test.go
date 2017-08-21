package roles_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	"github.com/cosmos/cosmos-sdk/stack"
)

func TestRole(t *testing.T) {
	assert := assert.New(t)

	// prepare some actors...
	a := sdk.Actor{App: "foo", Address: []byte("bar")}
	b := sdk.Actor{ChainID: "eth", App: "foo", Address: []byte("bar")}
	c := sdk.Actor{App: "foo", Address: []byte("baz")}
	d := sdk.Actor{App: "si-ly", Address: []byte("bar")}
	e := sdk.Actor{App: "si-ly", Address: []byte("big")}
	f := sdk.Actor{App: "sig", Address: []byte{1}}
	g := sdk.Actor{App: "sig", Address: []byte{2, 3, 4}}

	cases := []struct {
		sigs    uint32
		allowed []sdk.Actor
		signers []sdk.Actor
		valid   bool
	}{
		// make sure simple compare is correct
		{1, []sdk.Actor{a}, []sdk.Actor{a}, true},
		{1, []sdk.Actor{a}, []sdk.Actor{b}, false},
		{1, []sdk.Actor{a}, []sdk.Actor{c}, false},
		{1, []sdk.Actor{a}, []sdk.Actor{d}, false},
		// make sure multi-sig counts to 1
		{1, []sdk.Actor{a, b, c}, []sdk.Actor{d, e, a, f}, true},
		{1, []sdk.Actor{a, b, c}, []sdk.Actor{a, b, c, d}, true},
		{1, []sdk.Actor{a, b, c}, []sdk.Actor{d, e, f}, false},
		// make sure multi-sig counts higher
		{2, []sdk.Actor{b, e, g}, []sdk.Actor{g, c, a, d, b}, true},
		{2, []sdk.Actor{b, e, g}, []sdk.Actor{c, a, d, b}, false},
		{3, []sdk.Actor{a, b, c}, []sdk.Actor{g}, false},
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
