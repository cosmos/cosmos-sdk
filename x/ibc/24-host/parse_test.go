package host

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseClientPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expClientID string
		expErr      bool
	}{
		{"client path", "clients/tendermintclient", "tendermintclient", false},
		{"consensus path", FullClientPath("tendermintclient", ConsensusStatePath(1)), "tendermintclient", false},
		{"connection path", ConnectionPath("connection1"), "", true},
		{"invalid path", "clients", "", true},
		{"invalid client id", "clients/;'$#!@#™", "", true},
	}

	for _, tc := range tests {
		clientID, err := ParseClientPath(tc.path)
		if tc.expErr {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expClientID, clientID, tc.name)
		}
	}
}
