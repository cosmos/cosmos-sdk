package keeper_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// define constants used for testing
const (
	testClientType = clientexported.Tendermint

	testClientID1     = "testclientidone"
	testConnectionID1 = "connectionidone"

	testClientID2     = "testclientidtwo"
	testConnectionID2 = "connectionidtwo"

	testPort1 = "firstport"
	testPort2 = "secondport"
	testPort3 = "thirdport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
	testChannel3 = "thirdchannel"

	testChannelOrder   = exported.ORDERED
	testChannelVersion = "1.0"

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
)

type KeeperTestSuite struct {
	suite.Suite

	cdc    *codec.Codec
	ctx    sdk.Context
	app    *simapp.SimApp
	valSet *tmtypes.ValidatorSet
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.app = app

	privVal := tmtypes.NewMockPV()

	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
}

func (suite *KeeperTestSuite) TestSetChannel() {
	_, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort1, testChannel1)
	suite.False(found)

	channel := types.Channel{
		State:    exported.OPEN,
		Ordering: testChannelOrder,
		Counterparty: types.Counterparty{
			PortID:    testPort2,
			ChannelID: testChannel2,
		},
		ConnectionHops: []string{testConnectionID1},
		Version:        testChannelVersion,
	}
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, testPort1, testChannel1, channel)

	storedChannel, found := suite.app.IBCKeeper.ChannelKeeper.GetChannel(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(channel, storedChannel)
}

func (suite KeeperTestSuite) TestGetAllChannels() {
	// Channel (Counterparty): A(C) -> C(B) -> B(A)
	counterparty1 := types.NewCounterparty(testPort1, testChannel1)
	counterparty2 := types.NewCounterparty(testPort2, testChannel2)
	counterparty3 := types.NewCounterparty(testPort3, testChannel3)

	channel1 := types.Channel{
		State:          exported.INIT,
		Ordering:       testChannelOrder,
		Counterparty:   counterparty3,
		ConnectionHops: []string{testConnectionID1},
		Version:        testChannelVersion,
	}

	channel2 := types.Channel{
		State:          exported.INIT,
		Ordering:       testChannelOrder,
		Counterparty:   counterparty1,
		ConnectionHops: []string{testConnectionID1},
		Version:        testChannelVersion,
	}

	channel3 := types.Channel{
		State:          exported.CLOSED,
		Ordering:       testChannelOrder,
		Counterparty:   counterparty2,
		ConnectionHops: []string{testConnectionID1},
		Version:        testChannelVersion,
	}

	expChannels := []types.Channel{channel1, channel2, channel3}

	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, testPort1, testChannel1, expChannels[0])
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, testPort2, testChannel2, expChannels[1])
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, testPort3, testChannel3, expChannels[2])

	channels := suite.app.IBCKeeper.ChannelKeeper.GetAllChannels(suite.ctx)
	suite.Require().Len(channels, len(expChannels))
	suite.Require().Equal(expChannels, channels)
}

func (suite *KeeperTestSuite) TestSetChannelCapability() {
	_, found := suite.app.IBCKeeper.ChannelKeeper.GetChannelCapability(suite.ctx, testPort1, testChannel1)
	suite.False(found)

	channelCap := "test-channel-capability"
	suite.app.IBCKeeper.ChannelKeeper.SetChannelCapability(suite.ctx, testPort1, testChannel1, channelCap)

	storedChannelCap, found := suite.app.IBCKeeper.ChannelKeeper.GetChannelCapability(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(channelCap, storedChannelCap)
}

func (suite *KeeperTestSuite) TestSetSequence() {
	_, found := suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(suite.ctx, testPort1, testChannel1)
	suite.False(found)

	_, found = suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(suite.ctx, testPort1, testChannel1)
	suite.False(found)

	nextSeqSend, nextSeqRecv := uint64(10), uint64(10)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, nextSeqSend)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort1, testChannel1, nextSeqRecv)

	storedNextSeqSend, found := suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqSend, storedNextSeqSend)

	storedNextSeqRecv, found := suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqRecv, storedNextSeqRecv)
}

func (suite *KeeperTestSuite) TestPackageCommitment() {
	seq := uint64(10)
	storedCommitment := suite.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.ctx, testPort1, testChannel1, seq)
	suite.Equal([]byte(nil), storedCommitment)

	commitment := []byte("commitment")
	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, seq, commitment)

	storedCommitment = suite.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.ctx, testPort1, testChannel1, seq)
	suite.Equal(commitment, storedCommitment)
}

func (suite *KeeperTestSuite) TestSetPacketAcknowledgement() {
	seq := uint64(10)

	storedAckHash, found := suite.app.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(suite.ctx, testPort1, testChannel1, seq)
	suite.False(found)
	suite.Nil(storedAckHash)

	ackHash := []byte("ackhash")
	suite.app.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.ctx, testPort1, testChannel1, seq, ackHash)

	storedAckHash, found = suite.app.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(suite.ctx, testPort1, testChannel1, seq)
	suite.True(found)
	suite.Equal(ackHash, storedAckHash)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) createClient(clientID string) {
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{Height: suite.app.LastBlockHeight()})

	consensusState := tendermint.ConsensusState{
		Root:         commitment.NewRoot(commitID.Hash),
		ValidatorSet: suite.valSet,
	}

	_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, clientID, testClientType, consensusState, trustingPeriod, ubdPeriod)
	suite.Require().NoError(err)
}

// nolint: unused
func (suite *KeeperTestSuite) updateClient() {
	// always commit and begin a new block on updateClient
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	height := suite.app.LastBlockHeight() + 1
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: height}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{Height: suite.app.LastBlockHeight()})

	state := tendermint.ConsensusState{
		Root: commitment.NewRoot(commitID.Hash),
	}

	suite.app.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.ctx, testClientID1, uint64(height-1), state)
	csi, _ := suite.app.IBCKeeper.ClientKeeper.GetClientState(suite.ctx, testClientID1)
	cs, _ := csi.(tendermint.ClientState)
	cs.LatestHeight = uint64(height - 1)
	suite.app.IBCKeeper.ClientKeeper.SetClientState(suite.ctx, cs)
}

func (suite *KeeperTestSuite) createConnection(
	connID, counterpartyConnID, clientID, counterpartyClientID string,
	state connectionexported.State,
) connectiontypes.ConnectionEnd {
	counterparty := connectiontypes.NewCounterparty(counterpartyClientID, counterpartyConnID, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix())
	connection := connectiontypes.ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     connectiontypes.GetCompatibleVersions(),
	}
	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, connID, connection)
	return connection
}

func (suite *KeeperTestSuite) createChannel(
	portID, channelID, counterpartyPortID, counterpartyChannelID string,
	state exported.State, order exported.Order, connectionID string,
) types.Channel {
	counterparty := types.NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := types.NewChannel(state, order, counterparty,
		[]string{connectionID}, "1.0",
	)
	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, portID, channelID, channel)
	return channel
}

// nolint: unused
func (suite *KeeperTestSuite) queryProof(key []byte) (commitment.Proof, int64) {
	res := suite.app.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", ibctypes.StoreKey),
		Height: suite.app.LastBlockHeight(),
		Data:   key,
		Prove:  true,
	})

	proof := commitment.Proof{
		Proof: res.Proof,
	}

	return proof, res.Height
}

// Mocked types
// TODO: fix tests and replace for real proofs

var (
	_ commitment.ProofI = validProof{}
	_ commitment.ProofI = invalidProof{}
)

type (
	validProof   struct{}
	invalidProof struct{}
)

func (validProof) GetCommitmentType() commitment.Type {
	return commitment.Merkle
}

func (validProof) VerifyMembership(
	root commitment.RootI, path commitment.PathI, value []byte) error {
	return nil
}

func (validProof) VerifyNonMembership(root commitment.RootI, path commitment.PathI) error {
	return nil
}

func (validProof) ValidateBasic() error {
	return nil
}

func (validProof) IsEmpty() bool {
	return false
}

func (invalidProof) GetCommitmentType() commitment.Type {
	return commitment.Merkle
}

func (invalidProof) VerifyMembership(
	root commitment.RootI, path commitment.PathI, value []byte) error {
	return errors.New("proof failed")
}

func (invalidProof) VerifyNonMembership(root commitment.RootI, path commitment.PathI) error {
	return errors.New("proof failed")
}

func (invalidProof) ValidateBasic() error {
	return errors.New("invalid proof")
}

func (invalidProof) IsEmpty() bool {
	return true
}
