package ante_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/ante"
)

type HandlerTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *HandlerTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
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

	packet := channeltypes.NewPacket(newPacket(12345).GetData(), 1, portid, chanid, cpportid, cpchanid, 100, 0)

	ctx := suite.chainA.GetContext()
	cctx, _ := ctx.CacheContext()
	// suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctx, packet.SourcePort, packet.SourceChannel, 1)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), packet.SourcePort, packet.SourceChannel, packet.Sequence, channeltypes.CommitPacket(packet))
	msg := channeltypes.NewMsgPacket(packet, []byte{}, 0, addr1)
	_, err := handler(cctx, suite.newTx(msg), false)
	suite.Error(err, "%+v", err) // channel does not exist

	suite.chainA.createChannel(cpportid, cpchanid, portid, chanid, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
	suite.chainB.createChannel(portid, chanid, cpportid, cpchanid, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
	ctx = suite.chainA.GetContext()
	packetCommitmentPath := host.PacketCommitmentPath(packet.SourcePort, packet.SourceChannel, packet.Sequence)
	proof, proofHeight := queryProof(suite.chainB, packetCommitmentPath)
	msg = channeltypes.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)
	_, err = handler(cctx, suite.newTx(msg), false)
	suite.Error(err, "%+v", err) // invalid proof

	suite.chainA.updateClient(suite.chainB)
	// // commit chainA to flush to IAVL so we can get proof
	// suite.chainA.App.Commit()
	// suite.chainA.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.chainA.App.LastBlockHeight() + 1, Time: suite.chainA.Header.Time}})
	// ctx = suite.chainA.GetContext()

	proof, proofHeight = queryProof(suite.chainB, packetCommitmentPath)
	msg = channeltypes.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)

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
		packet = channeltypes.NewPacket(ibctesting.TestHash, uint64(i), portid, chanid, cpportid, cpchanid, 100, 0)
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainB.GetContext(), packet.SourcePort, packet.SourceChannel, uint64(i), channeltypes.CommitPacket(packet))
	}

	// suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.chainA.GetContext(), packet.SourcePort, packet.SourceChannel, uint64(10))

	suite.chainA.createChannel(cpportid, cpchanid, portid, chanid, channeltypes.OPEN, channeltypes.UNORDERED, testConnection)

	suite.chainA.updateClient(suite.chainB)

	for i := 10; i >= 0; i-- {
		cctx, write := suite.chainA.GetContext().CacheContext()
		packet = channeltypes.NewPacket(newPacket(uint64(i)).GetData(), uint64(i), portid, chanid, cpportid, cpchanid, 100, 0)
		packetCommitmentPath := host.PacketCommitmentPath(packet.SourcePort, packet.SourceChannel, uint64(i))
		proof, proofHeight := queryProof(suite.chainB, packetCommitmentPath)
		msg := channeltypes.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)
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
