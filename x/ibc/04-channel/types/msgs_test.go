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
	testMsgs := []MsgChannelOpenTry{
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),                      // valid msg
		NewMsgChannelOpenTry(invalidShortPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),                  // too short port id
		NewMsgChannelOpenTry(invalidLongPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),                   // too long port id
		NewMsgChannelOpenTry(invalidPort, "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),                       // port id contains non-alpha
		NewMsgChannelOpenTry("testportid", invalidShortChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),                // too short channel id
		NewMsgChannelOpenTry("testportid", invalidLongChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),                 // too long channel id
		NewMsgChannelOpenTry("testportid", invalidChannel, "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),                     // channel id contains non-alpha
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "", suite.proof, 1, addr),                         // empty counterparty version
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 0, addr),                      // suite.proof height is zero
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", Order(4), connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),                     // invalid channel order
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, invalidConnHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),             // connection hops more than 1
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, invalidShortConnHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),        // too short connection id
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, invalidLongConnHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),         // too long connection id
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, []string{invalidConnection}, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr), // connection id contains non-alpha
		NewMsgChannelOpenTry("testportid", "testchannel", "", UNORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, 1, addr),                       // empty channel version
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, connHops, invalidPort, "testcpchannel", "1.0", suite.proof, 1, addr),                     // invalid counterparty port id
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", UNORDERED, connHops, "testcpport", invalidChannel, "1.0", suite.proof, 1, addr),                     // invalid counterparty channel id
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
		{testMsgs[8], false, "proof height is zero"},
		{testMsgs[9], false, "invalid channel order"},
		{testMsgs[10], false, "connection hops more than 1 "},
		{testMsgs[11], false, "too short connection id"},
		{testMsgs[12], false, "too long connection id"},
		{testMsgs[13], false, "connection id contains non-alpha"},
		{testMsgs[14], false, "empty channel version"},
		{testMsgs[15], false, "invalid counterparty port id"},
		{testMsgs[16], false, "invalid counterparty channel id"},
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
	testMsgs := []MsgChannelOpenAck{
		NewMsgChannelOpenAck("testportid", "testchannel", "1.0", suite.proof, 1, addr),                             // valid msg
		NewMsgChannelOpenAck(invalidShortPort, "testchannel", "1.0", suite.proof, 1, addr),                         // too short port id
		NewMsgChannelOpenAck(invalidLongPort, "testchannel", "1.0", suite.proof, 1, addr),                          // too long port id
		NewMsgChannelOpenAck(invalidPort, "testchannel", "1.0", suite.proof, 1, addr),                              // port id contains non-alpha
		NewMsgChannelOpenAck("testportid", invalidShortChannel, "1.0", suite.proof, 1, addr),                       // too short channel id
		NewMsgChannelOpenAck("testportid", invalidLongChannel, "1.0", suite.proof, 1, addr),                        // too long channel id
		NewMsgChannelOpenAck("testportid", invalidChannel, "1.0", suite.proof, 1, addr),                            // channel id contains non-alpha
		NewMsgChannelOpenAck("testportid", "testchannel", "", suite.proof, 1, addr),                                // empty counterparty version
		NewMsgChannelOpenAck("testportid", "testchannel", "1.0", commitmenttypes.MerkleProof{Proof: nil}, 1, addr), // empty proof
		NewMsgChannelOpenAck("testportid", "testchannel", "1.0", suite.proof, 0, addr),                             // proof height is zero
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
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgChannelOpenConfirm tests ValidateBasic for MsgChannelOpenConfirm
func (suite *MsgTestSuite) TestMsgChannelOpenConfirm() {
	testMsgs := []MsgChannelOpenConfirm{
		NewMsgChannelOpenConfirm("testportid", "testchannel", suite.proof, 1, addr),                             // valid msg
		NewMsgChannelOpenConfirm(invalidShortPort, "testchannel", suite.proof, 1, addr),                         // too short port id
		NewMsgChannelOpenConfirm(invalidLongPort, "testchannel", suite.proof, 1, addr),                          // too long port id
		NewMsgChannelOpenConfirm(invalidPort, "testchannel", suite.proof, 1, addr),                              // port id contains non-alpha
		NewMsgChannelOpenConfirm("testportid", invalidShortChannel, suite.proof, 1, addr),                       // too short channel id
		NewMsgChannelOpenConfirm("testportid", invalidLongChannel, suite.proof, 1, addr),                        // too long channel id
		NewMsgChannelOpenConfirm("testportid", invalidChannel, suite.proof, 1, addr),                            // channel id contains non-alpha
		NewMsgChannelOpenConfirm("testportid", "testchannel", commitmenttypes.MerkleProof{Proof: nil}, 1, addr), // empty proof
		NewMsgChannelOpenConfirm("testportid", "testchannel", suite.proof, 0, addr),                             // proof height is zero
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
	testMsgs := []MsgChannelCloseConfirm{
		NewMsgChannelCloseConfirm("testportid", "testchannel", suite.proof, 1, addr),                             // valid msg
		NewMsgChannelCloseConfirm(invalidShortPort, "testchannel", suite.proof, 1, addr),                         // too short port id
		NewMsgChannelCloseConfirm(invalidLongPort, "testchannel", suite.proof, 1, addr),                          // too long port id
		NewMsgChannelCloseConfirm(invalidPort, "testchannel", suite.proof, 1, addr),                              // port id contains non-alpha
		NewMsgChannelCloseConfirm("testportid", invalidShortChannel, suite.proof, 1, addr),                       // too short channel id
		NewMsgChannelCloseConfirm("testportid", invalidLongChannel, suite.proof, 1, addr),                        // too long channel id
		NewMsgChannelCloseConfirm("testportid", invalidChannel, suite.proof, 1, addr),                            // channel id contains non-alpha
		NewMsgChannelCloseConfirm("testportid", "testchannel", commitmenttypes.MerkleProof{Proof: nil}, 1, addr), // empty proof
		NewMsgChannelCloseConfirm("testportid", "testchannel", suite.proof, 0, addr),                             // proof height is zero
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
	msg := NewMsgPacket(packet, proof, 1, addr1)

	require.Equal(t, []byte("testdata"), msg.Packet.GetData())
}

// TestMsgPacketValidation tests ValidateBasic for MsgPacket
func TestMsgPacketValidation(t *testing.T) {
	testMsgs := []MsgPacket{
		NewMsgPacket(packet, proof, 1, addr1),          // valid msg
		NewMsgPacket(packet, proof, 0, addr1),          // proof height is zero
		NewMsgPacket(packet, invalidProofs2, 1, addr1), // proof contain empty proof
		NewMsgPacket(packet, proof, 1, emptyAddr),      // missing signer address
		NewMsgPacket(unknownPacket, proof, 1, addr1),   // unknown packet
	}

	testCases := []struct {
		msg     MsgPacket
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], true, ""},
		{testMsgs[1], false, "proof height is zero"},
		{testMsgs[2], false, "proof contain empty proof"},
		{testMsgs[3], false, "missing signer address"},
		{testMsgs[4], false, "invalid packet"},
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
	msg := NewMsgPacket(packet, proof, 1, addr1)
	res := msg.GetSignBytes()

	expected := fmt.Sprintf(
		`{"type":"ibc/channel/MsgPacket","value":{"packet":{"data":%s,"destination_channel":"testcpchannel","destination_port":"testcpport","sequence":"1","source_channel":"testchannel","source_port":"testportid","timeout_height":"100","timeout_timestamp":"100"},"proof":{"proof":{"ops":[{"data":"ZGF0YQ==","key":"a2V5","type":"proof"}]}},"proof_height":"1","signer":"cosmos1w3jhxarpv3j8yvg4ufs4x"}}`,
		string(msg.GetDataSignBytes()),
	)
	require.Equal(t, expected, string(res))
}

// TestMsgPacketGetSigners tests GetSigners for MsgPacket
func TestMsgPacketGetSigners(t *testing.T) {
	msg := NewMsgPacket(packet, proof, 1, addr1)
	res := msg.GetSigners()

	expected := "[746573746164647231]"
	require.Equal(t, expected, fmt.Sprintf("%v", res))
}

// TestMsgTimeout tests ValidateBasic for MsgTimeout
func (suite *MsgTestSuite) TestMsgTimeout() {
	testMsgs := []MsgTimeout{
		NewMsgTimeout(packet, 1, proof, 1, addr),
		NewMsgTimeout(packet, 1, proof, 0, addr),
		NewMsgTimeout(packet, 1, proof, 1, emptyAddr),
		NewMsgTimeout(packet, 1, emptyProof, 1, addr),
		NewMsgTimeout(unknownPacket, 1, proof, 1, addr),
	}

	testCases := []struct {
		msg     MsgTimeout
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], true, ""},
		{testMsgs[1], false, "proof height must be > 0"},
		{testMsgs[2], false, "missing signer address"},
		{testMsgs[3], false, "cannot submit an empty proof"},
		{testMsgs[4], false, "invalid packet"},
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
	testMsgs := []MsgAcknowledgement{
		NewMsgAcknowledgement(packet, packet.GetData(), proof, 1, addr),
		NewMsgAcknowledgement(packet, packet.GetData(), proof, 0, addr),
		NewMsgAcknowledgement(packet, packet.GetData(), proof, 1, emptyAddr),
		NewMsgAcknowledgement(packet, packet.GetData(), emptyProof, 1, addr),
		NewMsgAcknowledgement(unknownPacket, packet.GetData(), proof, 1, addr),
		NewMsgAcknowledgement(packet, invalidAck, proof, 1, addr),
	}

	testCases := []struct {
		msg     MsgAcknowledgement
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], true, ""},
		{testMsgs[1], false, "proof height must be > 0"},
		{testMsgs[2], false, "missing signer address"},
		{testMsgs[3], false, "cannot submit an empty proof"},
		{testMsgs[4], false, "invalid packet"},
		{testMsgs[5], false, "invalid acknowledgement"},
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
