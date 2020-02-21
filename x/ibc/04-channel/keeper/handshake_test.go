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
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenInit(
				suite.chainA.GetContext(), exported.ORDERED, []string{testConnectionIDA},
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
			suite.chainA.CreateClient(suite.chainB)
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
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
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			proofHeight := suite.chainB.Header.Height

			if tc.expPass {
				err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenTry(
					suite.chainA.GetContext(), exported.ORDERED, []string{testConnectionIDB},
					testPort2, testChannel2, counterparty, testChannelVersion, testChannelVersion,
					validProof{}, uint64(proofHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenTry(
					suite.chainA.GetContext(), exported.ORDERED, []string{testConnectionIDB},
					testPort2, testChannel2, counterparty, testChannelVersion, testChannelVersion,
					invalidProof{}, uint64(proofHeight),
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanOpenAck() {
	testCases := []testCase{
		{"success", func() {
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainB.createConnection(
				testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createChannel(
				testPort1, testChannel1, testPort2, testChannel2, exported.TRYOPEN,
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
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			proofHeight := suite.chainA.Header.Height
			if tc.expPass {
				err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanOpenAck(
					suite.chainB.GetContext(), testPort1, testChannel1, testChannelVersion,
					validProof{}, uint64(proofHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanOpenAck(
					suite.chainB.GetContext(), testPort1, testChannel1, testChannelVersion,
					invalidProof{}, uint64(proofHeight),
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanOpenConfirm() {
	testCases := []testCase{
		{"success", func() {
			suite.chainA.CreateClient(suite.chainB)
			_ = suite.chainA.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
				connectionexported.OPEN,
			)
			_ = suite.chainA.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.TRYOPEN,
				exported.ORDERED, testConnectionIDB,
			)
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
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			proofHeight := suite.chainB.Header.Height
			if tc.expPass {
				err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenConfirm(
					suite.chainA.GetContext(), testPort2, testChannel2,
					validProof{}, uint64(proofHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenConfirm(
					suite.chainA.GetContext(), testPort2, testChannel2,
					invalidProof{}, uint64(proofHeight),
				)
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChanCloseInit() {
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
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanCloseInit(
				suite.chainA.GetContext(), testPort1, testChannel1,
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
			suite.chainB.CreateClient(suite.chainA)
			_ = suite.chainB.createConnection(
				testConnectionIDB, testConnectionIDA, testClientIDA, testClientIDB,
				connectionexported.OPEN,
			)
			_ = suite.chainB.createChannel(
				testPort2, testChannel2, testPort1, testChannel1, exported.OPEN,
				exported.ORDERED, testConnectionIDB,
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
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			proofHeight := suite.chainA.Header.Height
			if tc.expPass {
				err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanCloseConfirm(
					suite.chainB.GetContext(), testPort2, testChannel2,
					validProof{}, uint64(proofHeight),
				)
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanCloseConfirm(
					suite.chainB.GetContext(), testPort2, testChannel2,
					invalidProof{}, uint64(proofHeight),
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
