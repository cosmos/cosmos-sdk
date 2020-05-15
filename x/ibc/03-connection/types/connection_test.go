package types

import (
	"testing"

	"github.com/stretchr/testify/require"

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
			ConnectionEnd{connectionID, clientID, []string{"1.0.0"}, INIT, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			true,
		},
		{
			"invalid connection id",
			ConnectionEnd{"(connectionIDONE)", clientID, []string{"1.0.0"}, INIT, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			false,
		},
		{
			"invalid client id",
			ConnectionEnd{connectionID, "(clientID1)", []string{"1.0.0"}, INIT, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			false,
		},
		{
			"empty versions",
			ConnectionEnd{connectionID, clientID, nil, INIT, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			false,
		},
		{
			"invalid version",
			ConnectionEnd{connectionID, clientID, []string{""}, INIT, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			false,
		},
		{
			"invalid counterparty",
			ConnectionEnd{connectionID, clientID, []string{"1.0.0"}, INIT, Counterparty{clientID2, connectionID2, emptyPrefix}},
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
		{"valid counterparty", Counterparty{clientID, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, true},
		{"invalid client id", Counterparty{"(InvalidClient)", connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, false},
		{"invalid connection id", Counterparty{clientID, "(InvalidConnection)", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, false},
		{"invalid prefix", Counterparty{clientID, connectionID2, emptyPrefix}, false},
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
