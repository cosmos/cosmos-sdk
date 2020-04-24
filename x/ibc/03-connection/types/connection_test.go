package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

var (
	connectionID  = "connectionidone"
	clientID      = "clientidone"
	connectionID2 = "connectionidtwo"
	clientID2     = "clientidtwo"
)

func TestConnectionValidateBasic(t *testing.T) {
	testCases := []struct {
		name       string
		connection ConnectionEnd
		expPass    bool
	}{
		{
			"valid connection",
			ConnectionEnd{exported.INIT, connectionID, clientID, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{"1.0.0"}},
			true,
		},
		{
			"invalid connection id",
			ConnectionEnd{exported.INIT, "connectionIDONE", clientID, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{"1.0.0"}},
			false,
		},
		{
			"invalid client id",
			ConnectionEnd{exported.INIT, connectionID, "ClientIDTwo", Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{"1.0.0"}},
			false,
		},
		{
			"empty versions",
			ConnectionEnd{exported.INIT, connectionID, clientID, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, nil},
			false,
		},
		{
			"invalid version",
			ConnectionEnd{exported.INIT, connectionID, clientID, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{""}},
			false,
		},
		{
			"invalid counterparty",
			ConnectionEnd{exported.INIT, connectionID, clientID, Counterparty{clientID2, connectionID2, nil}, []string{"1.0.0"}},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.connection.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func TestCounterpartyValidateBasic(t *testing.T) {
	testCases := []struct {
		name         string
		counterparty Counterparty
		expPass      bool
	}{
		{"valid counterparty", Counterparty{"clientidone", connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, true},
		{"invalid client id", Counterparty{"InvalidClient", "channelidone", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, false},
		{"invalid connection id", Counterparty{"clientidone", "InvalidConnection", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, false},
		{"invalid prefix", Counterparty{"clientidone", connectionID2, nil}, false},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.counterparty.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
