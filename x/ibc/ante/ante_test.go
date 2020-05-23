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
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/ante"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// define constants used for testing
const (
	testClientIDA  = "testclientida"
	testClientIDB  = "testclientidb"
	testConnection = "testconnection"

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

// define variables used for testing
type HandlerTestSuite struct {
	suite.Suite

	cdc *codec.Codec

	chainA *TestChain
	chainB *TestChain
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.chainA = NewTestChain(testClientIDA)
	suite.chainB = NewTestChain(testClientIDB)

	suite.cdc = suite.chainA.App.Codec()

	// create client and connection during setups
	suite.chainA.CreateClient(suite.chainB)
	suite.chainB.CreateClient(suite.chainA)
	suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
	suite.chainB.createConnection(testConnection, testConnection, testClientIDA, testClientIDB, connectiontypes.OPEN)
}

func queryProof(chain *TestChain, key string) (proof commitmenttypes.MerkleProof, height int64) {
	res := chain.App.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", host.StoreKey),
		Data:  []byte(key),
		Prove: true,
	})

	height = res.Height
	proof = commitmenttypes.MerkleProof{
		Proof: res.Proof,
	}

	return
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

	suite.chainA.createChannel(cpportid, cpchanid, portid, chanid, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
	suite.chainB.createChannel(portid, chanid, cpportid, cpchanid, channeltypes.OPEN, channeltypes.ORDERED, testConnection)
	ctx = suite.chainA.GetContext()
	packetCommitmentPath := host.PacketCommitmentPath(packet.SourcePort, packet.SourceChannel, packet.Sequence)
	proof, proofHeight := queryProof(suite.chainB, packetCommitmentPath)
	msg = channel.NewMsgPacket(packet, proof, uint64(proofHeight), addr1)
	_, err = handler(cctx, suite.newTx(msg), false)
	suite.Error(err, "%+v", err) // invalid proof

	suite.chainA.updateClient(suite.chainB)
	// // commit chainA to flush to IAVL so we can get proof
	// suite.chainA.App.Commit()
	// suite.chainA.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.chainA.App.LastBlockHeight() + 1, Time: suite.chainA.Header.Time}})
	// ctx = suite.chainA.GetContext()

	proof, proofHeight = queryProof(suite.chainB, packetCommitmentPath)
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

	suite.chainA.createChannel(cpportid, cpchanid, portid, chanid, channeltypes.OPEN, channeltypes.UNORDERED, testConnection)

	suite.chainA.updateClient(suite.chainB)

	for i := 10; i >= 0; i-- {
		cctx, write := suite.chainA.GetContext().CacheContext()
		packet = channel.NewPacket(newPacket(uint64(i)).GetData(), uint64(i), portid, chanid, cpportid, cpchanid, 100, 0)
		packetCommitmentPath := host.PacketCommitmentPath(packet.SourcePort, packet.SourceChannel, uint64(i))
		proof, proofHeight := queryProof(suite.chainB, packetCommitmentPath)
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
	return chain.App.BaseApp.NewContext(false, abci.Header{ChainID: chain.Header.ChainID, Height: chain.Header.Height})
}

// createClient will create a client for clientChain on targetChain
func (chain *TestChain) CreateClient(client *TestChain) error {
	client.Header = nextHeader(client)
	// Commit and create a new block on appTarget to get a fresh CommitID
	client.App.Commit()
	commitID := client.App.LastCommitID()
	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: client.Header.Height, Time: client.Header.Time}})

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
	client.App.StakingKeeper.SetHistoricalInfo(ctxClient, client.Header.Height, histInfo)

	// Create target ctx
	ctxTarget := chain.GetContext()

	// create client
	clientState, err := ibctmtypes.Initialize(client.ClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, client.Header)
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
	// 	suite.chainA.App.BaseApp,
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
	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: client.Header.Height, Time: client.Header.Time}})

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
	client.App.StakingKeeper.SetHistoricalInfo(ctxClient, client.Header.Height, histInfo)

	consensusState := ibctmtypes.ConsensusState{
		Height:       uint64(client.Header.Height) - 1,
		Timestamp:    client.Header.Time,
		Root:         commitmenttypes.NewMerkleRoot(commitID.Hash),
		ValidatorSet: client.Vals,
	}

	chain.App.IBCKeeper.ClientKeeper.SetClientConsensusState(
		ctxTarget, client.ClientID, uint64(client.Header.Height-1), consensusState,
	)
	chain.App.IBCKeeper.ClientKeeper.SetClientState(
		ctxTarget, ibctmtypes.NewClientState(client.ClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, client.Header),
	)

	// _, _, err := simapp.SignCheckDeliver(
	// 	suite.T(),
	// 	suite.cdc,
	// 	suite.chainA.App.BaseApp,
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
	state connectiontypes.State,
) connectiontypes.ConnectionEnd {
	prefix := chain.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(commitmentexported.Merkle)
	counterparty, err := connectiontypes.NewCounterparty(counterpartyClientID, counterpartyConnID, prefix)
	if err != nil {
		panic(err)
	}
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
	state channeltypes.State, order channeltypes.Order, connectionID string,
) channeltypes.Channel {
	counterparty := channeltypes.NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := channeltypes.NewChannel(state, order, counterparty,
		[]string{connectionID}, "1.0",
	)
	ctx := chain.GetContext()
	chain.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, portID, channelID, channel)
	chain.App.ScopedIBCKeeper.NewCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	return channel
}

func nextHeader(chain *TestChain) ibctmtypes.Header {
	return ibctmtypes.CreateTestHeader(chain.Header.ChainID, chain.Header.Height+1,
		chain.Header.Time.Add(time.Minute), chain.Vals, chain.Signers)
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

// define variables used for testing
var (
	addr1 = sdk.AccAddress("testaddr1")

	portid   = "testportid"
	chanid   = "testchannel"
	cpportid = "testcpport"
	cpchanid = "testcpchannel"
)
