package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"

	"github.com/stretchr/testify/require"
)

// define constants used for testing
const (
	invalidPort      = "invalidport1"
	invalidShortPort = "p"
	invalidLongPort  = "invalidlongportinvalidlongport"

	invalidChannel      = "invalidchannel1"
	invalidShortChannel = "invalidch"
	invalidLongChannel  = "invalidlongchannelinvalidlongchannel"

	invalidConnection      = "invalidconnection1"
	invalidShortConnection = "invalidcn"
	invalidLongConnection  = "invalidlongconnection"
)

// define variables used for testing
var (
	connHops             = []string{"testconnection"}
	invalidConnHops      = []string{"testconnection", "testconnection"}
	invalidShortConnHops = []string{invalidShortConnection}
	invalidLongConnHops  = []string{invalidLongConnection}

	proof = commitment.Proof{}

	addr = sdk.AccAddress("test1")
)

// TestMsgChannelOpenInit tests ValidateBasic for MsgChannelOpenInit
func TestMsgChannelOpenInit(t *testing.T) {
	tests := []struct {
		name                  string
		portID                string
		channelID             string
		version               string
		channelOrder          Order
		connectionHops        []string
		counterpartyPortID    string
		counterpartyChannelID string
		signer                sdk.AccAddress
		expectPass            bool
	}{
		{"basic good", "testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr, true},
		{"too short port", invalidShortPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr, false},
		{"too long port", invalidLongPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr, false},
		{"port contains non-alpha", invalidPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr, false},
		{"too short channel", "testport", invalidShortChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr, false},
		{"too long channel", "testport", invalidLongChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr, false},
		{"channel contains non-alpha", "testport", invalidChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr, false},
		{"invalid order", "testport", "testchannel", "1.0", Order(3), connHops, "testcpport", "testcpchannel", addr, false},
		{"connection hops more than 1", "testport", "testchannel", "1.0", UNORDERED, invalidConnHops, "testcpport", "testcpchannel", addr, false},
		{"too short connection", "testport", "testchannel", "1.0", UNORDERED, invalidShortConnHops, "testcpport", "testcpchannel", addr, false},
		{"too long connection", "testport", "testchannel", "1.0", UNORDERED, invalidLongConnHops, "testcpport", "testcpchannel", addr, false},
		{"connection contains non-alpha", "testport", "testchannel", "1.0", UNORDERED, []string{invalidConnection}, "testcpport", "testcpchannel", addr, false},
		{"empty version", "testport", "testchannel", "", UNORDERED, connHops, "testcpport", "testcpchannel", addr, false},
		{"invalid counterparty port", "testport", "testchannel", "1.0", UNORDERED, connHops, invalidPort, "testcpchannel", addr, false},
		{"invalid counterparty channel", "testport", "testchannel", "1.0", UNORDERED, connHops, "testcpport", invalidChannel, addr, false},
	}

	for _, tc := range tests {
		msg := NewMsgChannelOpenInit(tc.portID, tc.channelID, tc.version, tc.channelOrder, tc.connectionHops, tc.counterpartyPortID, tc.counterpartyChannelID, tc.signer)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// TestMsgChannelOpenTry tests ValidateBasic for MsgChannelOpenTry
func TestMsgChannelOpenTry(t *testing.T) {
	tests := []struct {
		name                  string
		portID                string
		channelID             string
		version               string
		channelOrder          Order
		connectionHops        []string
		counterpartyPortID    string
		counterpartyChannelID string
		counterpartyVersion   string
		proofInit             commitment.ProofI
		proofHeight           uint64
		signer                sdk.AccAddress
		expectPass            bool
	}{
		{"basic good", "testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, true},
		{"too short port", invalidShortPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"too long port", invalidLongPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"port contains non-alpha", invalidPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"too short channel", "testport", invalidShortChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"too long channel", "testport", invalidLongChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"channel contains non-alpha", "testport", invalidChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"empty counterparty version", "testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "", proof, 1, addr, false},
		{"empty proof", "testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", nil, 1, addr, false},
		{"proof height is zero", "testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 0, addr, false},
		{"invalid order", "testport", "testchannel", "1.0", Order(4), connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"connection hops more than 1", "testport", "testchannel", "1.0", UNORDERED, invalidConnHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"too short connection", "testport", "testchannel", "1.0", UNORDERED, invalidShortConnHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"too long connection", "testport", "testchannel", "1.0", UNORDERED, invalidLongConnHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"connection contains non-alpha", "testport", "testchannel", "1.0", UNORDERED, []string{invalidConnection}, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"empty version", "testport", "testchannel", "", UNORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr, false},
		{"invalid counterparty port", "testport", "testchannel", "1.0", UNORDERED, connHops, invalidPort, "testcpchannel", "1.0", proof, 1, addr, false},
		{"invalid counterparty channel", "testport", "testchannel", "1.0", UNORDERED, connHops, "testcpport", invalidChannel, "1.0", proof, 1, addr, false},
	}

	for _, tc := range tests {
		msg := NewMsgChannelOpenTry(tc.portID, tc.channelID, tc.version, tc.channelOrder, tc.connectionHops, tc.counterpartyPortID, tc.counterpartyChannelID, tc.counterpartyVersion, tc.proofInit, tc.proofHeight, tc.signer)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// TestMsgChannelOpenAck tests ValidateBasic for MsgChannelOpenAck
func TestMsgChannelOpenAck(t *testing.T) {
	tests := []struct {
		name                string
		portID              string
		channelID           string
		counterpartyVersion string
		proofTry            commitment.ProofI
		proofHeight         uint64
		signer              sdk.AccAddress
		expectPass          bool
	}{
		{"basic good", "testport", "testchannel", "1.0", proof, 1, addr, true},
		{"too short port", invalidShortPort, "testchannel", "1.0", proof, 1, addr, false},
		{"too long port", invalidLongPort, "testchannel", "1.0", proof, 1, addr, false},
		{"port contains non-alpha", invalidPort, "testchannel", "1.0", proof, 1, addr, false},
		{"too short channel", "testport", invalidShortChannel, "1.0", proof, 1, addr, false},
		{"too long channel", "testport", invalidLongChannel, "1.0", proof, 1, addr, false},
		{"channel contains non-alpha", "testport", invalidChannel, "1.0", proof, 1, addr, false},
		{"empty counterparty version", "testport", "testchannel", "", proof, 1, addr, false},
		{"empty proof", "testport", "testchannel", "1.0", nil, 1, addr, false},
		{"proof height is zero", "testport", "testchannel", "1.0", proof, 0, addr, false},
	}

	for _, tc := range tests {
		msg := NewMsgChannelOpenAck(tc.portID, tc.channelID, tc.counterpartyVersion, tc.proofTry, tc.proofHeight, tc.signer)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// TestMsgChannelOpenConfirm tests ValidateBasic for MsgChannelOpenConfirm
func TestMsgChannelOpenConfirm(t *testing.T) {
	tests := []struct {
		name        string
		portID      string
		channelID   string
		proofAck    commitment.ProofI
		proofHeight uint64
		signer      sdk.AccAddress
		expectPass  bool
	}{
		{"basic good", "testport", "testchannel", proof, 1, addr, true},
		{"too short port", invalidShortPort, "testchannel", proof, 1, addr, false},
		{"too long port", invalidLongPort, "testchannel", proof, 1, addr, false},
		{"port contains non-alpha", invalidPort, "testchannel", proof, 1, addr, false},
		{"too short channel", "testport", invalidShortChannel, proof, 1, addr, false},
		{"too long channel", "testport", invalidLongChannel, proof, 1, addr, false},
		{"channel contains non-alpha", "testport", invalidChannel, proof, 1, addr, false},
		{"empty proof", "testport", "testchannel", nil, 1, addr, false},
		{"proof height is zero", "testport", "testchannel", proof, 0, addr, false},
	}

	for _, tc := range tests {
		msg := NewMsgChannelOpenConfirm(tc.portID, tc.channelID, tc.proofAck, tc.proofHeight, tc.signer)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// TestMsgChannelCloseInit tests ValidateBasic for MsgChannelCloseInit
func TestMsgChannelCloseInit(t *testing.T) {
	tests := []struct {
		name       string
		portID     string
		channelID  string
		signer     sdk.AccAddress
		expectPass bool
	}{
		{"basic good", "testport", "testchannel", addr, true},
		{"too short port", invalidShortPort, "testchannel", addr, false},
		{"too long port", invalidLongPort, "testchannel", addr, false},
		{"port contains non-alpha", invalidPort, "testchannel", addr, false},
		{"too short channel", "testport", invalidShortChannel, addr, false},
		{"too long channel", "testport", invalidLongChannel, addr, false},
		{"channel contains non-alpha", "testport", invalidChannel, addr, false},
	}

	for _, tc := range tests {
		msg := NewMsgChannelCloseInit(tc.portID, tc.channelID, tc.signer)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// TestMsgChannelCloseConfirm tests ValidateBasic for MsgChannelCloseConfirm
func TestMsgChannelCloseConfirm(t *testing.T) {
	tests := []struct {
		name        string
		portID      string
		channelID   string
		proofInit   commitment.ProofI
		proofHeight uint64
		signer      sdk.AccAddress
		expectPass  bool
	}{
		{"basic good", "testport", "testchannel", proof, 1, addr, true},
		{"too short port", invalidShortPort, "testchannel", proof, 1, addr, false},
		{"too long port", invalidLongPort, "testchannel", proof, 1, addr, false},
		{"port contains non-alpha", invalidPort, "testchannel", proof, 1, addr, false},
		{"too short channel", "testport", invalidShortChannel, proof, 1, addr, false},
		{"too long channel", "testport", invalidLongChannel, proof, 1, addr, false},
		{"channel contains non-alpha", "testport", invalidChannel, proof, 1, addr, false},
		{"empty proof", "testport", "testchannel", nil, 1, addr, false},
		{"proof height is zero", "testport", "testchannel", proof, 0, addr, false},
	}

	for _, tc := range tests {
		msg := NewMsgChannelCloseConfirm(tc.portID, tc.channelID, tc.proofInit, tc.proofHeight, tc.signer)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}
