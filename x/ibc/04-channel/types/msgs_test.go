package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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

type MsgTestSuite struct {
	suite.Suite

	proof commitment.Proof
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

	suite.proof = commitment.Proof{Proof: res.Proof}
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
		NewMsgChannelOpenTry("testportid", "testchannel", "1.0", ORDERED, connHops, "testcpport", "testcpchannel", "1.0", nil, 1, addr),                              // empty suite.proof
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
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

// TestMsgChannelOpenAck tests ValidateBasic for MsgChannelOpenAck
func (suite *MsgTestSuite) TestMsgChannelOpenAck() {
	testMsgs := []MsgChannelOpenAck{
		NewMsgChannelOpenAck("testportid", "testchannel", "1.0", suite.proof, 1, addr),                  // valid msg
		NewMsgChannelOpenAck(invalidShortPort, "testchannel", "1.0", suite.proof, 1, addr),              // too short port id
		NewMsgChannelOpenAck(invalidLongPort, "testchannel", "1.0", suite.proof, 1, addr),               // too long port id
		NewMsgChannelOpenAck(invalidPort, "testchannel", "1.0", suite.proof, 1, addr),                   // port id contains non-alpha
		NewMsgChannelOpenAck("testportid", invalidShortChannel, "1.0", suite.proof, 1, addr),            // too short channel id
		NewMsgChannelOpenAck("testportid", invalidLongChannel, "1.0", suite.proof, 1, addr),             // too long channel id
		NewMsgChannelOpenAck("testportid", invalidChannel, "1.0", suite.proof, 1, addr),                 // channel id contains non-alpha
		NewMsgChannelOpenAck("testportid", "testchannel", "", suite.proof, 1, addr),                     // empty counterparty version
		NewMsgChannelOpenAck("testportid", "testchannel", "1.0", nil, 1, addr),                          // empty proof
		NewMsgChannelOpenAck("testportid", "testchannel", "1.0", commitment.Proof{Proof: nil}, 1, addr), // empty proof
		NewMsgChannelOpenAck("testportid", "testchannel", "1.0", suite.proof, 0, addr),                  // proof height is zero
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
		{testMsgs[9], false, "empty proof"},
		{testMsgs[10], false, "proof height is zero"},
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
		NewMsgChannelOpenConfirm("testportid", "testchannel", suite.proof, 1, addr),                  // valid msg
		NewMsgChannelOpenConfirm(invalidShortPort, "testchannel", suite.proof, 1, addr),              // too short port id
		NewMsgChannelOpenConfirm(invalidLongPort, "testchannel", suite.proof, 1, addr),               // too long port id
		NewMsgChannelOpenConfirm(invalidPort, "testchannel", suite.proof, 1, addr),                   // port id contains non-alpha
		NewMsgChannelOpenConfirm("testportid", invalidShortChannel, suite.proof, 1, addr),            // too short channel id
		NewMsgChannelOpenConfirm("testportid", invalidLongChannel, suite.proof, 1, addr),             // too long channel id
		NewMsgChannelOpenConfirm("testportid", invalidChannel, suite.proof, 1, addr),                 // channel id contains non-alpha
		NewMsgChannelOpenConfirm("testportid", "testchannel", nil, 1, addr),                          // empty proof
		NewMsgChannelOpenConfirm("testportid", "testchannel", commitment.Proof{Proof: nil}, 1, addr), // empty proof
		NewMsgChannelOpenConfirm("testportid", "testchannel", suite.proof, 0, addr),                  // proof height is zero
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
		NewMsgChannelCloseConfirm("testportid", "testchannel", suite.proof, 1, addr),                  // valid msg
		NewMsgChannelCloseConfirm(invalidShortPort, "testchannel", suite.proof, 1, addr),              // too short port id
		NewMsgChannelCloseConfirm(invalidLongPort, "testchannel", suite.proof, 1, addr),               // too long port id
		NewMsgChannelCloseConfirm(invalidPort, "testchannel", suite.proof, 1, addr),                   // port id contains non-alpha
		NewMsgChannelCloseConfirm("testportid", invalidShortChannel, suite.proof, 1, addr),            // too short channel id
		NewMsgChannelCloseConfirm("testportid", invalidLongChannel, suite.proof, 1, addr),             // too long channel id
		NewMsgChannelCloseConfirm("testportid", invalidChannel, suite.proof, 1, addr),                 // channel id contains non-alpha
		NewMsgChannelCloseConfirm("testportid", "testchannel", nil, 1, addr),                          // empty proof
		NewMsgChannelCloseConfirm("testportid", "testchannel", commitment.Proof{Proof: nil}, 1, addr), // empty proof
		NewMsgChannelCloseConfirm("testportid", "testchannel", suite.proof, 0, addr),                  // proof height is zero
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
