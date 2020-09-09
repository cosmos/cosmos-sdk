package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// define constants used for testing
const (
	invalidPort      = "(invalidport1)"
	invalidShortPort = "p"
	invalidLongPort  = "invalidlongportinvalidlongportinvalidlongportidinvalidlongportidinvalid"

	invalidChannel      = "(invalidchannel1)"
	invalidShortChannel = "invalidch"
	invalidLongChannel  = "invalidlongchannelinvalidlongchannelinvalidlongchannelinvalidlongchannel"

	invalidConnection      = "(invalidconnection1)"
	invalidShortConnection = "invalidcn"
	invalidLongConnection  = "invalidlongconnectioninvalidlongconnectioninvalidlongconnectioninvalid"
)

// define variables used for testing
var (
	height            = clienttypes.NewHeight(0, 1)
	timeoutHeight     = clienttypes.NewHeight(0, 100)
	timeoutTimestamp  = uint64(100)
	disabledTimeout   = clienttypes.ZeroHeight()
	validPacketData   = []byte("testdata")
	unknownPacketData = []byte("unknown")

	packet        = types.NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp)
	invalidPacket = types.NewPacket(unknownPacketData, 0, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp)

	emptyProof     = []byte{}
	invalidProofs1 = exported.Proof(nil)
	invalidProofs2 = emptyProof

	addr1     = sdk.AccAddress("testaddr1")
	emptyAddr sdk.AccAddress

	portid   = "testportid"
	chanid   = "testchannel"
	cpportid = "testcpport"
	cpchanid = "testcpchannel"

	connHops             = []string{"testconnection"}
	invalidConnHops      = []string{"testconnection", "testconnection"}
	invalidShortConnHops = []string{invalidShortConnection}
	invalidLongConnHops  = []string{invalidLongConnection}

	addr = sdk.AccAddress("testaddr")
)

type TypesTestSuite struct {
	suite.Suite

	proof []byte
}

func (suite *TypesTestSuite) SetupTest() {
	app := simapp.Setup(false)
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

	merkleProof := commitmenttypes.MerkleProof{Proof: res.ProofOps}
	proof, err := app.AppCodec().MarshalBinaryBare(&merkleProof)
	suite.NoError(err)

	suite.proof = proof
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

// TestMsgChannelOpenInit tests ValidateBasic for MsgChannelOpenInit
func (suite *TypesTestSuite) TestMsgChannelOpenInit() {
	testMsgs := []*types.MsgChannelOpenInit{
		types.NewMsgChannelOpenInit("testportid", "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", addr),                      // valid msg
		types.NewMsgChannelOpenInit(invalidShortPort, "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", addr),                  // too short port id
		types.NewMsgChannelOpenInit(invalidLongPort, "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", addr),                   // too long port id
		types.NewMsgChannelOpenInit(invalidPort, "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", addr),                       // port id contains non-alpha
		types.NewMsgChannelOpenInit("testportid", invalidShortChannel, "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", addr),                // too short channel id
		types.NewMsgChannelOpenInit("testportid", invalidLongChannel, "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", addr),                 // too long channel id
		types.NewMsgChannelOpenInit("testportid", invalidChannel, "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", addr),                     // channel id contains non-alpha
		types.NewMsgChannelOpenInit("testportid", "testchannel", "1.0", types.Order(3), connHops, "testcpport", "testcpchannel", addr),                     // invalid channel order
		types.NewMsgChannelOpenInit("testportid", "testchannel", "1.0", types.ORDERED, invalidConnHops, "testcpport", "testcpchannel", addr),               // connection hops more than 1
		types.NewMsgChannelOpenInit("testportid", "testchannel", "1.0", types.UNORDERED, invalidShortConnHops, "testcpport", "testcpchannel", addr),        // too short connection id
		types.NewMsgChannelOpenInit("testportid", "testchannel", "1.0", types.UNORDERED, invalidLongConnHops, "testcpport", "testcpchannel", addr),         // too long connection id
		types.NewMsgChannelOpenInit("testportid", "testchannel", "1.0", types.UNORDERED, []string{invalidConnection}, "testcpport", "testcpchannel", addr), // connection id contains non-alpha
		types.NewMsgChannelOpenInit("testportid", "testchannel", "", types.UNORDERED, connHops, "testcpport", "testcpchannel", addr),                       // empty channel version
		types.NewMsgChannelOpenInit("testportid", "testchannel", "1.0", types.UNORDERED, connHops, invalidPort, "testcpchannel", addr),                     // invalid counterparty port id
		types.NewMsgChannelOpenInit("testportid", "testchannel", "1.0", types.UNORDERED, connHops, "testcpport", invalidChannel, addr),                     // invalid counterparty channel id
	}

	testCases := []struct {
		msg     *types.MsgChannelOpenInit
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
		{testMsgs[12], true, ""},
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
func (suite *TypesTestSuite) TestMsgChannelOpenTry() {
	testMsgs := []*types.MsgChannelOpenTry{
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),                      // valid msg
		types.NewMsgChannelOpenTry(invalidShortPort, "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),                  // too short port id
		types.NewMsgChannelOpenTry(invalidLongPort, "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),                   // too long port id
		types.NewMsgChannelOpenTry(invalidPort, "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),                       // port id contains non-alpha
		types.NewMsgChannelOpenTry("testportid", invalidShortChannel, "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),                // too short channel id
		types.NewMsgChannelOpenTry("testportid", invalidLongChannel, "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),                 // too long channel id
		types.NewMsgChannelOpenTry("testportid", invalidChannel, "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),                     // channel id contains non-alpha
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", "", suite.proof, height, addr),                         // empty counterparty version
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.ORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, clienttypes.ZeroHeight(), addr),    // suite.proof height is zero
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.Order(4), connHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),                     // invalid channel order
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.UNORDERED, invalidConnHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),             // connection hops more than 1
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.UNORDERED, invalidShortConnHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),        // too short connection id
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.UNORDERED, invalidLongConnHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),         // too long connection id
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.UNORDERED, []string{invalidConnection}, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr), // connection id contains non-alpha
		types.NewMsgChannelOpenTry("testportid", "testchannel", "", types.UNORDERED, connHops, "testcpport", "testcpchannel", "1.0", suite.proof, height, addr),                       // empty channel version
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.UNORDERED, connHops, invalidPort, "testcpchannel", "1.0", suite.proof, height, addr),                     // invalid counterparty port id
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.UNORDERED, connHops, "testcpport", invalidChannel, "1.0", suite.proof, height, addr),                     // invalid counterparty channel id
		types.NewMsgChannelOpenTry("testportid", "testchannel", "1.0", types.UNORDERED, connHops, "testcpport", "testcpchannel", "1.0", emptyProof, height, addr),                     // empty proof
	}

	testCases := []struct {
		msg     *types.MsgChannelOpenTry
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
		{testMsgs[7], true, ""},
		{testMsgs[8], false, "proof height is zero"},
		{testMsgs[9], false, "invalid channel order"},
		{testMsgs[10], false, "connection hops more than 1 "},
		{testMsgs[11], false, "too short connection id"},
		{testMsgs[12], false, "too long connection id"},
		{testMsgs[13], false, "connection id contains non-alpha"},
		{testMsgs[14], true, ""},
		{testMsgs[15], false, "invalid counterparty port id"},
		{testMsgs[16], false, "invalid counterparty channel id"},
		{testMsgs[17], false, "empty proof"},
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
func (suite *TypesTestSuite) TestMsgChannelOpenAck() {
	testMsgs := []*types.MsgChannelOpenAck{
		types.NewMsgChannelOpenAck("testportid", "testchannel", "1.0", suite.proof, height, addr),                   // valid msg
		types.NewMsgChannelOpenAck(invalidShortPort, "testchannel", "1.0", suite.proof, height, addr),               // too short port id
		types.NewMsgChannelOpenAck(invalidLongPort, "testchannel", "1.0", suite.proof, height, addr),                // too long port id
		types.NewMsgChannelOpenAck(invalidPort, "testchannel", "1.0", suite.proof, height, addr),                    // port id contains non-alpha
		types.NewMsgChannelOpenAck("testportid", invalidShortChannel, "1.0", suite.proof, height, addr),             // too short channel id
		types.NewMsgChannelOpenAck("testportid", invalidLongChannel, "1.0", suite.proof, height, addr),              // too long channel id
		types.NewMsgChannelOpenAck("testportid", invalidChannel, "1.0", suite.proof, height, addr),                  // channel id contains non-alpha
		types.NewMsgChannelOpenAck("testportid", "testchannel", "", suite.proof, height, addr),                      // empty counterparty version
		types.NewMsgChannelOpenAck("testportid", "testchannel", "1.0", emptyProof, height, addr),                    // empty proof
		types.NewMsgChannelOpenAck("testportid", "testchannel", "1.0", suite.proof, clienttypes.ZeroHeight(), addr), // proof height is zero
	}

	testCases := []struct {
		msg     *types.MsgChannelOpenAck
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
		{testMsgs[7], true, ""},
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
func (suite *TypesTestSuite) TestMsgChannelOpenConfirm() {
	testMsgs := []*types.MsgChannelOpenConfirm{
		types.NewMsgChannelOpenConfirm("testportid", "testchannel", suite.proof, height, addr),                   // valid msg
		types.NewMsgChannelOpenConfirm(invalidShortPort, "testchannel", suite.proof, height, addr),               // too short port id
		types.NewMsgChannelOpenConfirm(invalidLongPort, "testchannel", suite.proof, height, addr),                // too long port id
		types.NewMsgChannelOpenConfirm(invalidPort, "testchannel", suite.proof, height, addr),                    // port id contains non-alpha
		types.NewMsgChannelOpenConfirm("testportid", invalidShortChannel, suite.proof, height, addr),             // too short channel id
		types.NewMsgChannelOpenConfirm("testportid", invalidLongChannel, suite.proof, height, addr),              // too long channel id
		types.NewMsgChannelOpenConfirm("testportid", invalidChannel, suite.proof, height, addr),                  // channel id contains non-alpha
		types.NewMsgChannelOpenConfirm("testportid", "testchannel", emptyProof, height, addr),                    // empty proof
		types.NewMsgChannelOpenConfirm("testportid", "testchannel", suite.proof, clienttypes.ZeroHeight(), addr), // proof height is zero
	}

	testCases := []struct {
		msg     *types.MsgChannelOpenConfirm
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
func (suite *TypesTestSuite) TestMsgChannelCloseInit() {
	testMsgs := []*types.MsgChannelCloseInit{
		types.NewMsgChannelCloseInit("testportid", "testchannel", addr),       // valid msg
		types.NewMsgChannelCloseInit(invalidShortPort, "testchannel", addr),   // too short port id
		types.NewMsgChannelCloseInit(invalidLongPort, "testchannel", addr),    // too long port id
		types.NewMsgChannelCloseInit(invalidPort, "testchannel", addr),        // port id contains non-alpha
		types.NewMsgChannelCloseInit("testportid", invalidShortChannel, addr), // too short channel id
		types.NewMsgChannelCloseInit("testportid", invalidLongChannel, addr),  // too long channel id
		types.NewMsgChannelCloseInit("testportid", invalidChannel, addr),      // channel id contains non-alpha
	}

	testCases := []struct {
		msg     *types.MsgChannelCloseInit
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
func (suite *TypesTestSuite) TestMsgChannelCloseConfirm() {
	testMsgs := []*types.MsgChannelCloseConfirm{
		types.NewMsgChannelCloseConfirm("testportid", "testchannel", suite.proof, height, addr),                   // valid msg
		types.NewMsgChannelCloseConfirm(invalidShortPort, "testchannel", suite.proof, height, addr),               // too short port id
		types.NewMsgChannelCloseConfirm(invalidLongPort, "testchannel", suite.proof, height, addr),                // too long port id
		types.NewMsgChannelCloseConfirm(invalidPort, "testchannel", suite.proof, height, addr),                    // port id contains non-alpha
		types.NewMsgChannelCloseConfirm("testportid", invalidShortChannel, suite.proof, height, addr),             // too short channel id
		types.NewMsgChannelCloseConfirm("testportid", invalidLongChannel, suite.proof, height, addr),              // too long channel id
		types.NewMsgChannelCloseConfirm("testportid", invalidChannel, suite.proof, height, addr),                  // channel id contains non-alpha
		types.NewMsgChannelCloseConfirm("testportid", "testchannel", emptyProof, height, addr),                    // empty proof
		types.NewMsgChannelCloseConfirm("testportid", "testchannel", suite.proof, clienttypes.ZeroHeight(), addr), // proof height is zero
	}

	testCases := []struct {
		msg     *types.MsgChannelCloseConfirm
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

// TestMsgRecvPacketType tests Type for MsgRecvPacket.
func (suite *TypesTestSuite) TestMsgRecvPacketType() {
	msg := types.NewMsgRecvPacket(packet, suite.proof, height, addr1)

	suite.Equal("recv_packet", msg.Type())
}

// TestMsgRecvPacketValidation tests ValidateBasic for MsgRecvPacket
func (suite *TypesTestSuite) TestMsgRecvPacketValidation() {
	testMsgs := []*types.MsgRecvPacket{
		types.NewMsgRecvPacket(packet, suite.proof, height, addr1),                   // valid msg
		types.NewMsgRecvPacket(packet, suite.proof, clienttypes.ZeroHeight(), addr1), // proof height is zero
		types.NewMsgRecvPacket(packet, emptyProof, height, addr1),                    // empty proof
		types.NewMsgRecvPacket(packet, suite.proof, height, emptyAddr),               // missing signer address
		types.NewMsgRecvPacket(invalidPacket, suite.proof, height, addr1),            // unknown packet
	}

	testCases := []struct {
		msg     *types.MsgRecvPacket
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
			suite.NoError(err, "Msg %d failed: %v", i, err)
		} else {
			suite.Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgRecvPacketGetSignBytes tests GetSignBytes for MsgRecvPacket
func (suite *TypesTestSuite) TestMsgRecvPacketGetSignBytes() {
	msg := types.NewMsgRecvPacket(packet, suite.proof, height, addr1)
	res := msg.GetSignBytes()

	expected := fmt.Sprintf(
		`{"packet":{"data":%s,"destination_channel":"testcpchannel","destination_port":"testcpport","sequence":"1","source_channel":"testchannel","source_port":"testportid","timeout_height":{"epoch_height":"100","epoch_number":"0"},"timeout_timestamp":"100"},"proof":"Co0BCi4KCmljczIzOmlhdmwSA0tFWRobChkKA0tFWRIFVkFMVUUaCwgBGAEgASoDAAICClsKDGljczIzOnNpbXBsZRIMaWF2bFN0b3JlS2V5Gj0KOwoMaWF2bFN0b3JlS2V5EiAcIiDXSHQRSvh/Wa07MYpTK0B4XtbaXtzxBED76xk0WhoJCAEYASABKgEA","proof_height":{"epoch_height":"1","epoch_number":"0"},"signer":"cosmos1w3jhxarpv3j8yvg4ufs4x"}`,
		string(msg.GetDataSignBytes()),
	)
	suite.Equal(expected, string(res))
}

// TestMsgRecvPacketGetSigners tests GetSigners for MsgRecvPacket
func (suite *TypesTestSuite) TestMsgRecvPacketGetSigners() {
	msg := types.NewMsgRecvPacket(packet, suite.proof, height, addr1)
	res := msg.GetSigners()

	expected := "[746573746164647231]"
	suite.Equal(expected, fmt.Sprintf("%v", res))
}

// TestMsgTimeout tests ValidateBasic for MsgTimeout
func (suite *TypesTestSuite) TestMsgTimeout() {
	testMsgs := []*types.MsgTimeout{
		types.NewMsgTimeout(packet, 1, suite.proof, height, addr),
		types.NewMsgTimeout(packet, 1, suite.proof, clienttypes.ZeroHeight(), addr),
		types.NewMsgTimeout(packet, 1, suite.proof, height, emptyAddr),
		types.NewMsgTimeout(packet, 1, emptyProof, height, addr),
		types.NewMsgTimeout(invalidPacket, 1, suite.proof, height, addr),
	}

	testCases := []struct {
		msg     *types.MsgTimeout
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

// TestMsgTimeoutOnClose tests ValidateBasic for MsgTimeoutOnClose
func (suite *TypesTestSuite) TestMsgTimeoutOnClose() {
	testCases := []struct {
		name    string
		msg     sdk.Msg
		expPass bool
	}{
		{"success", types.NewMsgTimeoutOnClose(packet, 1, suite.proof, suite.proof, height, addr), true},
		{"empty proof", types.NewMsgTimeoutOnClose(packet, 1, emptyProof, suite.proof, height, addr), false},
		{"empty proof close", types.NewMsgTimeoutOnClose(packet, 1, suite.proof, emptyProof, height, addr), false},
		{"proof height is zero", types.NewMsgTimeoutOnClose(packet, 1, suite.proof, suite.proof, clienttypes.ZeroHeight(), addr), false},
		{"signer address is empty", types.NewMsgTimeoutOnClose(packet, 1, suite.proof, suite.proof, height, emptyAddr), false},
		{"invalid packet", types.NewMsgTimeoutOnClose(invalidPacket, 1, suite.proof, suite.proof, height, addr), false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			err := tc.msg.ValidateBasic()

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestMsgAcknowledgement tests ValidateBasic for MsgAcknowledgement
func (suite *TypesTestSuite) TestMsgAcknowledgement() {
	testMsgs := []*types.MsgAcknowledgement{
		types.NewMsgAcknowledgement(packet, packet.GetData(), suite.proof, height, addr),
		types.NewMsgAcknowledgement(packet, packet.GetData(), suite.proof, clienttypes.ZeroHeight(), addr),
		types.NewMsgAcknowledgement(packet, packet.GetData(), suite.proof, height, emptyAddr),
		types.NewMsgAcknowledgement(packet, packet.GetData(), emptyProof, height, addr),
		types.NewMsgAcknowledgement(invalidPacket, packet.GetData(), suite.proof, height, addr),
	}

	testCases := []struct {
		msg     *types.MsgAcknowledgement
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
