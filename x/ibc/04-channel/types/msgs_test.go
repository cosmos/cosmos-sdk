package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// define constants used for testing
const (
	invalidPort      = "(invalidport1)"
	invalidShortPort = "p"
	invalidLongPort  = "invalidlongportinvalidlongport"

	invalidChannel      = "(invalidchannel1)"
	invalidShortChannel = "invalidch"
	invalidLongChannel  = "invalidlongchannelinvalidlongchannel"

	invalidConnection      = "(invalidconnection1)"
	invalidShortConnection = "invalidcn"
	invalidLongConnection  = "invalidlongconnection"
)

// define variables used for testing
var (
	connHops             = []string{"testconnection"}
	invalidConnHops      = []string{"testconnection", "testconnection"}
	invalidShortConnHops = []string{invalidShortConnection}
	invalidLongConnHops  = []string{invalidLongConnection}

	proof = commitmenttypes.MerkleProof{Proof: &merkle.Proof{Ops: []merkle.ProofOp{{Type: "proof", Key: []byte("key"), Data: []byte("data")}}}}

	addr = sdk.AccAddress("testaddr")
)

type MsgTestSuite struct {
	suite.Suite

	proof commitmenttypes.MerkleProof
}

func (suite *MsgTestSuite) SetupTest() {
	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db)
	storeKey := storetypes.NewKVStoreKey("iavlStoreKey")

	store.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, nil)
	store.LoadVersion(0)
	iavlStore := store.GetCommitStore(storeKey).(*iavl.Store)

	iavlStore.Set([]byte("KEY"), []byte("VALUE"))
	_ = store.Commit()

	res := store.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("/%s/key", storeKey.Name()), // required path to get key/value+proof
		Data:  []byte("KEY"),
		Prove: true,
	})

	suite.proof = commitmenttypes.MerkleProof{Proof: res.Proof}
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}

// TestMsgChannelOpenInit tests ValidateBasic for MsgChannelOpenInit
func (suite *MsgTestSuite) TestMsgChannelOpenInit() {
	testMsgs := []MsgChannelOpenInit{
		NewMsgChannelOpenInit("testportid", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                      // valid msg
		NewMsgChannelOpenInit(invalidShortPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                  // too short port id
		NewMsgChannelOpenInit(invalidLongPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                   // too long port id
		NewMsgChannelOpenInit(invalidPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                       // port id contains non-alpha
		NewMsgChannelOpenInit("testportid", invalidShortChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                // too short channel id
		NewMsgChannelOpenInit("testportid", invalidLongChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                 // too long channel id
		NewMsgChannelOpenInit("testportid", invalidChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", addr),                     // channel id contains non-alpha
		NewMsgChannelOpenInit("testportid", "testchannel", "1.0", Order(3), connHops, "testcpport", "testcpchannel", addr),                     // invalid channel order
		NewMsgChannelOpenInit("testportid", "testchannel", "1.0", ORDERED, invalidConnHops, "testcpport", "testcpchannel", addr),               // connection hops more than 1
		NewMsgChannelOpenInit("testportid", "testchannel", "1.0", UNORDERED, invalidShortConnHops, "testcpport", "testcpchannel", addr),        // too short connection id
		NewMsgChannelOpenInit("testportid", "testchannel", "1.0", UNORDERED, invalidLongConnHops, "testcpport", "testcpchannel", addr),         // too long connection id
		NewMsgChannelOpenInit("testportid", "testchannel", "1.0", UNORDERED, []string{invalidConnection}, "testcpport", "testcpchannel", addr), // connection id contains non-alpha
		NewMsgChannelOpenInit("testportid", "testchannel", "", UNORDERED, connHops, "testcpport", "testcpchannel", addr),                       // empty channel version
		NewMsgChannelOpenInit("testportid", "testchannel", "1.0", UNORDERED, connHops, invalidPort, "testcpchannel", addr),                     // invalid counterparty port id
		NewMsgChannelOpenInit("testportid", "testchannel", "1.0", UNORDERED, connHops, "testcpport", invalidChannel, addr),                     // invalid counterparty channel id
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
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgChannelOpenTry tests ValidateBasic for MsgChannelOpenTry
func (suite *MsgTestSuite) TestMsgChannelOpenTry() {
	msg0, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // valid msg
	suite.Require().NoError(err)
	msg1, err := NewMsgChannelOpenTry(invalidShortPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // too short port id
	suite.Require().NoError(err)
	msg2, err := NewMsgChannelOpenTry(invalidLongPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // too long port id
	suite.Require().NoError(err)
	msg3, err := NewMsgChannelOpenTry(invalidPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // port id contains non-alpha
	suite.Require().NoError(err)
	msg4, err := NewMsgChannelOpenTry("testportid", invalidShortChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // too short channel id
	suite.Require().NoError(err)
	msg5, err := NewMsgChannelOpenTry("testportid", invalidLongChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // too long channel id
	suite.Require().NoError(err)
	msg6, err := NewMsgChannelOpenTry("testportid", invalidChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // channel id contains non-alpha
	suite.Require().NoError(err)
	msg7, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "", suite.proof, 1, addr) // empty counterparty version
	suite.Require().NoError(err)
	msg8, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 0, addr) // suite.proof height is zero
	suite.Require().NoError(err)
	msg9, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", Order(4), connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // invalid channel order
	suite.Require().NoError(err)
	msg10, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, invalidConnHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // connection hops more than 1
	suite.Require().NoError(err)
	msg11, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, invalidShortConnHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // too short connection id
	suite.Require().NoError(err)
	msg12, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, invalidLongConnHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // too long connection id
	suite.Require().NoError(err)
	msg13, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, []string{invalidConnection}, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // connection id contains non-alpha
	suite.Require().NoError(err)
	msg14, err := NewMsgChannelOpenTry("testportid", "testchannel", "", UNORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr) // empty channel version
	suite.Require().NoError(err)
	msg15, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, connHops, invalidPort, "testcpchannel", "1.0", suite.proof, 1, addr) // invalid counterparty port id
	suite.Require().NoError(err)
	msg16, err := NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, connHops, "testcpport", invalidChannel, "1.0", suite.proof, 1, addr) // invalid counterparty channel id
	suite.Require().NoError(err)

	testCases := []struct {
		msg     MsgChannelOpenTry
		expPass bool
		errMsg  string
	}{
		{msg0, true, ""},
		{msg1, false, "too short port id"},
		{msg2, false, "too long port id"},
		{msg3, false, "port id contains non-alpha"},
		{msg4, false, "too short channel id"},
		{msg5, false, "too long channel id"},
		{msg6, false, "channel id contains non-alpha"},
		{msg7, false, "empty counterparty version"},
		{msg8, false, "proof height is zero"},
		{msg9, false, "invalid channel order"},
		{msg10, false, "connection hops more than 1 "},
		{msg11, false, "too short connection id"},
		{msg12, false, "too long connection id"},
		{msg13, false, "connection id contains non-alpha"},
		{msg14, false, "empty channel version"},
		{msg15, false, "invalid counterparty port id"},
		{msg16, false, "invalid counterparty channel id"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgChannelOpenAck tests ValidateBasic for MsgChannelOpenAck
func (suite *MsgTestSuite) TestMsgChannelOpenAck() {
	msg0, err := NewMsgChannelOpenAck("testportid", "testchannel", "1.0", suite.proof, 1, addr) // valid msg
	suite.Require().NoError(err)
	msg1, err := NewMsgChannelOpenAck(invalidShortPort, "testchannel", "1.0", suite.proof, 1, addr) // too short port id
	suite.Require().NoError(err)
	msg2, err := NewMsgChannelOpenAck(invalidLongPort, "testchannel", "1.0", suite.proof, 1, addr) // too long port id
	suite.Require().NoError(err)
	msg3, err := NewMsgChannelOpenAck(invalidPort, "testchannel", "1.0", suite.proof, 1, addr) // port id contains non-alpha
	suite.Require().NoError(err)
	msg4, err := NewMsgChannelOpenAck("testportid", invalidShortChannel, "1.0", suite.proof, 1, addr) // too short channel id
	suite.Require().NoError(err)
	msg5, err := NewMsgChannelOpenAck("testportid", invalidLongChannel, "1.0", suite.proof, 1, addr) // too long channel id
	suite.Require().NoError(err)
	msg6, err := NewMsgChannelOpenAck("testportid", invalidChannel, "1.0", suite.proof, 1, addr) // channel id contains non-alpha
	suite.Require().NoError(err)
	msg7, err := NewMsgChannelOpenAck("testportid", "testchannel", "", suite.proof, 1, addr) // empty counterparty version
	suite.Require().NoError(err)
	msg8, err := NewMsgChannelOpenAck("testportid", "testchannel", "1.0", commitmenttypes.MerkleProof{Proof: nil}, 1, addr) // empty proof
	suite.Require().NoError(err)
	msg9, err := NewMsgChannelOpenAck("testportid", "testchannel", "1.0", suite.proof, 0, addr) // proof height is zero
	suite.Require().NoError(err)

	testCases := []struct {
		msg     MsgChannelOpenAck
		expPass bool
		errMsg  string
	}{
		{msg0, true, ""},
		{msg1, false, "too short port id"},
		{msg2, false, "too long port id"},
		{msg3, false, "port id contains non-alpha"},
		{msg4, false, "too short channel id"},
		{msg5, false, "too long channel id"},
		{msg6, false, "channel id contains non-alpha"},
		{msg7, false, "empty counterparty version"},
		{msg8, false, "empty proof"},
		{msg9, false, "proof height is zero"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgChannelOpenConfirm tests ValidateBasic for MsgChannelOpenConfirm
func (suite *MsgTestSuite) TestMsgChannelOpenConfirm() {
	msg0, err := NewMsgChannelOpenConfirm("testportid", "testchannel", suite.proof, 1, addr) // valid msg
	suite.Require().NoError(err)
	msg1, err := NewMsgChannelOpenConfirm(invalidShortPort, "testchannel", suite.proof, 1, addr) // too short port id
	suite.Require().NoError(err)
	msg2, err := NewMsgChannelOpenConfirm(invalidLongPort, "testchannel", suite.proof, 1, addr) // too long port id
	suite.Require().NoError(err)
	msg3, err := NewMsgChannelOpenConfirm(invalidPort, "testchannel", suite.proof, 1, addr) // port id contains non-alpha
	suite.Require().NoError(err)
	msg4, err := NewMsgChannelOpenConfirm("testportid", invalidShortChannel, suite.proof, 1, addr) // too short channel id
	suite.Require().NoError(err)
	msg5, err := NewMsgChannelOpenConfirm("testportid", invalidLongChannel, suite.proof, 1, addr) // too long channel id
	suite.Require().NoError(err)
	msg6, err := NewMsgChannelOpenConfirm("testportid", invalidChannel, suite.proof, 1, addr) // channel id contains non-alpha
	suite.Require().NoError(err)
	msg7, err := NewMsgChannelOpenConfirm("testportid", "testchannel", commitmenttypes.MerkleProof{Proof: nil}, 1, addr) // empty proof
	suite.Require().NoError(err)
	msg8, err := NewMsgChannelOpenConfirm("testportid", "testchannel", suite.proof, 0, addr) // proof height is zero
	suite.Require().NoError(err)

	testCases := []struct {
		msg     MsgChannelOpenConfirm
		expPass bool
		errMsg  string
	}{
		{msg0, true, ""},
		{msg1, false, "too short port id"},
		{msg2, false, "too long port id"},
		{msg3, false, "port id contains non-alpha"},
		{msg4, false, "too short channel id"},
		{msg5, false, "too long channel id"},
		{msg6, false, "channel id contains non-alpha"},
		{msg7, false, "empty proof"},
		{msg8, false, "proof height is zero"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgChannelCloseInit tests ValidateBasic for MsgChannelCloseInit
func (suite *MsgTestSuite) TestMsgChannelCloseInit() {
	testMsgs := []MsgChannelCloseInit{
		NewMsgChannelCloseInit("testportid", "testchannel", addr),       // valid msg
		NewMsgChannelCloseInit(invalidShortPort, "testchannel", addr),   // too short port id
		NewMsgChannelCloseInit(invalidLongPort, "testchannel", addr),    // too long port id
		NewMsgChannelCloseInit(invalidPort, "testchannel", addr),        // port id contains non-alpha
		NewMsgChannelCloseInit("testportid", invalidShortChannel, addr), // too short channel id
		NewMsgChannelCloseInit("testportid", invalidLongChannel, addr),  // too long channel id
		NewMsgChannelCloseInit("testportid", invalidChannel, addr),      // channel id contains non-alpha
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
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgChannelCloseConfirm tests ValidateBasic for MsgChannelCloseConfirm
func (suite *MsgTestSuite) TestMsgChannelCloseConfirm() {
	msg0, err := NewMsgChannelCloseConfirm("testportid", "testchannel", suite.proof, 1, addr) // valid msg
	suite.Require().NoError(err)
	msg1, err := NewMsgChannelCloseConfirm(invalidShortPort, "testchannel", suite.proof, 1, addr) // too short port id
	suite.Require().NoError(err)
	msg2, err := NewMsgChannelCloseConfirm(invalidLongPort, "testchannel", suite.proof, 1, addr) // too long port id
	suite.Require().NoError(err)
	msg3, err := NewMsgChannelCloseConfirm(invalidPort, "testchannel", suite.proof, 1, addr) // port id contains non-alpha
	suite.Require().NoError(err)
	msg4, err := NewMsgChannelCloseConfirm("testportid", invalidShortChannel, suite.proof, 1, addr) // too short channel id
	suite.Require().NoError(err)
	msg5, err := NewMsgChannelCloseConfirm("testportid", invalidLongChannel, suite.proof, 1, addr) // too long channel id
	suite.Require().NoError(err)
	msg6, err := NewMsgChannelCloseConfirm("testportid", invalidChannel, suite.proof, 1, addr) // channel id contains non-alpha
	suite.Require().NoError(err)
	msg7, err := NewMsgChannelCloseConfirm("testportid", "testchannel", commitmenttypes.MerkleProof{Proof: nil}, 1, addr) // empty proof
	suite.Require().NoError(err)
	msg8, err := NewMsgChannelCloseConfirm("testportid", "testchannel", suite.proof, 0, addr) // proof height is zero
	suite.Require().NoError(err)

	testCases := []struct {
		msg     MsgChannelCloseConfirm
		expPass bool
		errMsg  string
	}{
		{msg0, true, ""},
		{msg1, false, "too short port id"},
		{msg2, false, "too long port id"},
		{msg3, false, "port id contains non-alpha"},
		{msg4, false, "too short channel id"},
		{msg5, false, "too long channel id"},
		{msg6, false, "channel id contains non-alpha"},
		{msg7, false, "empty proof"},
		{msg8, false, "proof height is zero"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// define variables used for testing
var (
	timeoutHeight     = uint64(100)
	timeoutTimestamp  = uint64(100)
	disabledTimeout   = uint64(0)
	validPacketData   = []byte("testdata")
	unknownPacketData = []byte("unknown")
	invalidAckData    = []byte("123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890")

	packet        = NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp)
	unknownPacket = NewPacket(unknownPacketData, 0, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp)
	invalidAck    = invalidAckData

	emptyProof     = commitmenttypes.MerkleProof{Proof: nil}
	invalidProofs1 = commitmentexported.Proof(nil)
	invalidProofs2 = emptyProof

	addr1     = sdk.AccAddress("testaddr1")
	emptyAddr sdk.AccAddress

	portid   = "testportid"
	chanid   = "testchannel"
	cpportid = "testcpport"
	cpchanid = "testcpchannel"
)

// TestMsgPacketType tests Type for MsgPacket
func TestMsgPacketType(t *testing.T) {
	msg, err := NewMsgPacket(packet, proof, 1, addr1)
	require.NoError(t, err)

	require.Equal(t, []byte("testdata"), msg.Packet.GetData())
}

// TestMsgPacketValidation tests ValidateBasic for MsgPacket
func TestMsgPacketValidation(t *testing.T) {
	msg0, err := NewMsgPacket(packet, proof, 1, addr1) // valid msg
	require.NoError(t, err)
	msg1, err := NewMsgPacket(packet, proof, 0, addr1) // proof height is zero
	require.NoError(t, err)
	msg2, err := NewMsgPacket(packet, invalidProofs2, 1, addr1) // proof contain empty proof
	require.NoError(t, err)
	msg3, err := NewMsgPacket(packet, proof, 1, emptyAddr) // missing signer address
	require.NoError(t, err)
	msg4, err := NewMsgPacket(unknownPacket, proof, 1, addr1) // unknown packet
	require.NoError(t, err)

	testCases := []struct {
		msg     MsgPacket
		expPass bool
		errMsg  string
	}{
		{msg0, true, ""},
		{msg1, false, "proof height is zero"},
		{msg2, false, "proof contain empty proof"},
		{msg3, false, "missing signer address"},
		{msg4, false, "invalid packet"},
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

// TestMsgPacketGetSignBytes tests GetSignBytes for MsgPacket
func TestMsgPacketGetSignBytes(t *testing.T) {
	msg, err := NewMsgPacket(packet, proof, 1, addr1)
	require.NoError(t, err)

	res := msg.GetSignBytes()

	expected := fmt.Sprintf(
		`{"type":"ibc/channel/MsgPacket","value":{"packet":{"data":%s,"destination_channel":"testcpchannel","destination_port":"testcpport","sequence":"1","source_channel":"testchannel","source_port":"testportid","timeout_height":"100","timeout_timestamp":"100"},"proof":{"proof":{"ops":[{"data":"ZGF0YQ==","key":"a2V5","type":"proof"}]}},"proof_height":"1","signer":"cosmos1w3jhxarpv3j8yvg4ufs4x"}}`,
		string(msg.GetDataSignBytes()),
	)
	require.Equal(t, expected, string(res))
}

// TestMsgPacketGetSigners tests GetSigners for MsgPacket
func TestMsgPacketGetSigners(t *testing.T) {
	msg, err := NewMsgPacket(packet, proof, 1, addr1)
	require.NoError(t, err)

	res := msg.GetSigners()

	expected := "[746573746164647231]"
	require.Equal(t, expected, fmt.Sprintf("%v", res))
}

// TestMsgTimeout tests ValidateBasic for MsgTimeout
func (suite *MsgTestSuite) TestMsgTimeout() {
	msg0, err := NewMsgTimeout(packet, 1, proof, 1, addr)
	suite.Require().NoError(err)
	msg1, err := NewMsgTimeout(packet, 1, proof, 0, addr)
	suite.Require().NoError(err)
	msg2, err := NewMsgTimeout(packet, 1, proof, 1, emptyAddr)
	suite.Require().NoError(err)
	msg3, err := NewMsgTimeout(packet, 1, emptyProof, 1, addr)
	suite.Require().NoError(err)
	msg4, err := NewMsgTimeout(unknownPacket, 1, proof, 1, addr)
	suite.Require().NoError(err)

	testCases := []struct {
		msg     MsgTimeout
		expPass bool
		errMsg  string
	}{
		{msg0, true, ""},
		{msg1, false, "proof height must be > 0"},
		{msg2, false, "missing signer address"},
		{msg3, false, "cannot submit an empty proof"},
		{msg4, false, "invalid packet"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgAcknowledgement tests ValidateBasic for MsgAcknowledgement
func (suite *MsgTestSuite) TestMsgAcknowledgement() {
	msg0, err := NewMsgAcknowledgement(packet, packet.GetData(), proof, 1, addr)
	suite.Require().NoError(err)
	msg1, err := NewMsgAcknowledgement(packet, packet.GetData(), proof, 0, addr)
	suite.Require().NoError(err)
	msg2, err := NewMsgAcknowledgement(packet, packet.GetData(), proof, 1, emptyAddr)
	suite.Require().NoError(err)
	msg3, err := NewMsgAcknowledgement(packet, packet.GetData(), emptyProof, 1, addr)
	suite.Require().NoError(err)
	msg4, err := NewMsgAcknowledgement(unknownPacket, packet.GetData(), proof, 1, addr)
	suite.Require().NoError(err)
	msg5, err := NewMsgAcknowledgement(packet, invalidAck, proof, 1, addr)
	suite.Require().NoError(err)

	testCases := []struct {
		msg     MsgAcknowledgement
		expPass bool
		errMsg  string
	}{
		{msg0, true, ""},
		{msg1, false, "proof height must be > 0"},
		{msg2, false, "missing signer address"},
		{msg3, false, "cannot submit an empty proof"},
		{msg4, false, "invalid packet"},
		{msg5, false, "invalid acknowledgement"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}
