package exported

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionStateString(t *testing.T) {
	cases := []struct {
		name  string
		state State
	}{
		{StateUninitialized, UNINITIALIZED},
		{StateInit, INIT},
		{StateTryOpen, TRYOPEN},
		{StateOpen, OPEN},
	}

	for _, tt := range cases {
		tt := tt
		require.Equal(t, tt.state, StateFromString(tt.name))
		require.Equal(t, tt.name, tt.state.String())
	}
}

func TestConnectionlStateMarshalJSON(t *testing.T) {
	cases := []struct {
		name  string
		state State
	}{
		{StateUninitialized, UNINITIALIZED},
		{StateInit, INIT},
		{StateTryOpen, TRYOPEN},
		{StateOpen, OPEN},
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
