package exported

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientTypeString(t *testing.T) {
	cases := []struct {
		msg        string
		name       string
		clientType ClientType
	}{
		{"tendermint client", ClientTypeTendermint, Tendermint},
		{"empty type", "", 0},
	}

	for _, tt := range cases {
		tt := tt
		require.Equal(t, tt.clientType, ClientTypeFromString(tt.name), tt.msg)
		require.Equal(t, tt.name, tt.clientType.String(), tt.msg)
	}
}

func TestClientTypeMarshalJSON(t *testing.T) {
	cases := []struct {
		msg        string
		name       string
		clientType ClientType
		expectPass bool
	}{
		{"tendermint client should have passed", ClientTypeTendermint, Tendermint, true},
		{"empty type should have failed", "", 0, false},
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
