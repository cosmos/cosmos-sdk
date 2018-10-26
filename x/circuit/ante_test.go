package circuit

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params"
)

type tx []sdk.Msg

func (tx tx) GetMsgs() []sdk.Msg { return tx }

type msg struct{}

func (msg) ValidateBasic() sdk.Error { return nil }

func (msg) GetSignBytes() []byte { return nil }

func (msg) GetSigners() []sdk.AccAddress { return nil }

type msg1 struct{ msg }

func (msg1) Type() string { return "msg" }

func (msg1) Name() string { return "msg1" }

type msg2 struct{ msg }

func (msg2) Type() string { return "msg" }

func (msg2) Name() string { return "msg2" }

type othermsg struct{ msg }

func (othermsg) Type() string { return "othermsg" }

func (othermsg) Name() string { return "othermsg" }

func testMsg(t *testing.T, ctx sdk.Context, k Keeper, msg sdk.Msg, ty bool, name bool) {
	ante := NewAnteHandler(k)

	_, _, abort := ante(ctx, tx{msg}, false)
	require.Equal(t, ty || name, abort)

	table := []struct {
		ty   bool
		name bool
	}{
		{false, false},
		{false, true},
		{true, false},
		{true, true},
	}

	for i, tc := range table {
		require.NotPanics(t, func() { k.space.SetWithSubkey(ctx, MsgTypeKey, []byte(msg.Type()), tc.ty) }, "panic setting breaker, tc #%d", i)
		require.NotPanics(t, func() { k.space.SetWithSubkey(ctx, MsgNameKey, []byte(msg.Name()), tc.name) }, "panic setting breaker, tc #%d", i)

		_, _, abort := ante(ctx, tx{msg}, false)
		require.Equal(t, tc.ty || tc.name, abort)
	}
}

func TestAnteHandler(t *testing.T) {
	ctx, space, _ := params.DefaultTestComponents(t)

	k := NewKeeper(space)

	data := GenesisState{
		MsgTypes: []string{"othermsg"},
		MsgNames: []string{"msg2"},
	}

	InitGenesis(ctx, k, data)

	testMsg(t, ctx, k, msg1{}, false, false)
	testMsg(t, ctx, k, msg2{}, false, true)
	testMsg(t, ctx, k, othermsg{}, true, false)
}
