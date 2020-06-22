package keeper_test

import (
	"fmt"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type testCase = struct {
	msg      string
	malleate func()
	expPass  bool
}

// TestChanOpenInit tests the OpenInit handshake call for channels. It uses message passing
// to enter into the appropriate state and then calls ChanOpenInit directly. The channel is
// being created on chainA. The port capability must be created on chainA before ChanOpenInit
// can succeed.
func (suite *KeeperTestSuite) TestChanOpenInit() {
	var (
		connA   *ibctesting.TestConnection
		connB   *ibctesting.TestConnection
		portCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			suite.chainA.CreatePortCapability(connA.NextTestChannel().PortID)
			portCap = suite.chainA.GetPortCapability(connA.NextTestChannel().PortID)
		}, true},
		{"channel already exists", func() {
			_, _, connA, connB = suite.coordinator.Setup(suite.chainA, suite.chainB)
		}, false},
		{"connection doesn't exist", func() {
			// any non-nil values of connA and connB are acceptable
			suite.Require().NotNil(connA)
			suite.Require().NotNil(connB)
		}, false},
		{"connection is UNINITIALIZED", func() {
			channel := types.NewChannel(types.UNINITIALIZED, types.ORDERED, types.NewCounterparty("port", "channel"), []string{"connB"}, ibctesting.ChannelVersion)
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(suite.chainA.GetContext(), "port", "channel", channel)
			portCap = nil
		}, false},
		{"capability is incorrect", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			portCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			// run test for all types of ordering
			for _, order := range []types.Order{types.UNORDERED, types.ORDERED} {
				suite.SetupTest() // reset
				tc.malleate()

				counterparty := types.NewCounterparty(connB.FirstOrNextTestChannel().PortID, connB.FirstOrNextTestChannel().ID)
				channelA := connA.FirstOrNextTestChannel()

				cap, err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenInit(
					suite.chainA.GetContext(), order, []string{connA.ID},
					channelA.PortID, channelA.ID, portCap, counterparty, ibctesting.ChannelVersion,
				)

				if tc.expPass {
					suite.Require().NoError(err)
					suite.Require().NotNil(cap)

					chanCap, ok := suite.chainA.App.ScopedIBCKeeper.GetCapability(
						suite.chainA.GetContext(),
						host.ChannelCapabilityPath(channelA.PortID, channelA.ID),
					)
					suite.Require().True(ok, "could not retrieve channel capapbility after successful ChanOpenInit")
					suite.Require().Equal(chanCap.String(), cap.String(), "channel capability is not correct")
				} else {
					suite.Require().Error(err)
				}
			}
		})
	}
}

// TestChanOpenTry tests the OpenTry handshake call for channels. It uses message passing
// to enter into the appropriate state and then calls ChanOpenInit directly. The channel
// is being created on chainB. The port capability must be created on chainB before
// ChanOpenTry can succeed.
func (suite *KeeperTestSuite) TestChanOpenTry() {
	var (
		connA      *ibctesting.TestConnection
		connB      *ibctesting.TestConnection
		portCap    *capabilitytypes.Capability
		heightDiff uint64
	)

	testCases := []testCase{
		{"success", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			suite.coordinator.CreateChannelOpenInit(suite.chainA, suite.chainB, connA, connB, types.ORDERED)

			suite.chainB.CreatePortCapability(connB.NextTestChannel().PortID)
			portCap = suite.chainB.GetPortCapability(connB.NextTestChannel().PortID)
		}, true},
		{"previous channel with invalid state", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)

			// make previous channel have wrong ordering
			suite.coordinator.CreateChannelOpenInit(suite.chainB, suite.chainA, connB, connA, types.UNORDERED)
		}, false},
		{"connection doesn't exist", func() {
			// any non-nil values of connA and connB are acceptable
			suite.Require().NotNil(connA)
			suite.Require().NotNil(connB)
		}, false},
		{"connection is not OPEN", func() {
			clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)

			var err error
			connA, connB, err = suite.coordinator.CreateConnectionInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)
		}, false},
		{"consensus state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			suite.coordinator.CreateChannelOpenInit(suite.chainA, suite.chainB, connA, connB, types.ORDERED)

			suite.chainB.CreatePortCapability(connB.NextTestChannel().PortID)
			portCap = suite.chainB.GetPortCapability(connB.NextTestChannel().PortID)

			heightDiff = 3 // consensus state doesn't exist at this height
		}, false},
		{"channel verification failed", func() {
			// not creating a channel on chainA will result in an invalid proof of existence
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
		}, false},
		{"port capability not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
			suite.coordinator.CreateChannelOpenInit(suite.chainA, suite.chainB, connA, connB, types.ORDERED)

			portCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			heightDiff = 0    // must be explicitly changed

			tc.malleate()
			counterparty := types.NewCounterparty(connA.FirstOrNextTestChannel().PortID, connA.FirstOrNextTestChannel().ID)
			channelB := connB.FirstOrNextTestChannel()

			channelKey := host.KeyChannel(counterparty.PortID, counterparty.ChannelID)
			proof, proofHeight := suite.chainA.QueryProof(channelKey)

			cap, err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanOpenTry(
				suite.chainB.GetContext(), types.ORDERED, []string{connB.ID},
				channelB.PortID, channelB.ID, portCap, counterparty, ibctesting.ChannelVersion, ibctesting.ChannelVersion,
				proof, proofHeight+1+heightDiff,
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(cap)

				chanCap, ok := suite.chainA.App.ScopedIBCKeeper.GetCapability(
					suite.chainA.GetContext(),
					host.ChannelCapabilityPath(channelB.PortID, channelB.ID),
				)
				suite.Require().True(ok, "could not retrieve channel capapbility after successful ChanOpenInit")
				suite.Require().Equal(chanCap.Index, cap.Index, "channel capability is not correct")
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

/*
func (suite *KeeperTestSuite) TestChanOpenAck() {
	channelKey := host.KeyChannel(testPort2, testChannel2)

	var channelCap *capabilitytypes.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connection.OPEN,
			)
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.INIT,
				types.ORDERED, testConnectionIDB,
			)
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.TRYOPEN,
				types.ORDERED, testConnectionIDA,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not INIT or TRYOPEN", func() {
			suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.UNINITIALIZED,
				types.ORDERED, testConnectionIDA,
			)
		}, false},
		{"connection not found", func() {
			suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.TRYOPEN,
				types.ORDERED, testConnectionIDA,
			)
		}, false},
		{"connection is not OPEN", func() {
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.TRYOPEN,
			)
			suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.TRYOPEN,
				types.ORDERED, testConnectionIDA,
			)
		}, false},
		{"consensus state not found", func() {
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.TRYOPEN,
				types.ORDERED, testConnectionIDA,
			)
		}, false},
		{"channel verification failed", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.TRYOPEN,
				types.ORDERED, testConnectionIDA,
			)
		}, false},
		{"channel capability not found", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connection.OPEN,
			)
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.INIT,
				types.ORDERED, testConnectionIDB,
			)
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.TRYOPEN,
				types.ORDERED, testConnectionIDA,
			)
			channelCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainA.App.ScopedIBCKeeper.NewCapability(suite.chainA.GetContext(), host.ChannelCapabilityPath(testPort1, testChannel1))
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()

			suite.chainA.updateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			proof, proofHeight := queryProof(suite.chainB, channelKey)

			if tc.expPass {
				err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenAck(
					suite.chainA.GetContext(), testPort1, testChannel1, channelCap, testChannelVersion,
					proof, proofHeight+1,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenAck(
					suite.chainA.GetContext(), testPort1, testChannel1, channelCap, testChannelVersion,
					proof, proofHeight+1,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanOpenConfirm() {
	channelKey := host.KeyChannel(testPort2, testChannel2)

	var channelCap *capabilitytypes.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connection.TRYOPEN,
			)
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.OPEN,
				types.ORDERED, testConnectionIDB,
			)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2,
				types.TRYOPEN, types.ORDERED, testConnectionIDA)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not TRYOPEN", func() {
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.UNINITIALIZED,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"connection not found", func() {
			suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.TRYOPEN,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"connection is not OPEN", func() {
			suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connection.TRYOPEN,
			)
			suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.TRYOPEN,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"consensus state not found", func() {
			suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connection.OPEN,
			)
			suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.TRYOPEN,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"channel verification failed", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connection.OPEN,
			)
			suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.TRYOPEN,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"channel capability not found", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connection.TRYOPEN,
			)
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.OPEN,
				types.ORDERED, testConnectionIDB,
			)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2,
				types.TRYOPEN, types.ORDERED, testConnectionIDA)
			channelCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainB.App.ScopedIBCKeeper.NewCapability(suite.chainB.GetContext(), host.ChannelCapabilityPath(testPort1, testChannel1))
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()

			suite.chainA.updateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			proof, proofHeight := queryProof(suite.chainA, channelKey)

			if tc.expPass {
				err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanOpenConfirm(
					suite.chainB.GetContext(), testPort1, testChannel1,
					channelCap, proof, proofHeight+1,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanOpenConfirm(
					suite.chainB.GetContext(), testPort1, testChannel1, channelCap,
					proof, proofHeight+1,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanCloseInit() {
	var channelCap *capabilitytypes.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.OPEN,
				types.ORDERED, testConnectionIDA,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is CLOSED", func() {
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.CLOSED,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"connection not found", func() {
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.OPEN,
				types.ORDERED, testConnectionIDA,
			)
		}, false},
		{"connection is not OPEN", func() {
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.TRYOPEN,
			)
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.UNINITIALIZED,
				types.ORDERED, testConnectionIDA,
			)
		}, false},
		{"channel capability not found", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.OPEN,
				types.ORDERED, testConnectionIDA,
			)
			channelCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainA.App.ScopedIBCKeeper.NewCapability(suite.chainA.GetContext(), host.ChannelCapabilityPath(testPort1, testChannel1))
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()
			err = suite.chainA.App.IBCKeeper.ChannelKeeper.ChanCloseInit(
				suite.chainA.GetContext(), testPort1, testChannel1, channelCap,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanCloseConfirm() {
	channelKey := host.KeyChannel(testPort1, testChannel1)

	var channelCap *capabilitytypes.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA,
				connection.OPEN,
			)
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.OPEN,
				types.ORDERED, testConnectionIDB,
			)
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.CLOSED,
				types.ORDERED, testConnectionIDA,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is CLOSED", func() {
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.CLOSED,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"connection not found", func() {
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.OPEN,
				types.ORDERED, testConnectionIDA,
			)
		}, false},
		{"connection is not OPEN", func() {
			suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connection.TRYOPEN,
			)
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.OPEN,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"consensus state not found", func() {
			suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.OPEN,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"channel verification failed", func() {
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.OPEN,
				types.ORDERED, testConnectionIDB,
			)
		}, false},
		{"channel capability not found", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connection.OPEN,
			)
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA,
				connection.OPEN,
			)
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, types.OPEN,
				types.ORDERED, testConnectionIDB,
			)
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, types.CLOSED,
				types.ORDERED, testConnectionIDA,
			)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainB.App.ScopedIBCKeeper.NewCapability(suite.chainB.GetContext(), host.ChannelCapabilityPath(testPort2, testChannel2))
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()

			suite.chainA.updateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			proof, proofHeight := queryProof(suite.chainA, channelKey)

			if tc.expPass {
				err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanCloseConfirm(
					suite.chainB.GetContext(), testPort2, testChannel2, channelCap,
					proof, proofHeight+1,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanCloseConfirm(
					suite.chainB.GetContext(), testPort2, testChannel2, channelCap,
					proof, proofHeight,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}


*/
