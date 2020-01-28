package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

const (
	testPort1 = "firstport"
	testPort2 = "secondport"

	testChannel1 = "firstchannel"
	testChannel2 = "secondchannel"
)

func (suite *KeeperTestSuite) TestVerifyClientConsensusState() {
	counterparty := types.NewCounterparty(
		testClientID2, testConnectionID2,
		suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
	)

	connection1 := types.NewConnectionEnd(
		exported.UNINITIALIZED, testClientID1, counterparty,
		types.GetCompatibleVersions(),
	)

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

			proofHeight := suite.ctx.BlockHeight() - 1

			// TODO: remove mocked types and uncomment
			// consensusKey := ibctypes.KeyConsensusState(testClientID1, uint64(suite.app.LastBlockHeight()))
			// proof, proofHeight := suite.queryProof(consensusKey)

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
					suite.ctx, tc.connection, uint64(proofHeight), validProof{}, suite.consensusState,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
					suite.ctx, tc.connection, uint64(proofHeight), invalidProof{}, suite.consensusState,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyConnectionState() {
	// connectionKey := ibctypes.KeyConnection(testConnectionID1)
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

			proofHeight := suite.ctx.BlockHeight() - 1
			// proof, proofHeight := suite.queryProof(connectionKey)

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
					suite.ctx, uint64(proofHeight), validProof{}, testConnectionID1, connection,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyConnectionState(
					suite.ctx, uint64(proofHeight), invalidProof{}, testConnectionID1, connection,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyChannelState() {
	// channelKey := ibctypes.KeyChannel(testPort1, testChannel1)

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

			proofHeight := suite.ctx.BlockHeight() - 1
			// proof, proofHeight := suite.queryProof(channelKey)

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyChannelState(
					suite.ctx, connection, uint64(proofHeight), validProof{}, testPort1,
					testChannel1, channel,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyChannelState(
					suite.ctx, connection, uint64(proofHeight), invalidProof{}, testPort1,
					testChannel1, channel,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketCommitment() {
	// commitmentKey := ibctypes.KeyPacketCommitment(testPort1, testChannel1, 1)
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

			proofHeight := suite.ctx.BlockHeight() - 1
			// proof, proofHeight := suite.queryProof(commitmentKey)

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
					suite.ctx, connection, uint64(proofHeight), validProof{}, testPort1,
					testChannel1, 1, commitmentBz,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketCommitment(
					suite.ctx, connection, uint64(proofHeight), invalidProof{}, testPort1,
					testChannel1, 1, commitmentBz,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgement() {
	// packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)
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

			proofHeight := suite.ctx.BlockHeight() - 1
			// proof, proofHeight := suite.queryProof(packetAckKey)

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
					suite.ctx, connection, uint64(proofHeight), validProof{}, testPort1,
					testChannel1, 1, ack,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgement(
					suite.ctx, connection, uint64(proofHeight), invalidProof{}, testPort1,
					testChannel1, 1, ack,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyPacketAcknowledgementAbsence() {
	// packetAckKey := ibctypes.KeyPacketAcknowledgement(testPort1, testChannel1, 1)

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

			proofHeight := suite.ctx.BlockHeight() - 1
			// proof, proofHeight := suite.queryProof(packetAckKey)

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
					suite.ctx, connection, uint64(proofHeight), validProof{}, testPort1,
					testChannel1, 1,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyPacketAcknowledgementAbsence(
					suite.ctx, connection, uint64(proofHeight), invalidProof{}, testPort1,
					testChannel1, 1,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVerifyNextSequenceRecv() {
	// nextSeqRcvKey := ibctypes.KeyNextSequenceRecv(testPort1, testChannel1)

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

			proofHeight := suite.ctx.BlockHeight() - 1
			// proof, proofHeight := suite.queryProof(nextSeqRcvKey)

			if tc.expPass {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
					suite.ctx, connection, uint64(proofHeight), validProof{}, testPort1,
					testChannel1, 1,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ConnectionKeeper.VerifyNextSequenceRecv(
					suite.ctx, connection, uint64(proofHeight), invalidProof{}, testPort1,
					testChannel1, 1,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}
