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

	addr = sdk.AccAddress("testaddr")
)

// TestMsgChannelOpenInit tests ValidateBasic for MsgChannelOpenInit
func TestMsgChannelOpenInit(t *testing.T) {
	testMsgs := []MsgChannelOpenInit{
		NewMsgChannelOpenInit("testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                      // valid msg
		NewMsgChannelOpenInit(invalidShortPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                // too short port id
		NewMsgChannelOpenInit(invalidLongPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                 // too long port id
		NewMsgChannelOpenInit(invalidPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                     // port id contains non-alpha
		NewMsgChannelOpenInit("testport", invalidShortChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                // too short channel id
		NewMsgChannelOpenInit("testport", invalidLongChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                 // too long channel id
		NewMsgChannelOpenInit("testport", invalidChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                     // channel id contains non-alpha
		NewMsgChannelOpenInit("testport", "testchannel", "1.0", Order(3), connHops, "testcpport", "testcpchannel", addr),                     // invalid channel order
		NewMsgChannelOpenInit("testport", "testchannel", "1.0", ORDERED, invalidConnHops, "testcpport", "testcpchannel", addr),               // connection hops more than 1
		NewMsgChannelOpenInit("testport", "testchannel", "1.0", UNORDERED, invalidShortConnHops, "testcpport", "testcpchannel", addr),        // too short connection id
		NewMsgChannelOpenInit("testport", "testchannel", "1.0", UNORDERED, invalidLongConnHops, "testcpport", "testcpchannel", addr),         // too long connection id
		NewMsgChannelOpenInit("testport", "testchannel", "1.0", UNORDERED, []string{invalidConnection}, "testcpport", "testcpchannel", addr), // connection id contains non-alpha
		NewMsgChannelOpenInit("testport", "testchannel", "", UNORDERED, connHops, "testcpport", "testcpchannel", addr),                       // empty channel version
		NewMsgChannelOpenInit("testport", "testchannel", "1.0", UNORDERED, connHops, invalidPort, "testcpchannel", addr),                     // invalid counterparty port id
		NewMsgChannelOpenInit("testport", "testchannel", "1.0", UNORDERED, connHops, "testcpport", invalidChannel, addr),                     // invalid counterparty channel id
	}

	testCases := []struct {
		msg     MsgChannelOpenInit
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
		{testMsgs[7], false, "invalid channel order"},
		{testMsgs[8], false, "connection hops more than 1 "},
		{testMsgs[9], false, "too short connection id"},
		{testMsgs[10], false, "too long connection id"},
		{testMsgs[11], false, "connection id contains non-alpha"},
		{testMsgs[12], false, "empty channel version"},
		{testMsgs[13], false, "invalid counterparty port id"},
		{testMsgs[14], false, "invalid counterparty channel id"},
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

// TestMsgChannelOpenTry tests ValidateBasic for MsgChannelOpenTry
func TestMsgChannelOpenTry(t *testing.T) {
	testMsgs := []MsgChannelOpenTry{
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),                      // valid msg
		NewMsgChannelOpenTry(invalidShortPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),                // too short port id
		NewMsgChannelOpenTry(invalidLongPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),                 // too long port id
		NewMsgChannelOpenTry(invalidPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),                     // port id contains non-alpha
		NewMsgChannelOpenTry("testport", invalidShortChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),                // too short channel id
		NewMsgChannelOpenTry("testport", invalidLongChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),                 // too long channel id
		NewMsgChannelOpenTry("testport", invalidChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),                     // channel id contains non-alpha
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "", proof, 1, addr),                         // empty counterparty version
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", nil, 1, addr),                        // empty proof
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 0, addr),                      // proof height is zero
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", Order(4), connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),                     // invalid channel order
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", UNORDERED, invalidConnHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),             // connection hops more than 1
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", UNORDERED, invalidShortConnHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),        // too short connection id
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", UNORDERED, invalidLongConnHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),         // too long connection id
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", UNORDERED, []string{invalidConnection}, "testcpport", "testcpchannel", "1.0", proof, 1, addr), // connection id contains non-alpha
		NewMsgChannelOpenTry("testport", "testchannel", "", UNORDERED, connHops, "testcpport", "testcpchannel", "1.0", proof, 1, addr),                       // empty channel version
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", UNORDERED, connHops, invalidPort, "testcpchannel", "1.0", proof, 1, addr),                     // invalid counterparty port id
		NewMsgChannelOpenTry("testport", "testchannel", "1.0", UNORDERED, connHops, "testcpport", invalidChannel, "1.0", proof, 1, addr),                     // invalid counterparty channel id
	}

	testCases := []struct {
		msg     MsgChannelOpenTry
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
		{testMsgs[7], false, "empty counterparty version"},
		{testMsgs[8], false, "empty proof"},
		{testMsgs[9], false, "proof height is zero"},
		{testMsgs[10], false, "invalid channel order"},
		{testMsgs[11], false, "connection hops more than 1 "},
		{testMsgs[12], false, "too short connection id"},
		{testMsgs[13], false, "too long connection id"},
		{testMsgs[14], false, "connection id contains non-alpha"},
		{testMsgs[15], false, "empty channel version"},
		{testMsgs[16], false, "invalid counterparty port id"},
		{testMsgs[17], false, "invalid counterparty channel id"},
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

// TestMsgChannelOpenAck tests ValidateBasic for MsgChannelOpenAck
func TestMsgChannelOpenAck(t *testing.T) {
	testMsgs := []MsgChannelOpenAck{
		NewMsgChannelOpenAck("testport", "testchannel", "1.0", proof, 1, addr),       // valid msg
		NewMsgChannelOpenAck(invalidShortPort, "testchannel", "1.0", proof, 1, addr), // too short port id
		NewMsgChannelOpenAck(invalidLongPort, "testchannel", "1.0", proof, 1, addr),  // too long port id
		NewMsgChannelOpenAck(invalidPort, "testchannel", "1.0", proof, 1, addr),      // port id contains non-alpha
		NewMsgChannelOpenAck("testport", invalidShortChannel, "1.0", proof, 1, addr), // too short channel id
		NewMsgChannelOpenAck("testport", invalidLongChannel, "1.0", proof, 1, addr),  // too long channel id
		NewMsgChannelOpenAck("testport", invalidChannel, "1.0", proof, 1, addr),      // channel id contains non-alpha
		NewMsgChannelOpenAck("testport", "testchannel", "", proof, 1, addr),          // empty counterparty version
		NewMsgChannelOpenAck("testport", "testchannel", "1.0", nil, 1, addr),         // empty proof
		NewMsgChannelOpenAck("testport", "testchannel", "1.0", proof, 0, addr),       // proof height is zero
	}

	testCases := []struct {
		msg     MsgChannelOpenAck
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
		{testMsgs[7], false, "empty counterparty version"},
		{testMsgs[8], false, "empty proof"},
		{testMsgs[9], false, "proof height is zero"},
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

// TestMsgChannelOpenConfirm tests ValidateBasic for MsgChannelOpenConfirm
func TestMsgChannelOpenConfirm(t *testing.T) {
	testMsgs := []MsgChannelOpenConfirm{
		NewMsgChannelOpenConfirm("testport", "testchannel", proof, 1, addr),       // valid msg
		NewMsgChannelOpenConfirm(invalidShortPort, "testchannel", proof, 1, addr), // too short port id
		NewMsgChannelOpenConfirm(invalidLongPort, "testchannel", proof, 1, addr),  // too long port id
		NewMsgChannelOpenConfirm(invalidPort, "testchannel", proof, 1, addr),      // port id contains non-alpha
		NewMsgChannelOpenConfirm("testport", invalidShortChannel, proof, 1, addr), // too short channel id
		NewMsgChannelOpenConfirm("testport", invalidLongChannel, proof, 1, addr),  // too long channel id
		NewMsgChannelOpenConfirm("testport", invalidChannel, proof, 1, addr),      // channel id contains non-alpha
		NewMsgChannelOpenConfirm("testport", "testchannel", nil, 1, addr),         // empty proof
		NewMsgChannelOpenConfirm("testport", "testchannel", proof, 0, addr),       // proof height is zero
	}

	testCases := []struct {
		msg     MsgChannelOpenConfirm
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
		{testMsgs[7], false, "empty proof"},
		{testMsgs[8], false, "proof height is zero"},
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

// TestMsgChannelCloseInit tests ValidateBasic for MsgChannelCloseInit
func TestMsgChannelCloseInit(t *testing.T) {
	testMsgs := []MsgChannelCloseInit{
		NewMsgChannelCloseInit("testport", "testchannel", addr),       // valid msg
		NewMsgChannelCloseInit(invalidShortPort, "testchannel", addr), // too short port id
		NewMsgChannelCloseInit(invalidLongPort, "testchannel", addr),  // too long port id
		NewMsgChannelCloseInit(invalidPort, "testchannel", addr),      // port id contains non-alpha
		NewMsgChannelCloseInit("testport", invalidShortChannel, addr), // too short channel id
		NewMsgChannelCloseInit("testport", invalidLongChannel, addr),  // too long channel id
		NewMsgChannelCloseInit("testport", invalidChannel, addr),      // channel id contains non-alpha
	}

	testCases := []struct {
		msg     MsgChannelCloseInit
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

// TestMsgChannelCloseConfirm tests ValidateBasic for MsgChannelCloseConfirm
func TestMsgChannelCloseConfirm(t *testing.T) {
	testMsgs := []MsgChannelCloseConfirm{
		NewMsgChannelCloseConfirm("testport", "testchannel", proof, 1, addr),       // valid msg
		NewMsgChannelCloseConfirm(invalidShortPort, "testchannel", proof, 1, addr), // too short port id
		NewMsgChannelCloseConfirm(invalidLongPort, "testchannel", proof, 1, addr),  // too long port id
		NewMsgChannelCloseConfirm(invalidPort, "testchannel", proof, 1, addr),      // port id contains non-alpha
		NewMsgChannelCloseConfirm("testport", invalidShortChannel, proof, 1, addr), // too short channel id
		NewMsgChannelCloseConfirm("testport", invalidLongChannel, proof, 1, addr),  // too long channel id
		NewMsgChannelCloseConfirm("testport", invalidChannel, proof, 1, addr),      // channel id contains non-alpha
		NewMsgChannelCloseConfirm("testport", "testchannel", nil, 1, addr),         // empty proof
		NewMsgChannelCloseConfirm("testport", "testchannel", proof, 0, addr),       // proof height is zero
	}

	testCases := []struct {
		msg     MsgChannelCloseConfirm
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
		{testMsgs[7], false, "empty proof"},
		{testMsgs[8], false, "proof height is zero"},
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
