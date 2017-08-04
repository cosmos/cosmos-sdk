package roles_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

func TestCreateRole(t *testing.T) {
	assert := assert.New(t)

	a := basecoin.Actor{App: "foo", Address: []byte("bar")}
	b := basecoin.Actor{ChainID: "eth", App: "foo", Address: []byte("bar")}
	c := basecoin.Actor{App: "foo", Address: []byte("baz")}
	d := basecoin.Actor{App: "si-ly", Address: []byte("bar")}

	cases := []struct {
		valid bool
		role  string
		min   uint32
		sigs  []basecoin.Actor
	}{
		{true, "awesome", 1, []basecoin.Actor{a}},
		{true, "cool", 2, []basecoin.Actor{b, c, d}},
		{false, "oops", 3, []basecoin.Actor{a, d}}, // too many
		{false, "ugh", 0, []basecoin.Actor{a, d}},  // too few
		{false, "phew", 1, []basecoin.Actor{}},     // none
		{false, "cool", 1, []basecoin.Actor{c, d}}, // duplicate of existing one
	}

	h := roles.NewHandler()
	ctx := stack.MockContext("role-chain", 123)
	store := state.NewMemKVStore()
	for i, tc := range cases {
		tx := roles.NewCreateRoleTx([]byte(tc.role), tc.min, tc.sigs)
		cres, err := h.CheckTx(ctx, store, tx)
		_, err2 := h.DeliverTx(ctx, store, tx)
		if tc.valid {
			assert.Nil(err, "%d/%s: %+v", i, tc.role, err)
			assert.Nil(err2, "%d/%s: %+v", i, tc.role, err2)
			assert.Equal(roles.CostCreate, cres.GasAllocated)
			assert.Equal(uint64(0), cres.GasPayment)
		} else {
			assert.NotNil(err, "%d/%s", i, tc.role)
			assert.NotNil(err2, "%d/%s", i, tc.role)
		}
	}
}
