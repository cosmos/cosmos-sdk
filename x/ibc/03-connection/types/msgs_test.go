package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/stretchr/testify/require"
)

func TestNewMsgConnectionOpenInit(t *testing.T) {
	type TestCase = struct {
		connectionID string
		clientID     string
		counterparty Counterparty
		signer       sdk.AccAddress
		expected     bool
		msg          string
	}

	prefix := commitment.NewPrefix([]byte("storePrefixKey"))
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	var testCases = []TestCase{
		{
			connectionID: "gaia/conn1",
			clientID:     "gaiatoiris",
			counterparty: NewCounterparty("iristogaia", "ibcconniris", prefix),
			signer:       signer,
			expected:     false,
			msg:          "invalid connection ID",
		},
		{
			connectionID: "ibcconngaia",
			clientID:     "gaia/iris",
			counterparty: NewCounterparty("iristogaia", "ibcconniris", prefix),
			signer:       signer,
			expected:     false,
			msg:          "invalid client ID",
		},
		{
			connectionID: "ibcconngaia",
			clientID:     "gaiatoiris",
			counterparty: NewCounterparty("gaia/conn1", "ibcconniris", prefix),
			signer:       signer,
			expected:     false,
			msg:          "invalid counterparty client ID",
		},
		{
			connectionID: "ibcconngaia",
			clientID:     "gaiatoiris",
			counterparty: NewCounterparty("iristogaia", "ibc/gaia", prefix),
			signer:       signer,
			expected:     false,
			msg:          "invalid counterparty connection ID",
		},
		{
			connectionID: "ibcconngaia",
			clientID:     "gaiatoiris",
			counterparty: NewCounterparty("iristogaia", "ibcconniris", nil),
			signer:       signer,
			expected:     false,
			msg:          "empty counterparty prefix",
		},
		{
			connectionID: "ibcconngaia",
			clientID:     "gaiatoiris",
			counterparty: NewCounterparty("iristogaia", "ibcconniris", prefix),
			signer:       nil,
			expected:     false,
			msg:          "empty singer",
		},
		{
			connectionID: "ibcconngaia",
			clientID:     "gaiatoiris",
			counterparty: NewCounterparty("iristogaia", "ibcconniris", prefix),
			signer:       signer,
			expected:     true,
			msg:          "success",
		},
	}

	for i, tc := range testCases {
		msg := NewMsgConnectionOpenInit(tc.connectionID,
			tc.clientID, tc.counterparty.ConnectionID, tc.counterparty.ClientID, tc.counterparty.Prefix, tc.signer)
		require.Equal(t, tc.expected, msg.ValidateBasic() == nil, fmt.Sprintf("case: %d,msg: %s,", i, tc.msg))
	}
}
