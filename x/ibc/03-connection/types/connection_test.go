package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func TestConnectionValidateBasic(t *testing.T) {
	testCases := []struct {
		name       string
		connection ConnectionEnd
		expPass    bool
	}{
		{
			"valid connection",
			ConnectionEnd{exported.INIT, "clientidone", Counterparty{"clientidtwo", "connectionidone", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{"1.0.0"}},
			true,
		},
		{
			"invalid client id",
			ConnectionEnd{exported.INIT, "ClientIDTwo", Counterparty{"clientidtwo", "connectionidone", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{"1.0.0"}},
			false,
		},
		{
			"empty versions",
			ConnectionEnd{exported.INIT, "clientidone", Counterparty{"clientidtwo", "connectionidone", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, nil},
			false,
		},
		{
			"invalid version",
			ConnectionEnd{exported.INIT, "clientidone", Counterparty{"clientidtwo", "connectionidone", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{""}},
			false,
		},
		{
			"invalid counterparty",
			ConnectionEnd{exported.INIT, "clientidone", Counterparty{"clientidtwo", "connectionidone", nil}, []string{"1.0.0"}},
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
		{"valid counterparty", Counterparty{"clientidone", "connectionidone", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, true},
		{"invalid client id", Counterparty{"InvalidClient", "channelidone", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, false},
		{"invalid connection id", Counterparty{"clientidone", "InvalidConnection", commitmenttypes.NewMerklePrefix([]byte("prefix"))}, false},
		{"invalid prefix", Counterparty{"clientidone", "connectionidone", nil}, false},
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
