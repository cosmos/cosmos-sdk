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
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

const (
	// valid constatns used for testing
	portid   = "testportid"
	chanid   = "channel-0"
	cpportid = "testcpport"
	cpchanid = "testcpchannel"

	version = "1.0"

	// invalid constants used for testing
	invalidPort      = "(invalidport1)"
	invalidShortPort = "p"
	invalidLongPort  = "invalidlongportinvalidlongportinvalidlongportidinvalidlongportidinvalid"

	invalidChannel      = "(invalidchannel1)"
	invalidShortChannel = "invalid"
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

	addr      = sdk.AccAddress("testaddr111111111111")
	emptyAddr sdk.AccAddress

	connHops             = []string{"testconnection"}
	invalidConnHops      = []string{"testconnection", "testconnection"}
	invalidShortConnHops = []string{invalidShortConnection}
	invalidLongConnHops  = []string{invalidLongConnection}
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

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	suite.Require().NoError(err)
	proof, err := app.AppCodec().MarshalBinaryBare(&merkleProof)
	suite.Require().NoError(err)

	suite.proof = proof
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

func (suite *TypesTestSuite) TestMsgChannelOpenInitValidateBasic() {
	counterparty := types.NewCounterparty(cpportid, cpchanid)
	tryOpenChannel := types.NewChannel(types.TRYOPEN, types.ORDERED, counterparty, connHops, version)

	testCases := []struct {
		name    string
		msg     *types.MsgChannelOpenInit
		expPass bool
	}{
		{"", types.NewMsgChannelOpenInit(portid, version, types.ORDERED, connHops, cpportid, addr), true},
		{"too short port id", types.NewMsgChannelOpenInit(invalidShortPort, version, types.ORDERED, connHops, cpportid, addr), false},
		{"too long port id", types.NewMsgChannelOpenInit(invalidLongPort, version, types.ORDERED, connHops, cpportid, addr), false},
		{"port id contains non-alpha", types.NewMsgChannelOpenInit(invalidPort, version, types.ORDERED, connHops, cpportid, addr), false},
		{"invalid channel order", types.NewMsgChannelOpenInit(portid, version, types.Order(3), connHops, cpportid, addr), false},
		{"connection hops more than 1 ", types.NewMsgChannelOpenInit(portid, version, types.ORDERED, invalidConnHops, cpportid, addr), false},
		{"too short connection id", types.NewMsgChannelOpenInit(portid, version, types.UNORDERED, invalidShortConnHops, cpportid, addr), false},
		{"too long connection id", types.NewMsgChannelOpenInit(portid, version, types.UNORDERED, invalidLongConnHops, cpportid, addr), false},
		{"connection id contains non-alpha", types.NewMsgChannelOpenInit(portid, version, types.UNORDERED, []string{invalidConnection}, cpportid, addr), false},
		{"", types.NewMsgChannelOpenInit(portid, "", types.UNORDERED, connHops, cpportid, addr), true},
		{"invalid counterparty port id", types.NewMsgChannelOpenInit(portid, version, types.UNORDERED, connHops, invalidPort, addr), false},
		{"channel not in INIT state", &types.MsgChannelOpenInit{portid, tryOpenChannel, addr.String()}, false},
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

func (suite *TypesTestSuite) TestMsgChannelOpenTryValidateBasic() {
	counterparty := types.NewCounterparty(cpportid, cpchanid)
	initChannel := types.NewChannel(types.INIT, types.ORDERED, counterparty, connHops, version)

	testCases := []struct {
		name    string
		msg     *types.MsgChannelOpenTry
		expPass bool
	}{
		{"", types.NewMsgChannelOpenTry(portid, chanid, version, types.ORDERED, connHops, cpportid, cpchanid, version, suite.proof, height, addr), true},
		{"too short port id", types.NewMsgChannelOpenTry(invalidShortPort, chanid, version, types.ORDERED, connHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"too long port id", types.NewMsgChannelOpenTry(invalidLongPort, chanid, version, types.ORDERED, connHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"port id contains non-alpha", types.NewMsgChannelOpenTry(invalidPort, chanid, version, types.ORDERED, connHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"too short channel id", types.NewMsgChannelOpenTry(portid, invalidShortChannel, version, types.ORDERED, connHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"too long channel id", types.NewMsgChannelOpenTry(portid, invalidLongChannel, version, types.ORDERED, connHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"channel id contains non-alpha", types.NewMsgChannelOpenTry(portid, invalidChannel, version, types.ORDERED, connHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"", types.NewMsgChannelOpenTry(portid, chanid, version, types.ORDERED, connHops, cpportid, cpchanid, "", suite.proof, height, addr), true},
		{"proof height is zero", types.NewMsgChannelOpenTry(portid, chanid, version, types.ORDERED, connHops, cpportid, cpchanid, version, suite.proof, clienttypes.ZeroHeight(), addr), false},
		{"invalid channel order", types.NewMsgChannelOpenTry(portid, chanid, version, types.Order(4), connHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"connection hops more than 1 ", types.NewMsgChannelOpenTry(portid, chanid, version, types.UNORDERED, invalidConnHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"too short connection id", types.NewMsgChannelOpenTry(portid, chanid, version, types.UNORDERED, invalidShortConnHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"too long connection id", types.NewMsgChannelOpenTry(portid, chanid, version, types.UNORDERED, invalidLongConnHops, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"connection id contains non-alpha", types.NewMsgChannelOpenTry(portid, chanid, version, types.UNORDERED, []string{invalidConnection}, cpportid, cpchanid, version, suite.proof, height, addr), false},
		{"", types.NewMsgChannelOpenTry(portid, chanid, "", types.UNORDERED, connHops, cpportid, cpchanid, version, suite.proof, height, addr), true},
		{"invalid counterparty port id", types.NewMsgChannelOpenTry(portid, chanid, version, types.UNORDERED, connHops, invalidPort, cpchanid, version, suite.proof, height, addr), false},
		{"invalid counterparty channel id", types.NewMsgChannelOpenTry(portid, chanid, version, types.UNORDERED, connHops, cpportid, invalidChannel, version, suite.proof, height, addr), false},
		{"empty proof", types.NewMsgChannelOpenTry(portid, chanid, version, types.UNORDERED, connHops, cpportid, cpchanid, version, emptyProof, height, addr), false},
		{"channel not in TRYOPEN state", &types.MsgChannelOpenTry{portid, chanid, initChannel, version, suite.proof, height, addr.String()}, false},
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

func (suite *TypesTestSuite) TestMsgChannelOpenAckValidateBasic() {
	testCases := []struct {
		name    string
		msg     *types.MsgChannelOpenAck
		expPass bool
	}{
		{"", types.NewMsgChannelOpenAck(portid, chanid, chanid, version, suite.proof, height, addr), true},
		{"too short port id", types.NewMsgChannelOpenAck(invalidShortPort, chanid, chanid, version, suite.proof, height, addr), false},
		{"too long port id", types.NewMsgChannelOpenAck(invalidLongPort, chanid, chanid, version, suite.proof, height, addr), false},
		{"port id contains non-alpha", types.NewMsgChannelOpenAck(invalidPort, chanid, chanid, version, suite.proof, height, addr), false},
		{"too short channel id", types.NewMsgChannelOpenAck(portid, invalidShortChannel, chanid, version, suite.proof, height, addr), false},
		{"too long channel id", types.NewMsgChannelOpenAck(portid, invalidLongChannel, chanid, version, suite.proof, height, addr), false},
		{"channel id contains non-alpha", types.NewMsgChannelOpenAck(portid, invalidChannel, chanid, version, suite.proof, height, addr), false},
		{"", types.NewMsgChannelOpenAck(portid, chanid, chanid, "", suite.proof, height, addr), true},
		{"empty proof", types.NewMsgChannelOpenAck(portid, chanid, chanid, version, emptyProof, height, addr), false},
		{"proof height is zero", types.NewMsgChannelOpenAck(portid, chanid, chanid, version, suite.proof, clienttypes.ZeroHeight(), addr), false},
		{"invalid counterparty channel id", types.NewMsgChannelOpenAck(portid, chanid, invalidShortChannel, version, suite.proof, height, addr), false},
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

func (suite *TypesTestSuite) TestMsgChannelOpenConfirmValidateBasic() {
	testCases := []struct {
		name    string
		msg     *types.MsgChannelOpenConfirm
		expPass bool
	}{
		{"", types.NewMsgChannelOpenConfirm(portid, chanid, suite.proof, height, addr), true},
		{"too short port id", types.NewMsgChannelOpenConfirm(invalidShortPort, chanid, suite.proof, height, addr), false},
		{"too long port id", types.NewMsgChannelOpenConfirm(invalidLongPort, chanid, suite.proof, height, addr), false},
		{"port id contains non-alpha", types.NewMsgChannelOpenConfirm(invalidPort, chanid, suite.proof, height, addr), false},
		{"too short channel id", types.NewMsgChannelOpenConfirm(portid, invalidShortChannel, suite.proof, height, addr), false},
		{"too long channel id", types.NewMsgChannelOpenConfirm(portid, invalidLongChannel, suite.proof, height, addr), false},
		{"channel id contains non-alpha", types.NewMsgChannelOpenConfirm(portid, invalidChannel, suite.proof, height, addr), false},
		{"empty proof", types.NewMsgChannelOpenConfirm(portid, chanid, emptyProof, height, addr), false},
		{"proof height is zero", types.NewMsgChannelOpenConfirm(portid, chanid, suite.proof, clienttypes.ZeroHeight(), addr), false},
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

func (suite *TypesTestSuite) TestMsgChannelCloseInitValidateBasic() {
	testCases := []struct {
		name    string
		msg     *types.MsgChannelCloseInit
		expPass bool
	}{
		{"", types.NewMsgChannelCloseInit(portid, chanid, addr), true},
		{"too short port id", types.NewMsgChannelCloseInit(invalidShortPort, chanid, addr), false},
		{"too long port id", types.NewMsgChannelCloseInit(invalidLongPort, chanid, addr), false},
		{"port id contains non-alpha", types.NewMsgChannelCloseInit(invalidPort, chanid, addr), false},
		{"too short channel id", types.NewMsgChannelCloseInit(portid, invalidShortChannel, addr), false},
		{"too long channel id", types.NewMsgChannelCloseInit(portid, invalidLongChannel, addr), false},
		{"channel id contains non-alpha", types.NewMsgChannelCloseInit(portid, invalidChannel, addr), false},
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

func (suite *TypesTestSuite) TestMsgChannelCloseConfirmValidateBasic() {
	testCases := []struct {
		name    string
		msg     *types.MsgChannelCloseConfirm
		expPass bool
	}{
		{"", types.NewMsgChannelCloseConfirm(portid, chanid, suite.proof, height, addr), true},
		{"too short port id", types.NewMsgChannelCloseConfirm(invalidShortPort, chanid, suite.proof, height, addr), false},
		{"too long port id", types.NewMsgChannelCloseConfirm(invalidLongPort, chanid, suite.proof, height, addr), false},
		{"port id contains non-alpha", types.NewMsgChannelCloseConfirm(invalidPort, chanid, suite.proof, height, addr), false},
		{"too short channel id", types.NewMsgChannelCloseConfirm(portid, invalidShortChannel, suite.proof, height, addr), false},
		{"too long channel id", types.NewMsgChannelCloseConfirm(portid, invalidLongChannel, suite.proof, height, addr), false},
		{"channel id contains non-alpha", types.NewMsgChannelCloseConfirm(portid, invalidChannel, suite.proof, height, addr), false},
		{"empty proof", types.NewMsgChannelCloseConfirm(portid, chanid, emptyProof, height, addr), false},
		{"proof height is zero", types.NewMsgChannelCloseConfirm(portid, chanid, suite.proof, clienttypes.ZeroHeight(), addr), false},
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

func (suite *TypesTestSuite) TestMsgRecvPacketType() {
	msg := types.NewMsgRecvPacket(packet, suite.proof, height, addr)

	suite.Equal("recv_packet", msg.Type())
}

func (suite *TypesTestSuite) TestMsgRecvPacketValidateBasic() {
	testCases := []struct {
		name    string
		msg     *types.MsgRecvPacket
		expPass bool
	}{
		{"success", types.NewMsgRecvPacket(packet, suite.proof, height, addr), true},
		{"proof height is zero", types.NewMsgRecvPacket(packet, suite.proof, clienttypes.ZeroHeight(), addr), false},
		{"proof contain empty proof", types.NewMsgRecvPacket(packet, emptyProof, height, addr), false},
		{"missing signer address", types.NewMsgRecvPacket(packet, suite.proof, height, emptyAddr), false},
		{"invalid packet", types.NewMsgRecvPacket(invalidPacket, suite.proof, height, addr), false},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			err := tc.msg.ValidateBasic()

			if tc.expPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
			}
		})
	}
}

func (suite *TypesTestSuite) TestMsgRecvPacketGetSigners() {
	msg := types.NewMsgRecvPacket(packet, suite.proof, height, addr)
	res := msg.GetSigners()

	expected := "[7465737461646472313131313131313131313131]"
	suite.Equal(expected, fmt.Sprintf("%v", res))
}

func (suite *TypesTestSuite) TestMsgTimeoutValidateBasic() {
	testCases := []struct {
		name    string
		msg     *types.MsgTimeout
		expPass bool
	}{
		{"success", types.NewMsgTimeout(packet, 1, suite.proof, height, addr), true},
		{"proof height must be > 0", types.NewMsgTimeout(packet, 1, suite.proof, clienttypes.ZeroHeight(), addr), false},
		{"seq 0", types.NewMsgTimeout(packet, 0, suite.proof, height, addr), false},
		{"missing signer address", types.NewMsgTimeout(packet, 1, suite.proof, height, emptyAddr), false},
		{"cannot submit an empty proof", types.NewMsgTimeout(packet, 1, emptyProof, height, addr), false},
		{"invalid packet", types.NewMsgTimeout(invalidPacket, 1, suite.proof, height, addr), false},
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

func (suite *TypesTestSuite) TestMsgTimeoutOnCloseValidateBasic() {
	testCases := []struct {
		name    string
		msg     sdk.Msg
		expPass bool
	}{
		{"success", types.NewMsgTimeoutOnClose(packet, 1, suite.proof, suite.proof, height, addr), true},
		{"seq 0", types.NewMsgTimeoutOnClose(packet, 0, suite.proof, suite.proof, height, addr), false},
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

func (suite *TypesTestSuite) TestMsgAcknowledgementValidateBasic() {
	testCases := []struct {
		name    string
		msg     *types.MsgAcknowledgement
		expPass bool
	}{
		{"success", types.NewMsgAcknowledgement(packet, packet.GetData(), suite.proof, height, addr), true},
		{"proof height must be > 0", types.NewMsgAcknowledgement(packet, packet.GetData(), suite.proof, clienttypes.ZeroHeight(), addr), false},
		{"empty ack", types.NewMsgAcknowledgement(packet, nil, suite.proof, height, addr), false},
		{"missing signer address", types.NewMsgAcknowledgement(packet, packet.GetData(), suite.proof, height, emptyAddr), false},
		{"cannot submit an empty proof", types.NewMsgAcknowledgement(packet, packet.GetData(), emptyProof, height, addr), false},
		{"invalid packet", types.NewMsgAcknowledgement(invalidPacket, packet.GetData(), suite.proof, height, addr), false},
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
