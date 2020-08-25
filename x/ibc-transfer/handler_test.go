package transfer_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
<<<<<<< HEAD
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
=======
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type HandlerTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

// constructs a send from chainA to chainB on the established channel/connection
// and sends the same coin back from chainB to chainA.
func (suite *HandlerTestSuite) TestHandleMsgTransfer() {
<<<<<<< HEAD
	handler := transfer.NewHandler(suite.chainA.App.TransferKeeper)

	// create channel capability from ibc scoped keeper and claim with transfer scoped keeper
	capName := host.ChannelCapabilityPath(testPort1, testChannel1)
	cap, err := suite.chainA.App.ScopedIBCKeeper.NewCapability(suite.chainA.GetContext(), capName)
	suite.Require().Nil(err, "could not create capability")
	err = suite.chainA.App.ScopedTransferKeeper.ClaimCapability(suite.chainA.GetContext(), cap, capName)
	suite.Require().Nil(err, "transfer module could not claim capability")

	ctx := suite.chainA.GetContext()
	msg := types.NewMsgTransfer(testPort1, testChannel1, testPrefixedCoins2, testAddr1, testAddr2.String(), 3, 110, 0)
	res, err := handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // channel does not exist

	// Setup channel from A to B
	suite.chainA.CreateClient(suite.chainB)
	suite.chainA.createConnection(testConnection, testConnection, testClientIDB, testClientIDA, connectiontypes.OPEN)
	suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2, channeltypes.OPEN, channeltypes.ORDERED, testConnection)

	res, err = handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // next send sequence not found

	nextSeqSend := uint64(1)
	suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(ctx, testPort1, testChannel1, nextSeqSend)
	res, err = handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // sender has insufficient coins

	_ = suite.chainA.App.BankKeeper.SetBalances(ctx, testAddr1, testCoins)
	res, err = handler(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed

	// test when the source is false
	msg = types.NewMsgTransfer(testPort1, testChannel1, testPrefixedCoins2, testAddr1, testAddr2.String(), 3, 110, 0)
	_ = suite.chainA.App.BankKeeper.SetBalances(ctx, testAddr1, testPrefixedCoins2)

	res, err = handler(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Nil(res, "%+v", res) // incorrect denom prefix

	msg = types.NewMsgTransfer(testPort1, testChannel1, testPrefixedCoins1, testAddr1, testAddr2.String(), 3, 110, 0)
	suite.chainA.App.BankKeeper.SetSupply(ctx, banktypes.NewSupply(testPrefixedCoins1))
	_ = suite.chainA.App.BankKeeper.SetBalances(ctx, testAddr1, testPrefixedCoins1)

	res, err = handler(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res, "%+v", res) // successfully executed
}
=======
	clientA, clientB, _, _, channelA, channelB := suite.coordinator.Setup(suite.chainA, suite.chainB)
	originalBalance := suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74

	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))

	// send from chainA to chainB
	msg := types.NewMsgTransfer(channelA.PortID, channelA.ID, coinToSendToB, suite.chainA.SenderAccount.GetAddress(), suite.chainB.SenderAccount.GetAddress().String(), 110, 0)

	err := suite.coordinator.SendMsg(suite.chainA, suite.chainB, clientB, msg)
	suite.Require().NoError(err) // message committed

	// relay send
	fungibleTokenPacket := types.NewFungibleTokenPacketData(coinToSendToB.Denom, coinToSendToB.Amount.Uint64(), suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String())
	packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, 110, 0)
	ack := types.FungibleTokenPacketAcknowledgement{Success: true}
	err = suite.coordinator.RelayPacket(suite.chainA, suite.chainB, clientA, clientB, packet, ack.GetBytes())
	suite.Require().NoError(err) // relay committed

	// check that voucher exists on chain B
	voucherDenomTrace := types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), sdk.DefaultBondDenom))
	balance := suite.chainB.App.BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom())

<<<<<<< HEAD
	height := clientexported.NewHeight(0, 1)
	header := ibctmtypes.CreateTestHeader(clientID, height, now, valSet, signers)
=======
	coinToSendBackToA := types.GetTransferCoin(channelB.PortID, channelB.ID, sdk.DefaultBondDenom, 100)
	suite.Require().Equal(coinToSendBackToA, balance)
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74

	// send from chainB back to chainA
	msg = types.NewMsgTransfer(channelB.PortID, channelB.ID, coinToSendBackToA, suite.chainB.SenderAccount.GetAddress(), suite.chainA.SenderAccount.GetAddress().String(), 110, 0)

	err = suite.coordinator.SendMsg(suite.chainB, suite.chainA, clientA, msg)
	suite.Require().NoError(err) // message committed

	// relay send
	// NOTE: fungible token is prefixed with the full trace in order to verify the packet commitment
	voucherDenom := voucherDenomTrace.GetPrefix() + voucherDenomTrace.BaseDenom
	fungibleTokenPacket = types.NewFungibleTokenPacketData(voucherDenom, coinToSendBackToA.Amount.Uint64(), suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String())
	packet = channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), 1, channelB.PortID, channelB.ID, channelA.PortID, channelA.ID, 110, 0)
	err = suite.coordinator.RelayPacket(suite.chainB, suite.chainA, clientB, clientA, packet, ack.GetBytes())
	suite.Require().NoError(err) // relay committed

	balance = suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)

	// check that the balance on chainA returned back to the original state
	suite.Require().Equal(originalBalance, balance)

	// check that module account escrow address is empty
	escrowAddress := types.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
	balance = suite.chainA.App.BankKeeper.GetBalance(suite.chainA.GetContext(), escrowAddress, sdk.DefaultBondDenom)
	suite.Require().Equal(sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt()), balance)

	// check that balance on chain B is empty
	balance = suite.chainB.App.BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom())
	suite.Require().Zero(balance.Amount.Int64())
}

<<<<<<< HEAD
// nolint: unused
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
	client.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: client.Header.SignedHeader.Header.Height, Time: client.Header.Time}})

	// Set HistoricalInfo on client chain after Commit
	ctxClient := client.GetContext()
	validator := stakingtypes.NewValidator(
		sdk.ValAddress(client.Vals.Validators[0].Address), client.Vals.Validators[0].PubKey, stakingtypes.Description{},
	)
	validator.Status = sdk.Bonded
	validator.Tokens = sdk.NewInt(1000000)
	validators := []stakingtypes.Validator{validator}
	histInfo := stakingtypes.HistoricalInfo{
		Header: abci.Header{
			AppHash: commitID.Hash,
		},
		Valset: validators,
	}
	client.App.StakingKeeper.SetHistoricalInfo(ctxClient, client.Header.SignedHeader.Header.Height, histInfo)

	height := clientexported.NewHeight(0, uint64(client.Header.SignedHeader.Header.Height))
	consensusState := ibctmtypes.ConsensusState{
		Height:       height,
		Timestamp:    client.Header.Time,
		Root:         commitmenttypes.NewMerkleRoot(commitID.Hash),
		ValidatorSet: client.Vals,
	}

	chain.App.IBCKeeper.ClientKeeper.SetClientConsensusState(
		ctxTarget, client.ClientID, height, consensusState,
	)
	chain.App.IBCKeeper.ClientKeeper.SetClientState(
		ctxTarget, ibctmtypes.NewClientState(client.ClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, client.Header, commitmenttypes.GetSDKSpecs()),
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
	state connectiontypes.State,
) connectiontypes.ConnectionEnd {
	counterparty := connectiontypes.NewCounterparty(counterpartyClientID, counterpartyConnID, commitmenttypes.NewMerklePrefix(chain.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	connection := connectiontypes.ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     connectiontypes.GetCompatibleEncodedVersions(),
	}
	ctx := chain.GetContext()
	chain.App.IBCKeeper.ConnectionKeeper.SetConnection(ctx, connID, connection)
	return connection
}

// nolint: unused
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
	return channel
}

func nextHeader(chain *TestChain) ibctmtypes.Header {
	height := clientexported.NewHeight(0, uint64(chain.Header.SignedHeader.Header.Height+1))
	return ibctmtypes.CreateTestHeader(chain.Header.SignedHeader.Header.ChainID, height,
		chain.Header.Time.Add(time.Minute), chain.Vals, chain.Signers)
=======
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74
}
