package tx_test

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

func Test_SetMsg(t *testing.T) {
	cases := []struct {
		name   string
		msg    sdk.Msg
		expErr bool
	}{
		{
			name:   "Set nil Msg",
			msg:    nil,
			expErr: true,
		},
		{
			name:   "Set a valid message",
			msg:    &DummyProtoMessage1{},
			expErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tx.SetMsg(tc.msg)
			if tc.expErr {
				require.Error(t, err)
				return
			}

			expected := mustAny(tc.msg)
			require.Equal(t, actual, expected)
			require.Nil(t, err)
		})
	}
}

func Test_SetMsgs(t *testing.T) {
	cases := []struct {
		name   string
		msgs   []sdk.Msg
		expErr bool
	}{
		{
			name:   "Set nil slice of messages",
			msgs:   nil,
			expErr: false,
		},
		{
			name:   "Set empty slice of messages",
			msgs:   []sdk.Msg{},
			expErr: false,
		},
		{
			name:   "Set nil message inside the slice of messages",
			msgs:   []sdk.Msg{nil},
			expErr: true,
		},
		{
			name:   "Set valid messages",
			msgs:   []sdk.Msg{&DummyProtoMessage1{}, &DummyProtoMessage2{}},
			expErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tx.SetMsgs(tc.msgs)
			if tc.expErr {
				require.Error(t, err)
				return
			}

			expected := make([]*types.Any, len(tc.msgs))
			for i, msg := range tc.msgs {
				a := mustAny(msg)
				expected[i] = a
			}

			require.Equal(t, actual, expected)
			require.Nil(t, err)
		})
	}
}

func mustAny(msg proto.Message) *types.Any {
	a, err := types.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return a
}
