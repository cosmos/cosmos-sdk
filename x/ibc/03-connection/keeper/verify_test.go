package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
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
			suite.updateClient(testClientID1)

			consensusKey := ibctypes.KeyConsensusState(testClientID1, uint64(suite.ctx.BlockHeight()-1))

			proof, proofHeight := suite.queryProof(consensusKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.ctx, tc.connection, uint64(proofHeight), proof, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().False(proof.IsEmpty())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyConnectionState() {

	connectionKey := ibctypes.KeyConnection(testConnectionID1)

	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", func() {}, false},
		{"verification failed", func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			fmt.Println("In case", tc.msg)
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.updateClient(testClientID1)

			proof, proofHeight := suite.queryProof(connectionKey)
			fmt.Println(tc.msg, proof, proofHeight)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
				suite.ctx, uint64(proofHeight), proof, testConnectionID1, connection, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyChannelState() {
	channelKey := ibctypes.KeyChannel(testPort1, testChannel1)

	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", func() {}, false},
		{"verification failed", func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			channel := suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2,
				channelexported.OPEN, channelexported.ORDERED, testConnectionID1,
			)
			suite.updateClient(testClientID1)

			proof, proofHeight := suite.queryProof(channelKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyChannelState(
				suite.ctx, connection, uint64(proofHeight), proof, testPort1,
				testChannel1, channel, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketCommitment() {

	commitmentKey := ibctypes.KeyPacketCommitment(testPort1, testChannel1, 1)
	commitmentBz := []byte("commitment")

	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", func() {}, false},
		{"verification failed", func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort1, testChannel1, 1, commitmentBz)
			suite.updateClient(testClientID1)

			proof, proofHeight := suite.queryProof(commitmentKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
				suite.ctx, connection, uint64(proofHeight), proof, testPort1,
				testChannel1, 1, commitmentBz, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgement() {
	packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)
	ack := []byte("acknowledgement")

	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", func() {}, false},
		{"verification failed", func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.app.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.ctx, testPort1, testChannel1, 1, ack)
			suite.updateClient(testClientID1)

			proof, proofHeight := suite.queryProof(packetAckKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
				suite.ctx, connection, uint64(proofHeight), proof, testPort1,
				testChannel1, 1, ack, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)

	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", func() {}, false},
		{"verification failed", func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.updateClient(testClientID1)

			proof, proofHeight := suite.queryProof(packetAckKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
				suite.ctx, connection, uint64(proofHeight), proof, testPort1,
				testChannel1, 1, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyNextSequenceRecv() {
	nextSeqRcvKey := ibctypes.KeyNextSequenceRecv(testPort1, testChannel1)

	cases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"verification success", func() {
			suite.createClient(testClientID1)
		}, true},
		{"client state not found", func() {}, false},
		{"verification failed", func() {
			suite.createClient(testClientID2)
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			connection := suite.createConnection(testConnectionID1, testConnectionID2, testClientID1, testClientID2, exported.OPEN)
			suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.ctx, testPort1, testChannel1, 1)
			suite.updateClient(testClientID1)

			proof, proofHeight := suite.queryProof(nextSeqRcvKey)

			err := suite.app.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
				suite.ctx, connection, uint64(proofHeight), proof, testPort1,
				testChannel1, 1, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}
