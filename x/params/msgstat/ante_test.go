package msgstat

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/space"
)

type tx []sdk.Msg

func (tx tx) GetMsgs() []sdk.Msg { return tx }

type msg struct{}

func (msg) ValidateBasic() sdk.Error { return nil }

func (msg) GetSignBytes() []byte { return nil }

func (msg) GetSigners() []sdk.AccAddress { return nil }

type msg1 struct{ msg }

func (msg1) Type() string { return "msg1" }

type msg2 struct{ msg }

func (msg2) Type() string { return "msg2" }

func TestAnteHandler(t *testing.T) {
	msg1key := space.NewKey("msg1")
	msg2key := space.NewKey("msg2")

	ctx, space, _ := space.DefaultTestComponents(t)

	data := GenesisState{
		ActivatedTypes: []string{"msg1"},
	}

	InitGenesis(ctx, space, data)

	ante := NewAnteHandler(space)

	_, _, abort := ante(ctx, tx{msg1{}})
	require.False(t, abort)
	_, _, abort = ante(ctx, tx{msg2{}})
	require.True(t, abort)

	space.Set(ctx, msg1key, false)
	space.Set(ctx, msg2key, true)

	_, _, abort = ante(ctx, tx{msg1{}})
	require.True(t, abort)
	_, _, abort = ante(ctx, tx{msg2{}})
	require.False(t, abort)
}
