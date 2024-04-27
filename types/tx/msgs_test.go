package tx_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
			require.Equal(t, expected, actual)
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

			require.Equal(t, expected, actual)
			require.Nil(t, err)
		})
	}
}

func Test_GetMsgs(t *testing.T) {
	sdkMsgs := []sdk.Msg{&DummyProtoMessage1{}, &DummyProtoMessage2{}}
	anyMsg, err := tx.SetMsgs(sdkMsgs)
	require.Nil(t, err)

	cases := []struct {
		name     string
		msgs     []*types.Any
		expected []sdk.Msg
		expErr   bool
	}{
		{
			name:     "GetMsgs from a nil slice of Any messages",
			msgs:     nil,
			expected: []sdk.Msg{},
			expErr:   false,
		},
		{
			name:     "GetMsgs from empty slice of Any messages",
			msgs:     []*types.Any{},
			expected: []sdk.Msg{},
			expErr:   false,
		},
		{
			name:     "GetMsgs from a slice with valid Any messages",
			msgs:     anyMsg,
			expected: sdkMsgs,
			expErr:   false,
		},
		{
			name:   "GetMsgs from a slice that contains uncashed Any message",
			msgs:   []*types.Any{{}},
			expErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tx.GetMsgs(tc.msgs, "dummy")
			if tc.expErr {
				require.Error(t, err)
				return
			}

			require.Equal(t, tc.expected, actual)
			require.Nil(t, err)
		})
	}
}

func TestTx_UnpackInterfaces(t *testing.T) {
	unpacker := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	sdkMsgs := []sdk.Msg{&DummyProtoMessage1{}, &DummyProtoMessage2{}}
	anyMsg, err := tx.SetMsgs(sdkMsgs)
	require.Nil(t, err)

	cases := []struct {
		name   string
		msgs   []*types.Any
		expErr bool
	}{
		{
			name:   "Unpack nil slice messages",
			msgs:   nil,
			expErr: false,
		},
		{
			name:   "Unpack empty slice of messages",
			msgs:   []*types.Any{},
			expErr: false,
		},
		{
			name:   "Unpack valid messages",
			msgs:   anyMsg,
			expErr: false,
		},
		{
			name:   "Unpack uncashed message",
			msgs:   []*types.Any{{TypeUrl: "uncached"}},
			expErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err = tx.UnpackInterfaces(unpacker, tc.msgs)
			require.Equal(t, tc.expErr, err != nil)
		})
	}
}

// in get cached value there are an error

func mustAny(msg proto.Message) *types.Any {
	a, err := types.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return a
}
