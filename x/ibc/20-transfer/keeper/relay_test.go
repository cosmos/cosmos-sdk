package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypestm "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) createClient() {
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	consensusState := clienttypestm.ConsensusState{
		ChainID: testChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}

	_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClient, testClientType, consensusState)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) updateClient() {
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

func (suite *KeeperTestSuite) createConnection(state connection.State) {
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

func (suite *KeeperTestSuite) createChannel(portID string, chanID string, connID string, counterpartyPort string, counterpartyChan string, state channel.State) {
	ch := channel.Channel{
		State:    state,
		Ordering: testChannelOrder,
		Counterparty: channel.Counterparty{
			PortID:    counterpartyPort,
			ChannelID: counterpartyChan,
		},
		ConnectionHops: []string{connID},
		Version:        testChannelVersion,
	}

	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, portID, chanID, ch)
}

func (suite *KeeperTestSuite) queryProof(key string) (proof commitment.Proof, height int64) {
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

func (suite *KeeperTestSuite) TestSendTransfer() {
	// test the situation where the source is true
	isSourceChain := true

	err := suite.app.IBCKeeper.TransferKeeper.SendTransfer(suite.ctx, testPort1, testChannel1, testCoins, testAddr1, testAddr2, isSourceChain)
	suite.Error(err) // channel does not exist

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channel.OPEN)
	err = suite.app.IBCKeeper.TransferKeeper.SendTransfer(suite.ctx, testPort1, testChannel1, testCoins, testAddr1, testAddr2, isSourceChain)
	suite.Error(err) // next send sequence not found

	nextSeqSend := uint64(1)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, nextSeqSend)
	err = suite.app.IBCKeeper.TransferKeeper.SendTransfer(suite.ctx, testPort1, testChannel1, testCoins, testAddr1, testAddr2, isSourceChain)
	suite.Error(err) // sender has insufficient coins

	_ = suite.app.BankKeeper.SetCoins(suite.ctx, testAddr1, testCoins)
	err = suite.app.IBCKeeper.TransferKeeper.SendTransfer(suite.ctx, testPort1, testChannel1, testCoins, testAddr1, testAddr2, isSourceChain)
	suite.NoError(err) // successfully executed

	senderCoins := suite.app.BankKeeper.GetCoins(suite.ctx, testAddr1)
	suite.Equal(sdk.Coins(nil), senderCoins)

	escrowCoins := suite.app.BankKeeper.GetCoins(suite.ctx, types.GetEscrowAddress(testPort1, testChannel1))
	suite.Equal(testCoins, escrowCoins)

	newNextSeqSend, found := suite.app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(suite.ctx, testPort1, testChannel1)
	suite.True(found)
	suite.Equal(nextSeqSend+1, newNextSeqSend)

	packetCommitment := suite.app.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.ctx, testPort1, testChannel1, nextSeqSend)
	suite.NotNil(packetCommitment)

	// test the situation where the source is false
	isSourceChain = false

	_ = suite.app.BankKeeper.SetCoins(suite.ctx, testAddr1, testPrefixedCoins2)
	err = suite.app.IBCKeeper.TransferKeeper.SendTransfer(suite.ctx, testPort1, testChannel1, testPrefixedCoins2, testAddr1, testAddr2, isSourceChain)
	suite.Error(err) // incorrect denom prefix

	suite.app.SupplyKeeper.SetSupply(suite.ctx, supply.NewSupply(testPrefixedCoins1))
	_ = suite.app.BankKeeper.SetCoins(suite.ctx, testAddr1, testPrefixedCoins1)
	err = suite.app.IBCKeeper.TransferKeeper.SendTransfer(suite.ctx, testPort1, testChannel1, testPrefixedCoins1, testAddr1, testAddr2, isSourceChain)
	suite.NoError(err) // successfully executed

	senderCoins = suite.app.BankKeeper.GetCoins(suite.ctx, testAddr1)
	suite.Equal(sdk.Coins(nil), senderCoins)

	totalSupply := suite.app.SupplyKeeper.GetSupply(suite.ctx)
	suite.Equal(sdk.Coins(nil), totalSupply.GetTotal()) // supply should be deflated
}

func (suite *KeeperTestSuite) TestReceiveTransfer() {
	// test the situation where the source is true
	source := true

	packetData := types.NewPacketData(testPrefixedCoins1, testAddr1, testAddr2, source)
	err := suite.app.IBCKeeper.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.Error(err) // incorrect denom prefix

	packetData.Amount = testPrefixedCoins2
	err = suite.app.IBCKeeper.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.NoError(err) // successfully executed

	totalSupply := suite.app.SupplyKeeper.GetSupply(suite.ctx)
	suite.Equal(testPrefixedCoins2, totalSupply.GetTotal()) // supply should be inflated

	receiverCoins := suite.app.BankKeeper.GetCoins(suite.ctx, packetData.Receiver)
	suite.Equal(testPrefixedCoins2, receiverCoins)

	// test the situation where the source is false
	packetData.Source = false

	packetData.Amount = testPrefixedCoins2
	err = suite.app.IBCKeeper.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.Error(err) // incorrect denom prefix

	packetData.Amount = testPrefixedCoins1
	err = suite.app.IBCKeeper.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.Error(err) // insufficient coins in the corresponding escrow account

	escrowAddress := types.GetEscrowAddress(testPort2, testChannel2)
	_ = suite.app.BankKeeper.SetCoins(suite.ctx, escrowAddress, testCoins)
	_ = suite.app.BankKeeper.SetCoins(suite.ctx, packetData.Receiver, sdk.Coins{})
	err = suite.app.IBCKeeper.TransferKeeper.ReceiveTransfer(suite.ctx, testPort1, testChannel1, testPort2, testChannel2, packetData)
	suite.NoError(err) // successfully executed

	escrowCoins := suite.app.BankKeeper.GetCoins(suite.ctx, escrowAddress)
	suite.Equal(sdk.Coins(nil), escrowCoins)

	receiverCoins = suite.app.BankKeeper.GetCoins(suite.ctx, packetData.Receiver)
	suite.Equal(testCoins, receiverCoins)
}

func (suite *KeeperTestSuite) TestReceivePacket() {
	packetSeq := uint64(1)
	packetTimeout := uint64(100)

	packetDataBz := []byte("invaliddata")
	packet := channel.NewPacket(packetSeq, packetTimeout, testPort2, testChannel2, testPort2, testChannel1, packetDataBz)
	packetCommitmentPath := channel.PacketCommitmentPath(testPort2, testChannel2, packetSeq)

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, []byte("invalidcommitment"))
	suite.updateClient()
	proofPacket, proofHeight := suite.queryProof(packetCommitmentPath)

	suite.createChannel(testPort2, testChannel1, testConnection, testPort2, testChannel2, channel.OPEN)
	err := suite.app.IBCKeeper.TransferKeeper.ReceivePacket(suite.ctx, packet, proofPacket, uint64(proofHeight))
	suite.Error(err) // invalid port id

	packet.DestinationPort = testPort1
	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channel.OPEN)
	err = suite.app.IBCKeeper.TransferKeeper.ReceivePacket(suite.ctx, packet, proofPacket, uint64(proofHeight))
	suite.Error(err) // packet membership verification failed due to invalid counterparty packet commitment

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, packetDataBz)
	suite.updateClient()
	proofPacket, proofHeight = suite.queryProof(packetCommitmentPath)
	err = suite.app.IBCKeeper.TransferKeeper.ReceivePacket(suite.ctx, packet, proofPacket, uint64(proofHeight))
	suite.Error(err) // invalid packet data

	// test the situation where the source is true
	source := true

	packetData := types.NewPacketData(testPrefixedCoins2, testAddr1, testAddr2, source)
	packetDataBz, _ = suite.cdc.MarshalBinaryBare(packetData)
	packet = channel.NewPacket(packetSeq, packetTimeout, testPort2, testChannel2, testPort1, testChannel1, packetDataBz)

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, packetDataBz)
	suite.updateClient()
	proofPacket, proofHeight = suite.queryProof(packetCommitmentPath)
	err = suite.app.IBCKeeper.TransferKeeper.ReceivePacket(suite.ctx, packet, proofPacket, uint64(proofHeight))
	suite.Error(err) // invalid denom prefix

	packetData = types.NewPacketData(testPrefixedCoins1, testAddr1, testAddr2, source)
	packetDataBz, _ = suite.cdc.MarshalBinaryBare(packetData)
	packet = channel.NewPacket(packetSeq, packetTimeout, testPort2, testChannel2, testPort1, testChannel1, packetDataBz)

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, packetDataBz)
	suite.updateClient()
	proofPacket, proofHeight = suite.queryProof(packetCommitmentPath)
	err = suite.app.IBCKeeper.TransferKeeper.ReceivePacket(suite.ctx, packet, proofPacket, uint64(proofHeight))
	suite.NoError(err) // successfully executed

	totalSupply := suite.app.SupplyKeeper.GetSupply(suite.ctx)
	suite.Equal(testPrefixedCoins1, totalSupply.GetTotal()) // supply should be inflated

	receiverCoins := suite.app.BankKeeper.GetCoins(suite.ctx, packetData.Receiver)
	suite.Equal(testPrefixedCoins1, receiverCoins)

	// test the situation where the source is false
	source = false

	packetData = types.NewPacketData(testPrefixedCoins1, testAddr1, testAddr2, source)
	packetDataBz, _ = suite.cdc.MarshalBinaryBare(packetData)
	packet = channel.NewPacket(packetSeq, packetTimeout, testPort2, testChannel2, testPort1, testChannel1, packetDataBz)

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, packetDataBz)
	suite.updateClient()
	proofPacket, proofHeight = suite.queryProof(packetCommitmentPath)
	err = suite.app.IBCKeeper.TransferKeeper.ReceivePacket(suite.ctx, packet, proofPacket, uint64(proofHeight))
	suite.Error(err) // invalid denom prefix

	packetData = types.NewPacketData(testPrefixedCoins2, testAddr1, testAddr2, source)
	packetDataBz, _ = suite.cdc.MarshalBinaryBare(packetData)
	packet = channel.NewPacket(packetSeq, packetTimeout, testPort2, testChannel2, testPort1, testChannel1, packetDataBz)

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, packetDataBz)
	suite.updateClient()
	proofPacket, proofHeight = suite.queryProof(packetCommitmentPath)
	err = suite.app.IBCKeeper.TransferKeeper.ReceivePacket(suite.ctx, packet, proofPacket, uint64(proofHeight))
	suite.Error(err) // insufficient coins in the corresponding escrow account

	escrowAddress := types.GetEscrowAddress(testPort1, testChannel1)
	_ = suite.app.BankKeeper.SetCoins(suite.ctx, escrowAddress, testCoins)
	_ = suite.app.BankKeeper.SetCoins(suite.ctx, packetData.Receiver, sdk.Coins{})
	err = suite.app.IBCKeeper.TransferKeeper.ReceivePacket(suite.ctx, packet, proofPacket, uint64(proofHeight))
	suite.NoError(err) // successfully executed

	receiverCoins = suite.app.BankKeeper.GetCoins(suite.ctx, packetData.Receiver)
	suite.Equal(testCoins, receiverCoins)

	escrowCoins := suite.app.BankKeeper.GetCoins(suite.ctx, escrowAddress)
	suite.Equal(sdk.Coins(nil), escrowCoins)
}
