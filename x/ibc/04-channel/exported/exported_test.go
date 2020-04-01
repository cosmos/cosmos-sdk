package exported

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelStateString(t *testing.T) {
	cases := []struct {
		name  string
		state State
	}{
		{StateUninitialized, UNINITIALIZED},
		{StateInit, INIT},
		{StateTryOpen, TRYOPEN},
		{StateOpen, OPEN},
		{StateClosed, CLOSED},
	}

	for _, tt := range cases {
		tt := tt
		require.Equal(t, tt.state, StateFromString(tt.name))
		require.Equal(t, tt.name, tt.state.String())
	}
}

func TestChannelStateMarshalJSON(t *testing.T) {
	cases := []struct {
		name  string
		state State
	}{
		{StateUninitialized, UNINITIALIZED},
		{StateInit, INIT},
		{StateTryOpen, TRYOPEN},
		{StateOpen, OPEN},
		{StateClosed, CLOSED},
	}

	for _, tt := range cases {
		tt := tt
		bz, err := tt.state.MarshalJSON()
		require.NoError(t, err)
		var state State
		require.NoError(t, state.UnmarshalJSON(bz))
		require.Equal(t, tt.name, state.String())
	}
}

func TestOrderString(t *testing.T) {
	cases := []struct {
		name  string
		order Order
	}{
		{OrderNone, NONE},
		{OrderUnordered, UNORDERED},
		{OrderOrdered, ORDERED},
	}

	for _, tt := range cases {
		tt := tt
		require.Equal(t, tt.order, OrderFromString(tt.name))
		require.Equal(t, tt.name, tt.order.String())
	}
}

func TestOrderMarshalJSON(t *testing.T) {
	cases := []struct {
		msg        string
		name       string
		order      Order
		expectPass bool
	}{
		{"none ordering should have failed", OrderNone, NONE, false},
		{"unordered should have passed", OrderUnordered, UNORDERED, true},
		{"ordered should have passed", OrderOrdered, ORDERED, true},
	}

	for _, tt := range cases {
		tt := tt
		bz, err := tt.order.MarshalJSON()
		require.NoError(t, err)
		var order Order
		if tt.expectPass {
			require.NoError(t, order.UnmarshalJSON(bz), tt.msg)
			require.Equal(t, tt.name, order.String(), tt.msg)
		} else {
			require.Error(t, order.UnmarshalJSON(bz), tt.msg)
		}
	}
}
