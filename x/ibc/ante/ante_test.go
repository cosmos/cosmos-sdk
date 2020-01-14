package ante_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypestm "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/ante"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// define constants used for testing
const (
	testChainID    = "test-chain-id"
	testClient     = "test-client"
	testClientType = clientexported.Tendermint

	testConnection = "testconnection"

	testChannelVersion = "1.0"
)

// define variables used for testing
type HandlerTestSuite struct {
	suite.Suite

	cdc    *codec.Codec
	ctx    sdk.Context
	app    *simapp.SimApp
	valSet *tmtypes.ValidatorSet
}

func (suite *HandlerTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.app = app

	privVal := tmtypes.NewMockPV()

	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	suite.createClient()
	suite.createConnection(connection.OPEN)
}

func (suite *HandlerTestSuite) createClient() {
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	consensusState := clienttypestm.ConsensusState{
		ChainID:          testChainID,
		Height:           uint64(commitID.Version),
		Root:             commitment.NewRoot(commitID.Hash),
		ValidatorSet:     suite.valSet,
		NextValidatorSet: suite.valSet,
	}

	_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClient, testClientType, consensusState)
	suite.NoError(err)
}

func (suite *HandlerTestSuite) updateClient() {
	// always commit and begin a new block on updateClient
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	state := clienttypestm.ConsensusState{
		ChainID: testChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}

	suite.app.IBCKeeper.ClientKeeper.SetConsensusState(suite.ctx, testClient, state)
	suite.app.IBCKeeper.ClientKeeper.SetVerifiedRoot(suite.ctx, testClient, state.GetHeight(), state.GetRoot())
}

func (suite *HandlerTestSuite) createConnection(state connection.State) {
	connection := connection.ConnectionEnd{
		State:    state,
		ClientID: testClient,
		Counterparty: connection.Counterparty{
			ClientID:     testClient,
			ConnectionID: testConnection,
			Prefix:       suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
		},
		Versions: connection.GetCompatibleVersions(),
	}

	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnection, connection)
}

func (suite *HandlerTestSuite) createChannel(portID string, chanID string, connID string, counterpartyPort string, counterpartyChan string, state channel.State, order channel.Order) {
	ch := channel.Channel{
		State:    state,
		Ordering: order,
		Counterparty: channel.Counterparty{
			PortID:    counterpartyPort,
			ChannelID: counterpartyChan,
		},
		ConnectionHops: []string{connID},
		Version:        testChannelVersion,
	}

	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, portID, chanID, ch)
}

func (suite *HandlerTestSuite) queryProof(key string) (proof commitment.Proof, height int64) {
	res := suite.app.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", ibctypes.StoreKey),
		Data:  []byte(key),
		Prove: true,
	})

	height = res.Height
	proof = commitment.Proof{
		Proof: res.Proof,
	}

	return
}

func (suite *HandlerTestSuite) newTx(msg sdk.Msg) sdk.Tx {
	return auth.StdTx{
		Msgs: []sdk.Msg{msg},
	}
}

func (suite *HandlerTestSuite) TestHandleMsgPacketOrdered() {
	handler := sdk.ChainAnteDecorators(ante.NewProofVerificationDecorator(
		suite.app.IBCKeeper.ClientKeeper,
		suite.app.IBCKeeper.ChannelKeeper,
	))

	packet := channel.NewPacket(newPacket(12345), 1, portid, chanid, cpportid, cpchanid)

	cctx, _ := suite.ctx.CacheContext()
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, packet.SourcePort, packet.SourceChannel, 1)
	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, packet.SourcePort, packet.SourceChannel, packet.Sequence, channeltypes.CommitPacket(packet.Data))
	msg := channel.NewMsgPacket(packet, nil, 0, addr1)
	_, err := handler(cctx, suite.newTx(msg), false)
	suite.Error(err, "%+v", err) // channel does not exist

	cctx, _ = suite.ctx.CacheContext()
	suite.createChannel(cpportid, cpchanid, testConnection, portid, chanid, channel.OPEN, channel.ORDERED)
	packetCommitmentPath := channel.PacketCommitmentPath(packet.SourcePort, packet.SourceChannel, packet.Sequence)
	proof, proofHeight := suite.queryProof(packetCommitmentPath)
	msg = channel.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)
	_, err = handler(cctx, suite.newTx(msg), false)
	suite.Error(err, "%+v", err) // invalid proof

	suite.updateClient()
	cctx, _ = suite.ctx.CacheContext()
	proof, proofHeight = suite.queryProof(packetCommitmentPath)
	msg = channel.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)
	_, err = handler(cctx, suite.newTx(msg), false)
	suite.Error(err, "%+v", err) // next recvseq not set

	proof, proofHeight = suite.queryProof(packetCommitmentPath)
	msg = channel.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, cpportid, cpchanid, 1)
	cctx, write := suite.ctx.CacheContext()

	for i := 0; i < 10; i++ {
		suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(cctx, cpportid, cpchanid, uint64(i))
		_, err := handler(cctx, suite.newTx(msg), false)
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
		suite.app.IBCKeeper.ClientKeeper,
		suite.app.IBCKeeper.ChannelKeeper,
	))

	// Not testing nonexist channel, invalid proof, nextseqsend, they are already tested in TestHandleMsgPacketOrdered

	var packet channeltypes.Packet
	for i := 0; i < 5; i++ {
		packet = channel.NewPacket(newPacket(uint64(i)), uint64(i), portid, chanid, cpportid, cpchanid)
		suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, packet.SourcePort, packet.SourceChannel, uint64(i), channeltypes.CommitPacket(packet.Data))
	}

	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, packet.SourcePort, packet.SourceChannel, uint64(10))

	suite.createChannel(cpportid, cpchanid, testConnection, portid, chanid, channel.OPEN, channel.UNORDERED)

	suite.updateClient()

	for i := 10; i >= 0; i-- {
		cctx, write := suite.ctx.CacheContext()
		packet = channel.NewPacket(newPacket(uint64(i)), uint64(i), portid, chanid, cpportid, cpchanid)
		packetCommitmentPath := channel.PacketCommitmentPath(packet.SourcePort, packet.SourceChannel, uint64(i))
		proof, proofHeight := suite.queryProof(packetCommitmentPath)
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

var _ channelexported.PacketDataI = packetT{}

type packetT struct {
	Data uint64
}

func (packet packetT) GetBytes() []byte {
	return []byte(fmt.Sprintf("%d", packet.Data))
}

func (packetT) GetTimeoutHeight() uint64 {
	return 100
}

func (packetT) ValidateBasic() error {
	return nil
}

func (packetT) Type() string {
	return "valid"
}

func newPacket(data uint64) packetT {
	return packetT{data}
}

// define variables used for testing
var (
	addr1 = sdk.AccAddress("testaddr1")

	portid   = "testportid"
	chanid   = "testchannel"
	cpportid = "testcpport"
	cpchanid = "testcpchannel"
)
