package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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

	proof         = commitment.Proof{Proof: &merkle.Proof{}}
	proofs        = []commitment.Proof{proof}
	invalidProofs = []commitment.Proof{}

	addr1     = sdk.AccAddress("testaddr1")
	addr2     = sdk.AccAddress("testaddr2")
	emptyAddr sdk.AccAddress

	coins, _          = sdk.ParseCoins("100atom")
	invalidDenomCoins = sdk.Coins{sdk.Coin{Denom: "ato-m", Amount: sdk.NewInt(100)}}
	negativeCoins     = sdk.Coins{sdk.Coin{Denom: "atom", Amount: sdk.NewInt(-100)}}
)

// TestMsgTransfer tests ValidateBasic for MsgTransfer
func TestMsgTransfer(t *testing.T) {
	testMsgs := []MsgTransfer{
		NewMsgTransfer("testportid", "testchannel", coins, addr1, addr2, true),              // valid msg
		NewMsgTransfer(invalidShortPort, "testchannel", coins, addr1, addr2, true),          // too short port id
		NewMsgTransfer(invalidLongPort, "testchannel", coins, addr1, addr2, true),           // too long port id
		NewMsgTransfer(invalidPort, "testchannel", coins, addr1, addr2, true),               // port id contains non-alpha
		NewMsgTransfer("testportid", invalidShortChannel, coins, addr1, addr2, true),        // too short channel id
		NewMsgTransfer("testportid", invalidLongChannel, coins, addr1, addr2, false),        // too long channel id
		NewMsgTransfer("testportid", invalidChannel, coins, addr1, addr2, false),            // channel id contains non-alpha
		NewMsgTransfer("testportid", "testchannel", invalidDenomCoins, addr1, addr2, false), // invalid amount
		NewMsgTransfer("testportid", "testchannel", negativeCoins, addr1, addr2, false),     // negative amount
		NewMsgTransfer("testportid", "testchannel", coins, emptyAddr, addr2, false),         // missing sender address
		NewMsgTransfer("testportid", "testchannel", coins, addr1, emptyAddr, false),         // missing recipient address
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
		{testMsgs[8], false, "negative amount"},
		{testMsgs[9], false, "missing sender address"},
		{testMsgs[10], false, "missing recipient address"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.Nil(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.NotNil(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgRecvPacket tests ValidateBasic for MsgRecvPacket
func TestMsgRecvPacket(t *testing.T) {
	testMsgs := []MsgRecvPacket{
		NewMsgRecvPacket(packet, proofs, 1, addr1),        // valid msg
		NewMsgRecvPacket(packet, proofs, 0, addr1),        // proof height is zero
		NewMsgRecvPacket(packet, nil, 1, addr1),           // missing proofs
		NewMsgRecvPacket(packet, invalidProofs, 1, addr1), // missing proofs
		NewMsgRecvPacket(packet, proofs, 1, emptyAddr),    // missing signer address
		NewMsgRecvPacket(invalidPacket, proofs, 1, addr1), // invalid packet
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
		{testMsgs[4], false, "missing signer address"},
		{testMsgs[5], false, "invalid packet"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.Nil(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.NotNil(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}
