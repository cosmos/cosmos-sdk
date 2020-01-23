package keeper_test

import (
	"fmt"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	testPort1 = "firstport"
	testPort2 = "secondport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
)

func (suite *KeeperTestSuite) TestVerifyClientConsensusState() {
	counterparty := types.Counterparty{Prefix: suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix()}
	connection1 := types.ConnectionEnd{ClientID: testClientID1, Counterparty: counterparty}

	cases := []struct {
		msg        string
		connection types.ConnectionEnd
		malleate   func()
		expPass    bool
	}{
		{"verification success", connection1, func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", connection1, func() {}, false},
		{"verification failed", connection1, func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			consensusKey := ibctypes.KeyConsensusState(testClientID1, uint64(suite.ctx.BlockHeight()))

			proof, proofHeight := suite.queryProof(consensusKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.ctx, tc.connection, uint64(proofHeight), proof, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().False(proof.IsEmpty())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().True(proof.IsEmpty())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyConnectionState() {
	counterparty := types.NewCounterparty(
		testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)
	connection1 := types.ConnectionEnd{ClientID: testClientID1, Counterparty: counterparty}

	connectionKey := ibctypes.KeyConnection(testConnectionID1)

	cases := []struct {
		msg          string
		connectionID string
		connection   types.ConnectionEnd
		malleate     func() error
		expPass      bool
	}{
		{"verification success", testConnectionID2, connection1, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID1, clientexported.Tendermint, suite.consensusState)
			return err
		}, true},
		{"client state not found", testConnectionID2, connection1, func() error { return nil }, false},
		{"verification failed", testConnectionID2, connection1, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID2, clientexported.Tendermint, suite.consensusState)
			return err
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			err := tc.malleate()
			suite.Require().NoError(err)

			proof, proofHeight := suite.queryProof(connectionKey)

			err = suite.app.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.ctx, uint64(proofHeight), proof, tc.connectionID, tc.connection, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().False(proof.IsEmpty())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().True(proof.IsEmpty())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyChannelState() {
	counterpartyConn := types.NewCounterparty(
		testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)
	connection1 := types.ConnectionEnd{ClientID: testClientID1, Counterparty: counterpartyConn}

	counterparty := channeltypes.NewCounterparty(testPort2, testChannel2)
	channel := channeltypes.NewChannel(
		channelexported.OPEN, channelexported.ORDERED, counterparty,
		[]string{testConnectionID1}, "1.0",
	)

	channelKey := ibctypes.KeyChannel(testPort1, testChannel1)

	cases := []struct {
		msg       string
		portID    string
		channelID string
		channel   channeltypes.Channel
		malleate  func() error
		expPass   bool
	}{
		{"verification success", testPort1, testChannel1, channel, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID1, clientexported.Tendermint, suite.consensusState)
			return err
		}, true},
		{"client state not found", testPort1, testChannel1, channel, func() error { return nil }, false},
		{"verification failed", testPort1, testChannel1, channel, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID2, clientexported.Tendermint, suite.consensusState)
			return err
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			err := tc.malleate()
			suite.Require().NoError(err)

			proof, proofHeight := suite.queryProof(channelKey)

			err = suite.app.IBCKeeper.ConnectionKeeper.VerifyChannelState(
				suite.ctx, connection1, uint64(proofHeight), proof, tc.portID,
				tc.channelID, tc.channel, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().False(proof.IsEmpty())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().True(proof.IsEmpty())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketCommitment() {
	counterpartyConn := types.NewCounterparty(
		testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)
	connection1 := types.ConnectionEnd{ClientID: testClientID1, Counterparty: counterpartyConn}
	commitmentKey := ibctypes.KeyPacketCommitment(testPort1, testChannel1, 1)
	commitmentBz := []byte{1}

	cases := []struct {
		msg           string
		portID        string
		channelID     string
		sequence      uint64
		commitementBz []byte
		malleate      func() error
		expPass       bool
	}{
		{"verification success", testPort1, testChannel1, 1, commitmentBz, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID1, clientexported.Tendermint, suite.consensusState)
			return err
		}, true},
		{"client state not found", testPort1, testChannel1, 1, commitmentBz, func() error { return nil }, false},
		{"verification failed", testPort1, testChannel1, 1, commitmentBz, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID2, clientexported.Tendermint, suite.consensusState)
			return err
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			err := tc.malleate()
			suite.Require().NoError(err)

			proof, proofHeight := suite.queryProof(commitmentKey)

			err = suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
				suite.ctx, connection1, uint64(proofHeight), proof, tc.portID,
				tc.channelID, tc.sequence, tc.commitementBz, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().False(proof.IsEmpty())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().True(proof.IsEmpty())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgement() {
	counterpartyConn := types.NewCounterparty(
		testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)
	connection1 := types.ConnectionEnd{ClientID: testClientID1, Counterparty: counterpartyConn}

	packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)
	ack := []byte("hello")

	cases := []struct {
		msg       string
		portID    string
		channelID string
		sequence  uint64
		ack       []byte
		malleate  func() error
		expPass   bool
	}{
		{"verification success", testPort1, testChannel1, 1, ack, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID1, clientexported.Tendermint, suite.consensusState)
			return err
		}, true},
		{"client state not found", testPort1, testChannel1, 1, ack, func() error { return nil }, false},
		{"verification failed", testPort1, testChannel1, 1, ack, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID2, clientexported.Tendermint, suite.consensusState)
			return err
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			err := tc.malleate()
			suite.Require().NoError(err)

			proof, proofHeight := suite.queryProof(packetAckKey)

			err = suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
				suite.ctx, connection1, uint64(proofHeight), proof, tc.portID,
				tc.channelID, tc.sequence, tc.ack, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().False(proof.IsEmpty())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().True(proof.IsEmpty())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	counterpartyConn := types.NewCounterparty(
		testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)
	connection1 := types.ConnectionEnd{ClientID: testClientID1, Counterparty: counterpartyConn}

	channelKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)

	cases := []struct {
		msg       string
		portID    string
		channelID string
		sequence  uint64
		malleate  func() error
		expPass   bool
	}{
		{"verification success", testPort1, testChannel1, 1, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID1, clientexported.Tendermint, suite.consensusState)
			return err
		}, true},
		{"client state not found", testPort1, testChannel1, 1, func() error { return nil }, false},
		{"verification failed", testPort1, testChannel1, 1, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID2, clientexported.Tendermint, suite.consensusState)
			return err
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			err := tc.malleate()
			suite.Require().NoError(err)

			proof, proofHeight := suite.queryProof(channelKey)

			err = suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
				suite.ctx, connection1, uint64(proofHeight), proof, tc.portID,
				tc.channelID, tc.sequence, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().False(proof.IsEmpty())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().True(proof.IsEmpty())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyNextSequenceRecv() {
	counterpartyConn := types.NewCounterparty(
		testClientID2, testConnectionID2, suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)
	connection1 := types.ConnectionEnd{ClientID: testClientID1, Counterparty: counterpartyConn}

	nextSeqRcvKey := ibctypes.KeyNextSequenceRecv(testPort1, testChannel1)

	cases := []struct {
		msg         string
		portID      string
		channelID   string
		nextSeqRecv uint64
		malleate    func() error
		expPass     bool
	}{
		{"verification success", testPort1, testChannel1, 1, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID1, clientexported.Tendermint, suite.consensusState)
			return err
		}, true},
		{"client state not found", testPort1, testChannel1, 1, func() error { return nil }, false},
		{"verification failed", testPort1, testChannel1, 1, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID2, clientexported.Tendermint, suite.consensusState)
			return err
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			err := tc.malleate()
			suite.Require().NoError(err)

			proof, proofHeight := suite.queryProof(nextSeqRcvKey)

			err = suite.app.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
				suite.ctx, connection1, uint64(proofHeight), proof, tc.portID,
				tc.channelID, tc.nextSeqRecv, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().False(proof.IsEmpty())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().True(proof.IsEmpty())
			}
		})
	}
}
