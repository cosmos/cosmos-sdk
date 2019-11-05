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

func TestNewMsgConnectionOpenTry(t *testing.T) {
	type TestCase = struct {
		connectionID         string
		clientID             string
		counterparty         Counterparty
		counterpartyVersions []string
		proofInit            commitment.ProofI
		proofHeight          uint64
		consensusHeight      uint64
		signer               sdk.AccAddress
		expected             bool
		msg                  string
	}

	prefix := commitment.NewPrefix([]byte("storePrefixKey"))
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	var testCases = []TestCase{
		{
			connectionID:         "gaia/conn1",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", prefix),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      10,
			signer:               signer,
			expected:             false,
			msg:                  "invalid connection ID",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaia/iris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", prefix),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      10,
			signer:               signer,
			expected:             false,
			msg:                  "invalid client ID",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("gaia/conn1", "ibcconniris", prefix),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      10,
			signer:               signer,
			expected:             false,
			msg:                  "invalid counterparty client ID",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibc/gaia", prefix),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      10,
			signer:               signer,
			expected:             false,
			msg:                  "invalid counterparty connection ID",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", nil),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      10,
			signer:               signer,
			expected:             false,
			msg:                  "empty counterparty prefix",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", nil),
			counterpartyVersions: []string{},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      10,
			signer:               signer,
			expected:             false,
			msg:                  "empty counterpartyVersions",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", nil),
			counterpartyVersions: []string{""},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      10,
			signer:               signer,
			expected:             false,
			msg:                  "empty Versions",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", nil),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            nil,
			proofHeight:          10,
			consensusHeight:      10,
			signer:               signer,
			expected:             false,
			msg:                  "empty proof",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", nil),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            commitment.Proof{},
			proofHeight:          0,
			consensusHeight:      10,
			signer:               signer,
			expected:             false,
			msg:                  "invalid proofHeight",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", nil),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      0,
			signer:               signer,
			expected:             false,
			msg:                  "invalid consensusHeight",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", prefix),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      10,
			signer:               nil,
			expected:             false,
			msg:                  "empty singer",
		},
		{
			connectionID:         "ibcconngaia",
			clientID:             "gaiatoiris",
			counterparty:         NewCounterparty("iristogaia", "ibcconniris", prefix),
			counterpartyVersions: []string{"1.0.0"},
			proofInit:            commitment.Proof{},
			proofHeight:          10,
			consensusHeight:      10,
			signer:               signer,
			expected:             true,
			msg:                  "success",
		},
	}

	for i, tc := range testCases {
		msg := NewMsgConnectionOpenTry(tc.connectionID,
			tc.clientID, tc.counterparty.ConnectionID, tc.counterparty.ClientID, tc.counterparty.Prefix, tc.counterpartyVersions, tc.proofInit, tc.proofHeight, tc.consensusHeight, tc.signer)
		require.Equal(t, tc.expected, msg.ValidateBasic() == nil, fmt.Sprintf("case: %d,msg: %s,", i, tc.msg))
	}
}

func TestNewMsgConnectionOpenAck(t *testing.T) {
	type TestCase = struct {
		connectionID    string
		proofTry        commitment.ProofI
		proofHeight     uint64
		consensusHeight uint64
		version         string
		signer          sdk.AccAddress
		expected        bool
		msg             string
	}

	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	var testCases = []TestCase{
		{
			connectionID:    "gaia/conn1",
			proofTry:        commitment.Proof{},
			proofHeight:     10,
			consensusHeight: 10,
			version:         "1.0.0",
			signer:          signer,
			expected:        false,
			msg:             "invalid connection ID",
		},
		{
			connectionID:    "ibcconngaia",
			proofTry:        nil,
			proofHeight:     10,
			consensusHeight: 10,
			version:         "1.0.0",
			signer:          signer,
			expected:        false,
			msg:             "empty proofTry",
		},
		{
			connectionID:    "ibcconngaia",
			proofTry:        commitment.Proof{},
			proofHeight:     0,
			consensusHeight: 10,
			version:         "1.0.0",
			signer:          signer,
			expected:        false,
			msg:             "invalid proofHeight",
		},
		{
			connectionID:    "ibcconngaia",
			proofTry:        commitment.Proof{},
			proofHeight:     10,
			consensusHeight: 0,
			version:         "1.0.0",
			signer:          signer,
			expected:        false,
			msg:             "invalid consensusHeight",
		},
		{
			connectionID:    "ibcconngaia",
			proofTry:        commitment.Proof{},
			proofHeight:     10,
			consensusHeight: 10,
			version:         "",
			signer:          signer,
			expected:        false,
			msg:             "invalid version",
		},
		{
			connectionID:    "ibcconngaia",
			proofTry:        commitment.Proof{},
			proofHeight:     10,
			consensusHeight: 10,
			version:         "1.0.0",
			signer:          nil,
			expected:        false,
			msg:             "empty signer",
		},
		{
			connectionID:    "ibcconngaia",
			proofTry:        commitment.Proof{},
			proofHeight:     10,
			consensusHeight: 10,
			version:         "1.0.0",
			signer:          signer,
			expected:        true,
			msg:             "success",
		},
	}

	for i, tc := range testCases {
		msg := NewMsgConnectionOpenAck(tc.connectionID,
			tc.proofTry, tc.proofHeight, tc.consensusHeight, tc.version, tc.signer)
		require.Equal(t, tc.expected, msg.ValidateBasic() == nil, fmt.Sprintf("case: %d,msg: %s,", i, tc.msg))
	}
}

func TestNewMsgConnectionOpenConfirm(t *testing.T) {
	type TestCase = struct {
		connectionID string
		proofAck     commitment.ProofI
		proofHeight  uint64
		signer       sdk.AccAddress
		expected     bool
		msg          string
	}

	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	var testCases = []TestCase{
		{
			connectionID: "gaia/conn1",
			proofAck:     commitment.Proof{},
			proofHeight:  10,
			signer:       signer,
			expected:     false,
			msg:          "invalid connection ID",
		},
		{
			connectionID: "ibcconngaia",
			proofAck:     nil,
			proofHeight:  10,
			signer:       signer,
			expected:     false,
			msg:          "empty proofTry",
		},
		{
			connectionID: "ibcconngaia",
			proofAck:     commitment.Proof{},
			proofHeight:  0,
			signer:       signer,
			expected:     false,
			msg:          "invalid proofHeight",
		},
		{
			connectionID: "ibcconngaia",
			proofAck:     commitment.Proof{},
			proofHeight:  10,
			signer:       nil,
			expected:     false,
			msg:          "empty signer",
		},
		{
			connectionID: "ibcconngaia",
			proofAck:     commitment.Proof{},
			proofHeight:  10,
			signer:       signer,
			expected:     true,
			msg:          "success",
		},
	}

	for i, tc := range testCases {
		msg := NewMsgConnectionOpenConfirm(tc.connectionID,
			tc.proofAck, tc.proofHeight, tc.signer)
		require.Equal(t, tc.expected, msg.ValidateBasic() == nil, fmt.Sprintf("case: %d,msg: %s,", i, tc.msg))
	}
}
