package keeper_test

import (
	"fmt"

	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

func (suite *KeeperTestSuite) TestChanOpenInit() {
	counterparty := types.NewCounterparty(testPort2, testChannel2)

	testCases := []testCase{
		{"success", func() {
			suite.createConnection(
				testConnectionID1, testConnectionID2, testClientID1, testClientID2,
				connectionexported.INIT,
			)
		}, true},
		{"channel already exists", func() {
			suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.INIT,
				exported.ORDERED, testConnectionID1,
			)
		}, false},
		{"connection doesn't exist", func() {}, false},
		{"connection is UNINITIALIZED", func() {
			suite.createConnection(
				testConnectionID1, testConnectionID2, testClientID1, testClientID2,
				connectionexported.UNINITIALIZED,
			)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenInit(
				suite.ctx, exported.ORDERED, []string{testConnectionID1},
				testPort1, testChannel1, counterparty, testChannelVersion,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanOpenTry() {
	counterparty := types.NewCounterparty(testPort1, testChannel1)

	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID2)
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.OPEN,
			)
		}, true},
		{"previous channel with invalid state", func() {
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.UNINITIALIZED,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
		{"connection doesn't exist", func() {}, false},
		{"connection is not OPEN", func() {
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.INIT,
			)
		}, false},
		{"consensus state not found", func() {
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.OPEN,
			)
		}, false},
		{"channel verification failed", func() {
			suite.createClient(testClientID2)
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.OPEN,
			)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			if tc.expPass {
				err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenTry(
					suite.ctx, exported.ORDERED, []string{testConnectionID2},
					testPort2, testChannel2, counterparty, testChannelVersion, testChannelVersion,
					validProof{}, uint64(testHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenTry(
					suite.ctx, exported.ORDERED, []string{testConnectionID2},
					testPort2, testChannel2, counterparty, testChannelVersion, testChannelVersion,
					invalidProof{}, uint64(testHeight),
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanOpenAck() {
	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID1)
			_ = suite.createConnection(
				testConnectionID1, testConnectionID2, testClientID1, testClientID2,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
				exported.ORDERED, testConnectionID1,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not INIT or TRYOPEN", func() {
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.UNINITIALIZED,
				exported.ORDERED, testConnectionID1,
			)
		}, false},
		{"connection not found", func() {
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
				exported.ORDERED, testConnectionID1,
			)
		}, false},
		{"connection is not OPEN", func() {
			_ = suite.createConnection(
				testConnectionID1, testConnectionID2, testClientID1, testClientID2,
				connectionexported.TRYOPEN,
			)
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
				exported.ORDERED, testConnectionID1,
			)
		}, false},
		{"consensus state not found", func() {
			_ = suite.createConnection(
				testConnectionID1, testConnectionID2, testClientID1, testClientID2,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
				exported.ORDERED, testConnectionID1,
			)
		}, false},
		{"channel verification failed", func() {
			suite.createClient(testClientID1)
			_ = suite.createConnection(
				testConnectionID1, testConnectionID2, testClientID1, testClientID2,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
				exported.ORDERED, testConnectionID1,
			)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			if tc.expPass {
				err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenAck(
					suite.ctx, testPort1, testChannel1, testChannelVersion,
					validProof{}, uint64(testHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenAck(
					suite.ctx, testPort1, testChannel1, testChannelVersion,
					invalidProof{}, uint64(testHeight),
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanOpenConfirm() {
	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID2)
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionID2,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not TRYOPEN", func() {
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.UNINITIALIZED,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
		{"connection not found", func() {
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
		{"connection is not OPEN", func() {
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.TRYOPEN,
			)
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
		{"consensus state not found", func() {
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
		{"channel verification failed", func() {
			suite.createClient(testClientID2)
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			if tc.expPass {
				err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenConfirm(
					suite.ctx, testPort2, testChannel2,
					validProof{}, uint64(testHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ChannelKeeper.ChanOpenConfirm(
					suite.ctx, testPort2, testChannel2,
					invalidProof{}, uint64(testHeight),
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanCloseInit() {
	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID1)
			_ = suite.createConnection(
				testConnectionID1, testConnectionID2, testClientID1, testClientID2,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.OPEN,
				exported.ORDERED, testConnectionID1,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is CLOSED", func() {
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.CLOSED,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
		{"connection not found", func() {
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.OPEN,
				exported.ORDERED, testConnectionID1,
			)
		}, false},
		{"connection is not OPEN", func() {
			_ = suite.createConnection(
				testConnectionID1, testConnectionID2, testClientID1, testClientID2,
				connectionexported.TRYOPEN,
			)
			_ = suite.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.UNINITIALIZED,
				exported.ORDERED, testConnectionID1,
			)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			err := suite.app.IBCKeeper.ChannelKeeper.ChanCloseInit(
				suite.ctx, testPort1, testChannel1,
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
	testCases := []testCase{
		{"success", func() {
			suite.createClient(testClientID2)
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionID2,
			)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is CLOSED", func() {
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.CLOSED,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
		{"connection not found", func() {
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionID1,
			)
		}, false},
		{"connection is not OPEN", func() {
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.TRYOPEN,
			)
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
		{"consensus state not found", func() {
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
		{"channel verification failed", func() {
			suite.createClient(testClientID2)
			_ = suite.createConnection(
				testConnectionID2, testConnectionID1, testClientID2, testClientID1,
				connectionexported.OPEN,
			)
			_ = suite.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionID2,
			)
		}, false},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			if tc.expPass {
				err := suite.app.IBCKeeper.ChannelKeeper.ChanCloseConfirm(
					suite.ctx, testPort2, testChannel2,
					validProof{}, uint64(testHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.app.IBCKeeper.ChannelKeeper.ChanCloseConfirm(
					suite.ctx, testPort2, testChannel2,
					invalidProof{}, uint64(testHeight),
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
