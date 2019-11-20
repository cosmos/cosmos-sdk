package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/merkle"
)

// define constants used for testing
const (
	invalidPort      = "invalidport1"
	invalidShortPort = "p"
	invalidLongPort  = "invalidlongportinvalidlongport"

	invalidChannel      = "invalidchannel1"
	invalidShortChannel = "invalidch"
	invalidLongChannel  = "invalidlongchannelinvalidlongchannel"
)

// define variables used for testing
var (
	packet        = channel.NewPacket(1, 100, "testportid", "testchannel", "testcpport", "testcpchannel", []byte("testdata"))
	invalidPacket = channel.NewPacket(0, 100, "testportid", "testchannel", "testcpport", "testcpchannel", []byte{})

	proof          = commitment.Proof{Proof: &merkle.Proof{}}
	emptyProof     = commitment.Proof{Proof: nil}
	proofs         = []commitment.Proof{proof}
	invalidProofs1 = []commitment.Proof{}
	invalidProofs2 = []commitment.Proof{emptyProof}

	addr1     = sdk.AccAddress("testaddr1")
	addr2     = sdk.AccAddress("testaddr2")
	emptyAddr sdk.AccAddress

	coins, _          = sdk.ParseCoins("100atom")
	invalidDenomCoins = sdk.Coins{sdk.Coin{Denom: "ato-m", Amount: sdk.NewInt(100)}}
	negativeCoins     = sdk.Coins{sdk.Coin{Denom: "atom", Amount: sdk.NewInt(100)}, sdk.Coin{Denom: "atoms", Amount: sdk.NewInt(-100)}}
)

// TestMsgTransferRoute tests Route for MsgTransfer
func TestMsgTransferRoute(t *testing.T) {
	msg := NewMsgTransfer("testportid", "testchannel", coins, addr1, addr2, true)

	require.Equal(t, ibctypes.RouterKey, msg.Route())
}

// TestMsgTransferType tests Type for MsgTransfer
func TestMsgTransferType(t *testing.T) {
	msg := NewMsgTransfer("testportid", "testchannel", coins, addr1, addr2, true)

	require.Equal(t, "transfer", msg.Type())
}

// TestMsgTransferValidation tests ValidateBasic for MsgTransfer
func TestMsgTransferValidation(t *testing.T) {
	testMsgs := []MsgTransfer{
		NewMsgTransfer("testportid", "testchannel", coins, addr1, addr2, true),              // valid msg
		NewMsgTransfer(invalidShortPort, "testchannel", coins, addr1, addr2, true),          // too short port id
		NewMsgTransfer(invalidLongPort, "testchannel", coins, addr1, addr2, true),           // too long port id
		NewMsgTransfer(invalidPort, "testchannel", coins, addr1, addr2, true),               // port id contains non-alpha
		NewMsgTransfer("testportid", invalidShortChannel, coins, addr1, addr2, true),        // too short channel id
		NewMsgTransfer("testportid", invalidLongChannel, coins, addr1, addr2, false),        // too long channel id
		NewMsgTransfer("testportid", invalidChannel, coins, addr1, addr2, false),            // channel id contains non-alpha
		NewMsgTransfer("testportid", "testchannel", invalidDenomCoins, addr1, addr2, false), // invalid amount
		NewMsgTransfer("testportid", "testchannel", negativeCoins, addr1, addr2, false),     // amount contains negative coin
		NewMsgTransfer("testportid", "testchannel", coins, emptyAddr, addr2, false),         // missing sender address
		NewMsgTransfer("testportid", "testchannel", coins, addr1, emptyAddr, false),         // missing recipient address
		NewMsgTransfer("testportid", "testchannel", sdk.Coins{}, addr1, addr2, false),       // not possitive coin
	}

	testCases := []struct {
		msg     MsgTransfer
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], true, ""},
		{testMsgs[1], false, "too short port id"},
		{testMsgs[2], false, "too long port id"},
		{testMsgs[3], false, "port id contains non-alpha"},
		{testMsgs[4], false, "too short channel id"},
		{testMsgs[5], false, "too long channel id"},
		{testMsgs[6], false, "channel id contains non-alpha"},
		{testMsgs[7], false, "invalid amount"},
		{testMsgs[8], false, "amount contains negative coin"},
		{testMsgs[9], false, "missing sender address"},
		{testMsgs[10], false, "missing recipient address"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.Error(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgTransferGetSignBytes tests GetSignBytes for MsgTransfer
func TestMsgTransferGetSignBytes(t *testing.T) {
	msg := NewMsgTransfer("testportid", "testchannel", coins, addr1, addr2, true)
	res := msg.GetSignBytes()

	expected := `{"type":"ibc/transfer/MsgTransfer","value":{"amount":[{"amount":"100","denom":"atom"}],"receiver":"cosmos1w3jhxarpv3j8yvs7f9y7g","sender":"cosmos1w3jhxarpv3j8yvg4ufs4x","source":true,"source_channel":"testchannel","source_port":"testportid"}}`
	require.Equal(t, expected, string(res))
}

// TestMsgTransferGetSigners tests GetSigners for MsgTransfer
func TestMsgTransferGetSigners(t *testing.T) {
	msg := NewMsgTransfer("testportid", "testchannel", coins, addr1, addr2, true)
	res := msg.GetSigners()

	expected := "[746573746164647231]"
	require.Equal(t, expected, fmt.Sprintf("%v", res))
}

// TestMsgRecvPacketRoute tests Route for MsgRecvPacket
func TestMsgRecvPacketRoute(t *testing.T) {
	msg := NewMsgRecvPacket(packet, proofs, 1, addr1)

	require.Equal(t, ibctypes.RouterKey, msg.Route())
}

// TestMsgRecvPacketType tests Type for MsgRecvPacket
func TestMsgRecvPacketType(t *testing.T) {
	msg := NewMsgRecvPacket(packet, proofs, 1, addr1)

	require.Equal(t, "recv_packet", msg.Type())
}

// TestMsgRecvPacketValidation tests ValidateBasic for MsgRecvPacket
func TestMsgRecvPacketValidation(t *testing.T) {
	testMsgs := []MsgRecvPacket{
		NewMsgRecvPacket(packet, proofs, 1, addr1),         // valid msg
		NewMsgRecvPacket(packet, proofs, 0, addr1),         // proof height is zero
		NewMsgRecvPacket(packet, nil, 1, addr1),            // missing proofs
		NewMsgRecvPacket(packet, invalidProofs1, 1, addr1), // missing proofs
		NewMsgRecvPacket(packet, invalidProofs2, 1, addr1), // proofs contain empty proof
		NewMsgRecvPacket(packet, proofs, 1, emptyAddr),     // missing signer address
		NewMsgRecvPacket(invalidPacket, proofs, 1, addr1),  // invalid packet
	}

	testCases := []struct {
		msg     MsgRecvPacket
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], true, ""},
		{testMsgs[1], false, "proof height is zero"},
		{testMsgs[2], false, "missing proofs"},
		{testMsgs[3], false, "missing proofs"},
		{testMsgs[4], false, "proofs contain empty proof"},
		{testMsgs[5], false, "missing signer address"},
		{testMsgs[6], false, "invalid packet"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.Nil(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgRecvPacketGetSignBytes tests GetSignBytes for MsgRecvPacket
func TestMsgRecvPacketGetSignBytes(t *testing.T) {
	msg := NewMsgRecvPacket(packet, proofs, 1, addr1)
	res := msg.GetSignBytes()

	expected := `{"type":"ibc/transfer/MsgRecvPacket","value":{"height":"1","packet":{"type":"ibc/channel/Packet","value":{"data":"dGVzdGRhdGE=","destination_channel":"testcpchannel","destination_port":"testcpport","sequence":"1","source_channel":"testchannel","source_port":"testportid","timeout":"100"}},"proofs":[{"proof":{"ops":[]}}],"signer":"cosmos1w3jhxarpv3j8yvg4ufs4x"}}`
	require.Equal(t, expected, string(res))
}

// TestMsgRecvPacketGetSigners tests GetSigners for MsgRecvPacket
func TestMsgRecvPacketGetSigners(t *testing.T) {
	msg := NewMsgRecvPacket(packet, proofs, 1, addr1)
	res := msg.GetSigners()

	expected := "[746573746164647231]"
	require.Equal(t, expected, fmt.Sprintf("%v", res))
}
