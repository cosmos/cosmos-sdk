package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionStateString(t *testing.T) {
	cases := []struct {
		state State
		name  string
		msg   string
	}{
		{UNINITIALIZED, StateUninitialized, "uninitialized"},
		{INIT, StateInit, "init"},
		{TRYOPEN, StateTryOpen, "tryopen"},
		{OPEN, StateOpen, "open"},
	}

	for _, tt := range cases {
		tt := tt
		require.Equal(t, tt.state, StateFromString(tt.name), tt.msg)
		require.Equal(t, tt.name, tt.state.String(), tt.msg)
	}
}

func TestConnectionlStateMarshalJSON(t *testing.T) {
	cases := []struct {
		state State
		name  string
		msg   string
	}{
		{UNINITIALIZED, StateUninitialized, "uninitialized"},
		{INIT, StateInit, "init"},
		{TRYOPEN, StateTryOpen, "tryopen"},
		{OPEN, StateOpen, "open"},
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
