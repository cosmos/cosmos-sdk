package ante_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/ante"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

// define constants used for testing
const (
	testClientIDA  = "testclientida"
	testClientIDB  = "testclientidb"
	testConnection = "testconnection"
)

// define variables used for testing
var (
	addr1 = sdk.AccAddress("testaddr1")

	portid   = "testportid"
	chanid   = "testchannel"
	cpportid = "testcpport"
	cpchanid = "testcpchannel"
)

type HandlerTestSuite struct {
	suite.Suite

	cdc *codec.Codec

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.chainA = ibctesting.NewTestChain(testClientIDA)
	suite.chainB = ibctesting.NewTestChain(testClientIDB)

	suite.cdc = suite.chainA.App.Codec()

	// create client and connection during setups
	suite.chainA.CreateClient(suite.chainB)
	suite.chainB.CreateClient(suite.chainA)
	suite.chainA.CreateConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
	suite.chainB.CreateConnection(testConnection, testConnection, testClientIDA, testClientIDB, connectiontypes.OPEN)
}

func (suite *HandlerTestSuite) newTx(msg sdk.Msg) sdk.Tx {
	return authtypes.StdTx{
		Msgs: []sdk.Msg{msg},
	}
}

func (suite *HandlerTestSuite) TestHandleMsgPacketOrdered() {
	handler := sdk.ChainAnteDecorators(ante.NewProofVerificationDecorator(
		suite.chainA.App.IBCKeeper.ClientKeeper,
		suite.chainA.App.IBCKeeper.ChannelKeeper,
	))

	packet := channel.NewPacket(newPacket(12345).GetData(), 1, portid, chanid, cpportid, cpchanid, 100, 0)

	ctx := suite.chainA.GetContext()
	cctx, _ := ctx.CacheContext()
	// suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctx, packet.SourcePort, packet.SourceChannel, 1)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), packet.SourcePort, packet.SourceChannel, packet.Sequence, channeltypes.CommitPacket(packet))
	msg := channel.NewMsgPacket(packet, commitmenttypes.MerkleProof{}, 0, addr1)
	_, err := handler(cctx, suite.newTx(msg), false)
	suite.Error(err, "%+v", err) // channel does not exist

	suite.chainA.CreateChannel(cpportid, cpchanid, portid, chanid, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
	suite.chainB.CreateChannel(portid, chanid, cpportid, cpchanid, channeltypes.OPEN, channeltypes.ORDERED, testConnection)

	packetCommitmentPath := host.PacketCommitmentPath(packet.SourcePort, packet.SourceChannel, packet.Sequence)
	proof, proofHeight := suite.chainB.QueryProof([]byte(packetCommitmentPath))
	msg = channel.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)
	_, err = handler(cctx, suite.newTx(msg), false)
	suite.Error(err, "%+v", err) // invalid proof

	suite.chainA.UpdateClient(suite.chainB)
	// // commit chainA to flush to IAVL so we can get proof
	// suite.chainA.App.Commit()
	// suite.chainA.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.chainA.App.LastBlockHeight() + 1, Time: suite.chainA.Header.Time}})
	// ctx = suite.chainA.GetContext()

	proof, proofHeight = suite.chainB.QueryProof([]byte(packetCommitmentPath))
	msg = channel.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)

	for i := 0; i < 10; i++ {
		cctx, write := suite.chainA.GetContext().CacheContext()
		suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(cctx, cpportid, cpchanid, uint64(i))
		_, err := handler(cctx, suite.newTx(msg), false)
		if err == nil {
			// retrieve channelCapability from scopedIBCKeeper and pass into PacketExecuted
			chanCap, ok := suite.chainA.App.ScopedIBCKeeper.GetCapability(cctx, host.ChannelCapabilityPath(
				packet.GetDestPort(), packet.GetDestChannel()),
			)
			suite.Require().True(ok, "could not retrieve capability")
			err = suite.chainA.App.IBCKeeper.ChannelKeeper.PacketExecuted(cctx, chanCap, packet, packet.Data)
		}
		if i == 1 {
			suite.NoError(err, "%d", i) // successfully executed
			write()
		} else {
			suite.Error(err, "%d", i) // wrong incoming sequence
		}
	}
}

func (suite *HandlerTestSuite) TestHandleMsgPacketUnordered() {
	handler := sdk.ChainAnteDecorators(ante.NewProofVerificationDecorator(
		suite.chainA.App.IBCKeeper.ClientKeeper,
		suite.chainA.App.IBCKeeper.ChannelKeeper,
	))

	// Not testing nonexist channel, invalid proof, nextseqsend, they are already tested in TestHandleMsgPacketOrdered

	var packet channeltypes.Packet
	for i := 0; i < 5; i++ {
		packet = channel.NewPacket(newPacket(uint64(i)).GetData(), uint64(i), portid, chanid, cpportid, cpchanid, 100, 0)
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), packet.SourcePort, packet.SourceChannel, uint64(i), channeltypes.CommitPacket(packet))
	}

	// suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), packet.SourcePort, packet.SourceChannel, uint64(10))

	suite.chainA.CreateChannel(cpportid, cpchanid, portid, chanid, channeltypes.OPEN, channeltypes.UNORDERED, testConnection)

	suite.chainA.UpdateClient(suite.chainB)

	for i := 10; i >= 0; i-- {
		cctx, write := suite.chainA.GetContext().CacheContext()
		packet = channel.NewPacket(newPacket(uint64(i)).GetData(), uint64(i), portid, chanid, cpportid, cpchanid, 100, 0)
		packetCommitmentPath := host.PacketCommitmentPath(packet.SourcePort, packet.SourceChannel, uint64(i))
		proof, proofHeight := suite.chainB.QueryProof([]byte(packetCommitmentPath))
		msg := channel.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)
		_, err := handler(cctx, suite.newTx(msg), false)
		if i < 5 {
			suite.NoError(err, "%d", i) // successfully executed
			write()
		} else {
			suite.Error(err, "%d", i) // wrong incoming sequence
		}
	}
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

type packetT struct {
	Data uint64
}

func (packet packetT) GetData() []byte {
	return []byte(fmt.Sprintf("%d", packet.Data))
}

func newPacket(data uint64) packetT {
	return packetT{data}
}
