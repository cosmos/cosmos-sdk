package tx_test

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

func Test_SetMsg(t *testing.T) {
	cases := map[string]struct {
		msg    sdk.Msg
		expErr bool
	}{
		"Set nil Msg": {
			msg:    nil,
			expErr: true,
		},
		"Set a valid message": {
			msg:    &DummyProtoMessage1{},
			expErr: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actual, err := tx.SetMsg(tc.msg)
			if tc.expErr {
				require.Error(t, err)
				return
			}

			expected := mustAny(tc.msg)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}
}

func Test_SetMsgs(t *testing.T) {
	cases := map[string]struct {
		msgs   []sdk.Msg
		expErr bool
	}{
		"Set nil slice of messages": {
			msgs:   nil,
			expErr: false,
		},
		"Set empty slice of messages": {
			msgs:   []sdk.Msg{},
			expErr: false,
		},
		"Set nil message inside the slice of messages": {
			msgs:   []sdk.Msg{nil},
			expErr: true,
		},
		"Set valid messages": {
			msgs:   []sdk.Msg{&DummyProtoMessage1{}, &DummyProtoMessage2{}},
			expErr: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
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

			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}
}

func Test_GetMsgs(t *testing.T) {
	sdkMsgs := []sdk.Msg{&DummyProtoMessage1{}, &DummyProtoMessage2{}}
	anyMsg, err := tx.SetMsgs(sdkMsgs)
	require.NoError(t, err)

	cases := map[string]struct {
		msgs     []*types.Any
		expected []sdk.Msg
		expErr   bool
	}{
		"GetMsgs from a nil slice of Any messages": {
			msgs:     nil,
			expected: []sdk.Msg{},
			expErr:   false,
		},
		"GetMsgs from empty slice of Any messages": {
			msgs:     []*types.Any{},
			expected: []sdk.Msg{},
			expErr:   false,
		},
		"GetMsgs from a slice with valid Any messages": {
			msgs:     anyMsg,
			expected: sdkMsgs,
			expErr:   false,
		},
		"GetMsgs from a slice that contains uncached Any message": {
			msgs:   []*types.Any{{}},
			expErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actual, err := tx.GetMsgs(tc.msgs, "dummy")
			if tc.expErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestTx_UnpackInterfaces(t *testing.T) {
	unpacker := codec.NewProtoCodec(types.NewInterfaceRegistry())
	sdkMsgs := []sdk.Msg{&DummyProtoMessage1{}, &DummyProtoMessage2{}}
	anyMsg, err := tx.SetMsgs(sdkMsgs)
	require.NoError(t, err)

	cases := map[string]struct {
		msgs   []*types.Any
		expErr bool
	}{
		"Unpack nil slice messages": {
			msgs:   nil,
			expErr: false,
		},
		"Unpack empty slice of messages": {
			msgs:   []*types.Any{},
			expErr: false,
		},
		"Unpack valid messages": {
			msgs:   anyMsg,
			expErr: false,
		},
		"Unpack uncashed message": {
			msgs:   []*types.Any{{TypeUrl: "uncached"}},
			expErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err = tx.UnpackInterfaces(unpacker, tc.msgs)
			require.Equal(t, tc.expErr, err != nil)
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
