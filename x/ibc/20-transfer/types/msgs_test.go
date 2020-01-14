package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
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

var (
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
