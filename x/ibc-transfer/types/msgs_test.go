package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// define constants used for testing
const (
	validPort        = "testportid"
	invalidPort      = "(invalidport1)"
	invalidShortPort = "p"
	invalidLongPort  = "invalidlongportinvalidlongport"

	validChannel        = "testchannel"
	invalidChannel      = "(invalidchannel1)"
	invalidShortChannel = "invalidch"
	invalidLongChannel  = "invalidlongchannelinvalidlongchannel"
)

var (
	addr1     = sdk.AccAddress("testaddr1")
	addr2     = sdk.AccAddress("testaddr2").String()
	emptyAddr sdk.AccAddress

	coins, _          = sdk.ParseCoins("100atom")
	invalidDenomCoins = sdk.Coins{sdk.Coin{Denom: "ato-m", Amount: sdk.NewInt(100)}}
	negativeCoins     = sdk.Coins{sdk.Coin{Denom: "atom", Amount: sdk.NewInt(100)}, sdk.Coin{Denom: "atoms", Amount: sdk.NewInt(-100)}}
)

// TestMsgTransferRoute tests Route for MsgTransfer
func TestMsgTransferRoute(t *testing.T) {
	msg := NewMsgTransfer(validPort, validChannel, 10, coins, addr1, addr2)

	require.Equal(t, RouterKey, msg.Route())
}

// TestMsgTransferType tests Type for MsgTransfer
func TestMsgTransferType(t *testing.T) {
	msg := NewMsgTransfer(validPort, validChannel, 10, coins, addr1, addr2)

	require.Equal(t, "transfer", msg.Type())
}

// TestMsgTransferValidation tests ValidateBasic for MsgTransfer
func TestMsgTransferValidation(t *testing.T) {
	testMsgs := []MsgTransfer{
		NewMsgTransfer(validPort, validChannel, 10, coins, addr1, addr2),             // valid msg
		NewMsgTransfer(invalidShortPort, validChannel, 10, coins, addr1, addr2),      // too short port id
		NewMsgTransfer(invalidLongPort, validChannel, 10, coins, addr1, addr2),       // too long port id
		NewMsgTransfer(invalidPort, validChannel, 10, coins, addr1, addr2),           // port id contains non-alpha
		NewMsgTransfer(validPort, invalidShortChannel, 10, coins, addr1, addr2),      // too short channel id
		NewMsgTransfer(validPort, invalidLongChannel, 10, coins, addr1, addr2),       // too long channel id
		NewMsgTransfer(validPort, invalidChannel, 10, coins, addr1, addr2),           // channel id contains non-alpha
		NewMsgTransfer(validPort, validChannel, 10, invalidDenomCoins, addr1, addr2), // invalid amount
		NewMsgTransfer(validPort, validChannel, 10, negativeCoins, addr1, addr2),     // amount contains negative coin
		NewMsgTransfer(validPort, validChannel, 10, coins, emptyAddr, addr2),         // missing sender address
		NewMsgTransfer(validPort, validChannel, 10, coins, addr1, ""),                // missing recipient address
		NewMsgTransfer(validPort, validChannel, 10, sdk.Coins{}, addr1, addr2),       // not possitive coin
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
	msg := NewMsgTransfer(validPort, validChannel, 10, coins, addr1, addr2)
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/MsgTransfer","value":{"amount":[{"amount":"100","denom":"atom"}],"destination_height":"10","receiver":"cosmos1w3jhxarpv3j8yvs7f9y7g","sender":"cosmos1w3jhxarpv3j8yvg4ufs4x","source_channel":"testchannel","source_port":"testportid"}}`
	require.Equal(t, expected, string(res))
}

// TestMsgTransferGetSigners tests GetSigners for MsgTransfer
func TestMsgTransferGetSigners(t *testing.T) {
	msg := NewMsgTransfer(validPort, validChannel, 10, coins, addr1, addr2)
	res := msg.GetSigners()

	expected := "[746573746164647231]"
	require.Equal(t, expected, fmt.Sprintf("%v", res))
}
