package exported

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientTypeString(t *testing.T) {
	cases := []struct {
		clientType ClientType
		name       string
		msg        string
	}{
		{Tendermint, ClientTypeTendermint, "tendermint client"},
		{0, "", "empty type"},
	}

	for _, tt := range cases {
		tt := tt
		require.Equal(t, tt.clientType, ClientTypeFromString(tt.name), tt.msg)
		require.Equal(t, tt.name, tt.clientType.String(), tt.msg)
	}
}

func TestClientTypeMarshalJSON(t *testing.T) {
	cases := []struct {
		clientType ClientType
		name       string
		msg        string
		expectPass bool
	}{
		{Tendermint, ClientTypeTendermint, "tendermint client should have passed", true},
		{0, "", "empty type should have failed", false},
	}

	for _, tt := range cases {
		tt := tt
		bz, err := tt.clientType.MarshalJSON()
		require.NoError(t, err)
		var ct ClientType
		if tt.expectPass {
			require.NoError(t, ct.UnmarshalJSON(bz), tt.msg)
			require.Equal(t, tt.name, ct.String(), tt.msg)
		} else {
			require.Error(t, ct.UnmarshalJSON(bz), tt.msg)
		}
	}
}
