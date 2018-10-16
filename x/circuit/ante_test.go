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

func (msg) Name() string { return "" }

type msg1 struct{ msg }

func (msg1) Type() string { return "msg1" }

type msg2 struct{ msg }

func (msg2) Type() string { return "msg2" }

func TestAnteHandler(t *testing.T) {
	ctx, space, _ := params.DefaultTestComponents(t, ParamTypeTable())

	data := GenesisState{
		CircuitBreakTypes: []string{"msg2"},
	}

	InitGenesis(ctx, space, data)

	ante := NewAnteHandler(space)

	_, _, abort := ante(ctx, tx{msg1{}}, false)
	require.False(t, abort)

	_, _, abort = ante(ctx, tx{msg2{}}, false)
	require.True(t, abort)

	space.SetWithSubkey(ctx, MsgTypeKey, []byte(msg1{}.Type()), true)
	space.SetWithSubkey(ctx, MsgTypeKey, []byte(msg2{}.Type()), false)

	_, _, abort = ante(ctx, tx{msg1{}}, false)
	require.True(t, abort)
	_, _, abort = ante(ctx, tx{msg2{}}, false)
	require.False(t, abort)
}
