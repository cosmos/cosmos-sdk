package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelStateString(t *testing.T) {
	cases := []struct {
		state State
		name  string
		msg   string
	}{
		{UNINITIALIZED, StateUninitialized, "uninitialized"},
		{INIT, StateInit, "init"},
		{TRYOPEN, StateTryOpen, "tryopen"},
		{OPEN, StateOpen, "open"},
		{CLOSED, StateClosed, "closed"},
	}

	for _, tt := range cases {
		tt := tt
		require.Equal(t, tt.state, StateFromString(tt.name), tt.msg)
		require.Equal(t, tt.name, tt.state.String(), tt.msg)
	}
}

func TestChannelStateMarshalJSON(t *testing.T) {
	cases := []struct {
		state State
		name  string
		msg   string
	}{
		{UNINITIALIZED, StateUninitialized, "uninitialized"},
		{INIT, StateInit, "init"},
		{TRYOPEN, StateTryOpen, "tryopen"},
		{OPEN, StateOpen, "open"},
		{CLOSED, StateClosed, "closed"},
	}

	for _, tt := range cases {
		tt := tt
		bz, err := tt.state.MarshalJSON()
		require.NoError(t, err)
		var state State
		require.NoError(t, state.UnmarshalJSON(bz), tt.msg)
		require.Equal(t, tt.name, state.String(), tt.msg)
	}
}

func TestOrderString(t *testing.T) {
	cases := []struct {
		order Order
		name  string
		msg   string
	}{
		{NONE, OrderNone, "none ordering"},
		{UNORDERED, OrderUnordered, "unordered"},
		{ORDERED, OrderOrdered, "ordered"},
	}

	for _, tt := range cases {
		tt := tt
		require.Equal(t, tt.order, OrderFromString(tt.name), tt.msg)
		require.Equal(t, tt.name, tt.order.String(), tt.msg)
	}
}

func TestOrderMarshalJSON(t *testing.T) {
	cases := []struct {
		order      Order
		name       string
		msg        string
		expectPass bool
	}{
		{NONE, OrderNone, "none ordering should have failed", false},
		{UNORDERED, OrderUnordered, "unordered should have passed", true},
		{ORDERED, OrderOrdered, "ordered should have passed", true},
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
