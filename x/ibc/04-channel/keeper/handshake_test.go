package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/capability"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/05-port/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *KeeperTestSuite) TestChanOpenInit() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)

	var portCap *capability.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA,
				connectionexported.INIT,
			)
		}, true},
		{"channel already exists", func() {
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.INIT,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
		{"connection doesn't exist", func() {}, false},
		{"connection is UNINITIALIZED", func() {
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA,
				connectionexported.UNINITIALIZED,
			)
		}, false},
		{"capability is incorrect", func() {
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA,
				connectionexported.INIT,
			)
			portCap = capability.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			portCap, err = suite.chainA.App.ScopedIBCKeeper.NewCapability(
				suite.chainA.GetContext(), porttypes.PortPath(testPort1),
			)
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()
			cap, err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenInit(
				suite.chainA.GetContext(), exported.ORDERED, []string{testConnectionIDA},
				testPort1, testChannel1, portCap, counterparty, testChannelVersion,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().NotNil(cap)
				chanCap, ok := suite.chainA.App.ScopedIBCKeeper.GetCapability(
					suite.chainA.GetContext(),
					ibctypes.ChannelCapabilityPath(testPort1, testChannel1),
				)
				suite.Require().True(ok, "could not retrieve channel capapbility after successful ChanOpenInit")
				suite.Require().Equal(chanCap.String(), cap.String(), "channel capability is not correct")
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanOpenTry() {
	counterparty := types.NewCounterparty(testPort1, testChannel1)
	channelKey := ibctypes.KeyChannel(testPort1, testChannel1)

	var portCap *capability.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.INIT, exported.ORDERED, testConnectionIDA)
		}, true},
		{"previous channel with invalid state", func() {
			_ = suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.UNINITIALIZED,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"connection doesn't exist", func() {}, false},
		{"connection is not OPEN", func() {
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.INIT,
			)
		}, false},
		{"consensus state not found", func() {
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
		}, false},
		{"channel verification failed", func() {
			suite.chainA.CreateClient(suite.chainB)
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
		}, false},
		{"port capability not found", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB, connectionexported.OPEN)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2, exported.INIT, exported.ORDERED, testConnectionIDA)
			portCap = capability.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			portCap, err = suite.chainA.App.ScopedIBCKeeper.NewCapability(suite.chainA.GetContext(), porttypes.PortPath(testPort2))
			suite.Require().NoError(err, "could not create capability")

			tc.malleate()

			suite.chainA.updateClient(suite.chainB)
			suite.chainB.updateClient(suite.chainA)
			proof, proofHeight := queryProof(suite.chainB, channelKey)

			if tc.expPass {
				cap, err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenTry(
					suite.chainA.GetContext(), exported.ORDERED, []string{testConnectionIDB},
					testPort2, testChannel2, portCap, counterparty, testChannelVersion, testChannelVersion,
					proof, proofHeight+1,
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().NotNil(cap)
				chanCap, ok := suite.chainA.App.ScopedIBCKeeper.GetCapability(
					suite.chainA.GetContext(),
					ibctypes.ChannelCapabilityPath(testPort2, testChannel2),
				)
				suite.Require().True(ok, "could not retrieve channel capapbility after successful ChanOpenInit")
				suite.Require().Equal(chanCap.String(), cap.String(), "channel capability is not correct")
			} else {
				_, err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenTry(
					suite.chainA.GetContext(), exported.ORDERED, []string{testConnectionIDB},
					testPort2, testChannel2, portCap, counterparty, testChannelVersion, testChannelVersion,
					invalidProof{}, proofHeight,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanOpenAck() {
	channelKey := ibctypes.KeyChannel(testPort2, testChannel2)

	var channelCap *capability.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.INIT,
				exported.ORDERED, testConnectionIDB,
			)
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDA,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not INIT or TRYOPEN", func() {
			_ = suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.UNINITIALIZED,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
		{"connection not found", func() {
			_ = suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
		{"connection is not OPEN", func() {
			_ = suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.TRYOPEN,
			)
			_ = suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
		{"consensus state not found", func() {
			_ = suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
		{"channel verification failed", func() {
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
		{"channel capability not found", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.INIT,
				exported.ORDERED, testConnectionIDB,
			)
			suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDA,
			)
			channelCap = capability.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainA.App.ScopedIBCKeeper.NewCapability(suite.chainA.GetContext(), ibctypes.ChannelCapabilityPath(testPort1, testChannel1))
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
					invalidProof{}, proofHeight+1,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanOpenConfirm() {
	channelKey := ibctypes.KeyChannel(testPort2, testChannel2)

	var channelCap *capability.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.TRYOPEN,
			)
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionIDB,
			)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2,
				exported.TRYOPEN, exported.ORDERED, testConnectionIDA)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not TRYOPEN", func() {
			_ = suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.UNINITIALIZED,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"connection not found", func() {
			_ = suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"connection is not OPEN", func() {
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.TRYOPEN,
			)
			_ = suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"consensus state not found", func() {
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
			_ = suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"channel verification failed", func() {
			suite.chainA.CreateClient(suite.chainB)
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
			_ = suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"channel capability not found", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.TRYOPEN,
			)
			suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionIDB,
			)
			suite.chainB.createChannel(testPort1, testChannel1, testPort2, testChannel2,
				exported.TRYOPEN, exported.ORDERED, testConnectionIDA)
			channelCap = capability.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainB.App.ScopedIBCKeeper.NewCapability(suite.chainB.GetContext(), ibctypes.ChannelCapabilityPath(testPort1, testChannel1))
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
					invalidProof{}, proofHeight+1,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanCloseInit() {
	var channelCap *capability.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.OPEN,
				exported.ORDERED, testConnectionIDA,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is CLOSED", func() {
			_ = suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"connection not found", func() {
			_ = suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.OPEN,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
		{"connection is not OPEN", func() {
			_ = suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.TRYOPEN,
			)
			_ = suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.UNINITIALIZED,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
		{"channel capability not found", func() {
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.OPEN,
				exported.ORDERED, testConnectionIDA,
			)
			channelCap = capability.NewCapability(3)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainA.App.ScopedIBCKeeper.NewCapability(suite.chainA.GetContext(), ibctypes.ChannelCapabilityPath(testPort1, testChannel1))
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
	channelKey := ibctypes.KeyChannel(testPort1, testChannel1)

	var channelCap *capability.Capability
	testCases := []testCase{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionIDB,
			)
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED,
				exported.ORDERED, testConnectionIDA,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is CLOSED", func() {
			_ = suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.CLOSED,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"connection not found", func() {
			_ = suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
		{"connection is not OPEN", func() {
			_ = suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connectionexported.TRYOPEN,
			)
			_ = suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"consensus state not found", func() {
			_ = suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"channel verification failed", func() {
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionIDB,
			)
		}, false},
		{"channel capability not found", func() {
			suite.chainA.CreateClient(suite.chainB)
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			suite.chainA.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionIDB,
			)
			suite.chainA.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED,
				exported.ORDERED, testConnectionIDA,
			)
		}, false},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			var err error
			channelCap, err = suite.chainB.App.ScopedIBCKeeper.NewCapability(suite.chainB.GetContext(), ibctypes.ChannelCapabilityPath(testPort2, testChannel2))
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
					invalidProof{}, proofHeight,
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

type testCase = struct {
	msg      string
	malleate func()
	expPass  bool
}
