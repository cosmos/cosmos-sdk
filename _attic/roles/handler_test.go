package roles_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

func TestCreateRole(t *testing.T) {
	assert := assert.New(t)

	a := sdk.Actor{App: "foo", Address: []byte("bar")}
	b := sdk.Actor{ChainID: "eth", App: "foo", Address: []byte("bar")}
	c := sdk.Actor{App: "foo", Address: []byte("baz")}
	d := sdk.Actor{App: "si-ly", Address: []byte("bar")}

	cases := []struct {
		valid bool
		role  string
		min   uint32
		sigs  []sdk.Actor
	}{
		{true, "awesome", 1, []sdk.Actor{a}},
		{true, "cool", 2, []sdk.Actor{b, c, d}},
		{false, "oops", 3, []sdk.Actor{a, d}}, // too many
		{false, "ugh", 0, []sdk.Actor{a, d}},  // too few
		{false, "phew", 1, []sdk.Actor{}},     // none
		{false, "cool", 1, []sdk.Actor{c, d}}, // duplicate of existing one
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
