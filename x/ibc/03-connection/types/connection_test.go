package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
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
		connection types.ConnectionEnd
		expPass    bool
	}{
		{
			"valid connection",
			types.ConnectionEnd{connectionID, clientID, []string{types.DefaultIBCVersion}, types.INIT, types.Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			true,
		},
		{
			"invalid connection id",
			types.ConnectionEnd{"(connectionIDONE)", clientID, []string{types.DefaultIBCVersion}, types.INIT, types.Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			false,
		},
		{
			"invalid client id",
			types.ConnectionEnd{connectionID, "(clientID1)", []string{types.DefaultIBCVersion}, types.INIT, types.Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			false,
		},
		{
			"empty versions",
			types.ConnectionEnd{connectionID, clientID, nil, types.INIT, types.Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			false,
		},
		{
			"invalid version",
			types.ConnectionEnd{connectionID, clientID, []string{"1.0.0"}, types.INIT, types.Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}},
			false,
		},
		{
			"invalid counterparty",
			types.ConnectionEnd{connectionID, clientID, []string{types.DefaultIBCVersion}, types.INIT, types.Counterparty{clientID2, connectionID2, emptyPrefix}},
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
		counterparty types.Counterparty
		expPass      bool
	}{
		{"valid counterparty", types.Counterparty{clientID, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, true},
		{"invalid client id", types.Counterparty{"(InvalidClient)", connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, false},
		{"invalid connection id", types.Counterparty{clientID, "(InvalidConnection)", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, false},
		{"invalid prefix", types.Counterparty{clientID, connectionID2, emptyPrefix}, false},
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
