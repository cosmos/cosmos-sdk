package keeper_test

import (
	"bytes"
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
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// define constants used for testing
const (
	testClientIDA     = "testclientida"
	testConnectionIDA = "connectionidatob"

	testClientIDB     = "testclientidb"
	testConnectionIDB = "connectionidbtoa"

	testPort1 = "firstport"
	testPort2 = "secondport"
	testPort3 = "thirdport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
	testChannel3 = "thirdchannel"

	testChannelOrder   = ibctypes.ORDERED
	testChannelVersion = "1.0"

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10

	timeoutHeight            = 100
	timeoutTimestamp         = 100
	disabledTimeoutTimestamp = 0
	disabledTimeoutHeight    = 0
)

type KeeperTestSuite struct {
	suite.Suite

	cdc *codec.Codec

	chainA *TestChain
	chainB *TestChain
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.chainA = NewTestChain(testClientIDA)
	suite.chainB = NewTestChain(testClientIDB)

	suite.cdc = suite.chainA.App.Codec()
}

func (suite *KeeperTestSuite) TestSetChannel() {
	ctx := suite.chainB.GetContext()
	_, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetChannel(ctx, testPort1, testChannel1)
	suite.False(found)

	counterparty2 := types.NewCounterparty(testPort2, testChannel2)
	channel := types.NewChannel(
		ibctypes.INIT, testChannelOrder,
		counterparty2, []string{testConnectionIDA}, testChannelVersion,
	)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, testPort1, testChannel1, channel)

	storedChannel, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetChannel(ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(channel, storedChannel)
}

func (suite KeeperTestSuite) TestGetAllChannels() {
	// Channel (Counterparty): A(C) -> C(B) -> B(A)
	counterparty1 := types.NewCounterparty(testPort1, testChannel1)
	counterparty2 := types.NewCounterparty(testPort2, testChannel2)
	counterparty3 := types.NewCounterparty(testPort3, testChannel3)

	channel1 := types.NewChannel(
		ibctypes.INIT, testChannelOrder,
		counterparty3, []string{testConnectionIDA}, testChannelVersion,
	)
	channel2 := types.NewChannel(
		ibctypes.INIT, testChannelOrder,
		counterparty1, []string{testConnectionIDA}, testChannelVersion,
	)
	channel3 := types.NewChannel(
		ibctypes.CLOSED, testChannelOrder,
		counterparty2, []string{testConnectionIDA}, testChannelVersion,
	)

	expChannels := []types.IdentifiedChannel{
		types.NewIdentifiedChannel(testPort1, testChannel1, channel1),
		types.NewIdentifiedChannel(testPort2, testChannel2, channel2),
		types.NewIdentifiedChannel(testPort3, testChannel3, channel3),
	}

	ctx := suite.chainB.GetContext()

	suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, testPort1, testChannel1, channel1)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, testPort2, testChannel2, channel2)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, testPort3, testChannel3, channel3)

	channels := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllChannels(ctx)
	suite.Require().Len(channels, len(expChannels))
	suite.Require().Equal(expChannels, channels)
}

func (suite KeeperTestSuite) TestGetAllSequences() {
	seq1 := types.NewPacketSequence(testPort1, testChannel1, 1)
	seq2 := types.NewPacketSequence(testPort2, testChannel2, 2)

	expSeqs := []types.PacketSequence{seq1, seq2}

	ctx := suite.chainB.GetContext()

	for _, seq := range expSeqs {
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctx, seq.PortID, seq.ChannelID, seq.Sequence)
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(ctx, seq.PortID, seq.ChannelID, seq.Sequence)
	}

	sendSeqs := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketSendSeqs(ctx)
	recvSeqs := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketRecvSeqs(ctx)
	suite.Require().Len(sendSeqs, 2)
	suite.Require().Len(recvSeqs, 2)

	suite.Require().Equal(expSeqs, sendSeqs)
	suite.Require().Equal(expSeqs, recvSeqs)
}

func (suite KeeperTestSuite) TestGetAllCommitmentsAcks() {
	ack1 := types.NewPacketAckCommitment(testPort1, testChannel1, 1, []byte("ack"))
	ack2 := types.NewPacketAckCommitment(testPort1, testChannel1, 2, []byte("ack"))
	comm1 := types.NewPacketAckCommitment(testPort1, testChannel1, 1, []byte("hash"))
	comm2 := types.NewPacketAckCommitment(testPort1, testChannel1, 2, []byte("hash"))

	expAcks := []types.PacketAckCommitment{ack1, ack2}
	expCommitments := []types.PacketAckCommitment{comm1, comm2}

	ctx := suite.chainB.GetContext()

	for i := 0; i < 2; i++ {
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctx, expAcks[i].PortID, expAcks[i].ChannelID, expAcks[i].Sequence, expAcks[i].Hash)
		suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctx, expCommitments[i].PortID, expCommitments[i].ChannelID, expCommitments[i].Sequence, expCommitments[i].Hash)
	}

	acks := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketAcks(ctx)
	commitments := suite.chainB.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitments(ctx)
	suite.Require().Len(acks, 2)
	suite.Require().Len(commitments, 2)

	suite.Require().Equal(expAcks, acks)
	suite.Require().Equal(expCommitments, commitments)
}

func (suite *KeeperTestSuite) TestSetSequence() {
	ctx := suite.chainB.GetContext()
	_, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, testPort1, testChannel1)
	suite.False(found)

	_, found = suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(ctx, testPort1, testChannel1)
	suite.False(found)

	nextSeqSend, nextSeqRecv := uint64(10), uint64(10)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctx, testPort1, testChannel1, nextSeqSend)
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(ctx, testPort1, testChannel1, nextSeqRecv)

	storedNextSeqSend, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqSend, storedNextSeqSend)

	storedNextSeqRecv, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqRecv, storedNextSeqRecv)
}

func (suite *KeeperTestSuite) TestPackageCommitment() {
	ctx := suite.chainB.GetContext()
	seq := uint64(10)
	storedCommitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(ctx, testPort1, testChannel1, seq)
	suite.Equal([]byte(nil), storedCommitment)

	commitment := []byte("commitment")
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(ctx, testPort1, testChannel1, seq, commitment)

	storedCommitment = suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(ctx, testPort1, testChannel1, seq)
	suite.Equal(commitment, storedCommitment)
}

func (suite *KeeperTestSuite) TestSetPacketAcknowledgement() {
	ctx := suite.chainB.GetContext()
	seq := uint64(10)

	storedAckHash, found := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(ctx, testPort1, testChannel1, seq)
	suite.False(found)
	suite.Nil(storedAckHash)

	ackHash := []byte("ackhash")
	suite.chainB.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctx, testPort1, testChannel1, seq, ackHash)

	storedAckHash, found = suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(ctx, testPort1, testChannel1, seq)
	suite.True(found)
	suite.Equal(ackHash, storedAckHash)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func commitNBlocks(chain *TestChain, n int) {
	for i := 0; i < n; i++ {
		chain.App.Commit()
		chain.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: chain.App.LastBlockHeight() + 1}})
	}
}

// commit current block and start the next block with the provided time
func commitBlockWithNewTimestamp(chain *TestChain, timestamp int64) {
	chain.App.Commit()
	chain.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: chain.App.LastBlockHeight() + 1, Time: time.Unix(timestamp, 0)}})
}

// nolint: unused
func queryProof(chain *TestChain, key []byte) (commitmenttypes.MerkleProof, uint64) {
	res := chain.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", ibctypes.StoreKey),
		Height: chain.App.LastBlockHeight(),
		Data:   key,
		Prove:  true,
	})

	proof := commitmenttypes.MerkleProof{
		Proof: res.Proof,
	}

	return proof, uint64(res.Height)
}

type TestChain struct {
	ClientID string
	App      *simapp.SimApp
	Header   ibctmtypes.Header
	Vals     *tmtypes.ValidatorSet
	Signers  []tmtypes.PrivValidator
}

func NewTestChain(clientID string) *TestChain {
	privVal := tmtypes.NewMockPV()

	pubKey, err := privVal.GetPubKey()
	if err != nil {
		panic(err)
	}

	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := []tmtypes.PrivValidator{privVal}
	now := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	header := ibctmtypes.CreateTestHeader(clientID, 1, now, valSet, signers)

	return &TestChain{
		ClientID: clientID,
		App:      simapp.Setup(false),
		Header:   header,
		Vals:     valSet,
		Signers:  signers,
	}
}

// Creates simple context for testing purposes
func (chain *TestChain) GetContext() sdk.Context {
	return chain.App.BaseApp.NewContext(false, abci.Header{ChainID: chain.Header.SignedHeader.Header.ChainID, Height: int64(chain.Header.GetHeight())})
}

// createClient will create a client for clientChain on targetChain
func (chain *TestChain) CreateClient(client *TestChain) error {
	client.Header = nextHeader(client)
	// Commit and create a new block on appTarget to get a fresh CommitID
	client.App.Commit()
	commitID := client.App.LastCommitID()
	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: int64(client.Header.GetHeight()), Time: client.Header.Time}})

	// Set HistoricalInfo on client chain after Commit
	ctxClient := client.GetContext()
	validator := staking.NewValidator(
		sdk.ValAddress(client.Vals.Validators[0].Address), client.Vals.Validators[0].PubKey, staking.Description{},
	)
	validator.Status = sdk.Bonded
	validator.Tokens = sdk.NewInt(1000000) // get one voting power
	validators := []staking.Validator{validator}
	histInfo := staking.HistoricalInfo{
		Header: abci.Header{
			AppHash: commitID.Hash,
		},
		Valset: validators,
	}
	client.App.StakingKeeper.SetHistoricalInfo(ctxClient, int64(client.Header.GetHeight()), histInfo)

	// Create target ctx
	ctxTarget := chain.GetContext()

	// create client
	clientState, err := ibctmtypes.Initialize(client.ClientID, trustingPeriod, ubdPeriod, maxClockDrift, client.Header)
	if err != nil {
		return err
	}
	_, err = chain.App.IBCKeeper.ClientKeeper.CreateClient(ctxTarget, clientState, client.Header.ConsensusState())
	if err != nil {
		return err
	}
	return nil

	// _, _, err := simapp.SignCheckDeliver(
	// 	suite.T(),
	// 	suite.cdc,
	// 	suite.app.BaseApp,
	// 	ctx.BlockHeader(),
	// 	[]sdk.Msg{clienttypes.NewMsgCreateClient(clientID, clientexported.ClientTypeTendermint, consState, accountAddress)},
	// 	[]uint64{baseAccount.GetAccountNumber()},
	// 	[]uint64{baseAccount.GetSequence()},
	// 	true, true, accountPrivKey,
	// )
}

func (chain *TestChain) updateClient(client *TestChain) {
	// Create target ctx
	ctxTarget := chain.GetContext()

	// if clientState does not already exist, return without updating
	_, found := chain.App.IBCKeeper.ClientKeeper.GetClientState(
		ctxTarget, client.ClientID,
	)
	if !found {
		return
	}

	// always commit when updateClient and begin a new block
	client.App.Commit()
	commitID := client.App.LastCommitID()

	client.Header = nextHeader(client)
	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: int64(client.Header.GetHeight()), Time: client.Header.Time}})

	// Set HistoricalInfo on client chain after Commit
	ctxClient := client.GetContext()
	validator := staking.NewValidator(
		sdk.ValAddress(client.Vals.Validators[0].Address), client.Vals.Validators[0].PubKey, staking.Description{},
	)
	validator.Status = sdk.Bonded
	validator.Tokens = sdk.NewInt(1000000)
	validators := []staking.Validator{validator}
	histInfo := staking.HistoricalInfo{
		Header: abci.Header{
			AppHash: commitID.Hash,
		},
		Valset: validators,
	}
	client.App.StakingKeeper.SetHistoricalInfo(ctxClient, int64(client.Header.GetHeight()), histInfo)

	consensusState := ibctmtypes.ConsensusState{
		Height:       client.Header.GetHeight(),
		Timestamp:    client.Header.Time,
		Root:         commitmenttypes.NewMerkleRoot(commitID.Hash),
		ValidatorSet: client.Vals,
	}

	chain.App.IBCKeeper.ClientKeeper.SetClientConsensusState(
		ctxTarget, client.ClientID, client.Header.GetHeight(), consensusState,
	)
	chain.App.IBCKeeper.ClientKeeper.SetClientState(
		ctxTarget, ibctmtypes.NewClientState(client.ClientID, trustingPeriod, ubdPeriod, maxClockDrift, client.Header),
	)

	// _, _, err := simapp.SignCheckDeliver(
	// 	suite.T(),
	// 	suite.cdc,
	// 	suite.app.BaseApp,
	// 	ctx.BlockHeader(),
	// 	[]sdk.Msg{clienttypes.NewMsgUpdateClient(clientID, suite.header, accountAddress)},
	// 	[]uint64{baseAccount.GetAccountNumber()},
	// 	[]uint64{baseAccount.GetSequence()},
	// 	true, true, accountPrivKey,
	// )
	// suite.Require().NoError(err)
}

func (chain *TestChain) createConnection(
	connID, counterpartyConnID, clientID, counterpartyClientID string,
	state ibctypes.State,
) connectiontypes.ConnectionEnd {
	counterparty := connectiontypes.NewCounterparty(counterpartyClientID, counterpartyConnID, commitmenttypes.NewMerklePrefix(chain.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	connection := connectiontypes.ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     connectiontypes.GetCompatibleVersions(),
	}
	ctx := chain.GetContext()
	chain.App.IBCKeeper.ConnectionKeeper.SetConnection(ctx, connID, connection)
	return connection
}

func (chain *TestChain) createChannel(
	portID, channelID, counterpartyPortID, counterpartyChannelID string,
	state ibctypes.State, order ibctypes.Order, connectionID string,
) types.Channel {
	counterparty := types.NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := types.NewChannel(state, order, counterparty,
		[]string{connectionID}, "1.0",
	)
	ctx := chain.GetContext()
	chain.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, portID, channelID, channel)
	return channel
}

func nextHeader(chain *TestChain) ibctmtypes.Header {
	return ibctmtypes.CreateTestHeader(chain.Header.SignedHeader.Header.ChainID, int64(chain.Header.GetHeight())+1,
		chain.Header.Time.Add(time.Minute), chain.Vals, chain.Signers)
}

// Mocked types
// TODO: fix tests and replace for real proofs

var (
	_ commitmentexported.Proof = validProof{nil, nil, nil}
	_ commitmentexported.Proof = invalidProof{}
)

type (
	validProof struct {
		root  commitmentexported.Root
		path  commitmentexported.Path
		value []byte
	}
	invalidProof struct{}
)

func (validProof) GetCommitmentType() commitmentexported.Type {
	return commitmentexported.Merkle
}

func (proof validProof) VerifyMembership(
	root commitmentexported.Root, path commitmentexported.Path, value []byte,
) error {
	if bytes.Equal(root.GetHash(), proof.root.GetHash()) &&
		path.String() == proof.path.String() &&
		bytes.Equal(value, proof.value) {
		return nil
	}
	return errors.New("invalid proof")
}

func (validProof) VerifyNonMembership(root commitmentexported.Root, path commitmentexported.Path) error {
	return nil
}

func (validProof) ValidateBasic() error {
	return nil
}

func (validProof) IsEmpty() bool {
	return false
}

func (invalidProof) GetCommitmentType() commitmentexported.Type {
	return commitmentexported.Merkle
}

func (invalidProof) VerifyMembership(
	root commitmentexported.Root, path commitmentexported.Path, value []byte) error {
	return errors.New("proof failed")
}

func (invalidProof) VerifyNonMembership(root commitmentexported.Root, path commitmentexported.Path) error {
	return errors.New("proof failed")
}

func (invalidProof) ValidateBasic() error {
	return errors.New("invalid proof")
}

func (invalidProof) IsEmpty() bool {
	return true
}
