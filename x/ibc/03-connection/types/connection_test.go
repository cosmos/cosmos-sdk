package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

var (
	connectionID  = "connectionidone"
	clientID      = "clientidone"
	connectionID2 = "connectionidtwo"
	clientID2     = "clientidtwo"
)

func TestConnectionValidateBasic(t *testing.T) {
	prefixMerkleAny, err := commitmenttypes.NewMerklePrefix([]byte("prefix")).PackAny()
	require.NoError(t, err)

	testCases := []struct {
		name       string
		connection ConnectionEnd
		expPass    bool
	}{
		{
			"valid connection",
			ConnectionEnd{connectionID, clientID, []string{"1.0.0"}, INIT, Counterparty{clientID2, connectionID2, *prefixMerkleAny}},
			true,
		},
		{
			"invalid connection id",
			ConnectionEnd{"(connectionIDONE)", clientID, []string{"1.0.0"}, INIT, Counterparty{clientID2, connectionID2, *prefixMerkleAny}},
			false,
		},
		{
			"invalid client id",
			ConnectionEnd{connectionID, "(clientID1)", []string{"1.0.0"}, INIT, Counterparty{clientID2, connectionID2, *prefixMerkleAny}},
			false,
		},
		{
			"empty versions",
			ConnectionEnd{connectionID, clientID, nil, INIT, Counterparty{clientID2, connectionID2, *prefixMerkleAny}},
			false,
		},
		{
			"invalid version",
			ConnectionEnd{connectionID, clientID, []string{""}, INIT, Counterparty{clientID2, connectionID2, *prefixMerkleAny}},
			false,
		},
		{
			"invalid counterparty",
			ConnectionEnd{connectionID, clientID, []string{"1.0.0"}, INIT, Counterparty{clientID2, connectionID2, cdctypes.Any{}}},
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
	prefixMerkleAny, err := commitmenttypes.NewMerklePrefix([]byte("prefix")).PackAny()
	require.NoError(t, err)

	prefixSigAny, err := commitmenttypes.NewSignaturePrefix([]byte("prefix")).PackAny()
	require.NoError(t, err)

	testCases := []struct {
		name         string
		counterparty Counterparty
		expPass      bool
	}{
		{"valid merkle counterparty", Counterparty{clientID, connectionID2, *prefixMerkleAny}, true},
		{"valid sig counterparty", Counterparty{clientID, connectionID2, *prefixSigAny}, true},
		{"invalid client id", Counterparty{"(InvalidClient)", connectionID2, *prefixMerkleAny}, false},
		{"invalid connection id", Counterparty{clientID, "(InvalidConnection)", *prefixMerkleAny}, false},
		{"invalid prefix", Counterparty{clientID, connectionID2, cdctypes.Any{}}, false},
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
