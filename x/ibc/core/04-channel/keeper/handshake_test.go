package keeper_test

import (
	"fmt"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
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
		connA    *ibctesting.TestConnection
		connB    *ibctesting.TestConnection
		features []string
		portCap  *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			features = []string{"ORDER_ORDERED", "ORDER_UNORDERED"}
			suite.chainA.CreatePortCapability(suite.chainA.NextTestChannel(connA, ibctesting.MockPort).PortID)
			portCap = suite.chainA.GetPortCapability(suite.chainA.NextTestChannel(connA, ibctesting.MockPort).PortID)
		}, true},
		{"channel already exists", func() {
			_, _, connA, connB, _, _ = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
		}, false},
		{"connection doesn't exist", func() {
			// any non-nil values of connA and connB are acceptable
			suite.Require().NotNil(connA)
			suite.Require().NotNil(connB)
		}, false},
		{"capability is incorrect", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			features = []string{"ORDER_ORDERED", "ORDER_UNORDERED"}
			portCap = capabilitytypes.NewCapability(3)
		}, false},
		{"connection version not negotiated", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

			// modify connA versions
			conn := suite.chainA.GetConnection(connA)

			version := connectiontypes.NewVersion("2", []string{"ORDER_ORDERED", "ORDER_UNORDERED"})
			conn.Versions = append(conn.Versions, version)

			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(
				suite.chainA.GetContext(),
				connA.ID, conn,
			)
			features = []string{"ORDER_ORDERED", "ORDER_UNORDERED"}
			suite.chainA.CreatePortCapability(suite.chainA.NextTestChannel(connA, ibctesting.MockPort).PortID)
			portCap = suite.chainA.GetPortCapability(suite.chainA.NextTestChannel(connA, ibctesting.MockPort).PortID)
		}, false},
		{"connection does not support ORDERED channels", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

			// modify connA versions to only support UNORDERED channels
			conn := suite.chainA.GetConnection(connA)

			version := connectiontypes.NewVersion("1", []string{"ORDER_UNORDERED"})
			conn.Versions = []*connectiontypes.Version{version}

			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(
				suite.chainA.GetContext(),
				connA.ID, conn,
			)
			// NOTE: Opening UNORDERED channels is still expected to pass but ORDERED channels should fail
			features = []string{"ORDER_UNORDERED"}
			suite.chainA.CreatePortCapability(suite.chainA.NextTestChannel(connA, ibctesting.MockPort).PortID)
			portCap = suite.chainA.GetPortCapability(suite.chainA.NextTestChannel(connA, ibctesting.MockPort).PortID)
		}, true},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			// run test for all types of ordering
			for _, order := range []types.Order{types.UNORDERED, types.ORDERED} {
				suite.SetupTest() // reset
				tc.malleate()

				counterparty := types.NewCounterparty(connB.FirstOrNextTestChannel(ibctesting.MockPort).PortID, connB.FirstOrNextTestChannel(ibctesting.MockPort).ID)
				channelA := connA.FirstOrNextTestChannel(ibctesting.MockPort)

				channelID, cap, err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenInit(
					suite.chainA.GetContext(), order, []string{connA.ID},
					channelA.PortID, portCap, counterparty, channelA.Version,
				)

				// check if order is supported by channel to determine expected behaviour
				orderSupported := false
				for _, f := range features {
					if f == order.String() {
						orderSupported = true
					}
				}

				// Testcase must have expectedPass = true AND channel order supported before
				// asserting the channel handshake initiation succeeded
				if tc.expPass && orderSupported {
					suite.Require().NoError(err)
					suite.Require().NotNil(cap)
					suite.Require().Equal(types.FormatChannelIdentifier(0), channelID)

					chanCap, ok := suite.chainA.App.ScopedIBCKeeper.GetCapability(
						suite.chainA.GetContext(),
						host.ChannelCapabilityPath(channelA.PortID, channelA.ID),
					)
					suite.Require().True(ok, "could not retrieve channel capability after successful ChanOpenInit")
					suite.Require().Equal(chanCap.String(), cap.String(), "channel capability is not correct")
				} else {
					suite.Require().Error(err)
					suite.Require().Nil(cap)
					suite.Require().Equal("", channelID)
				}
			}
		})
	}
}

// TestChanOpenTry tests the OpenTry handshake call for channels. It uses message passing
// to enter into the appropriate state and then calls ChanOpenTry directly. The channel
// is being created on chainB. The port capability must be created on chainB before
// ChanOpenTry can succeed.
func (suite *KeeperTestSuite) TestChanOpenTry() {
	var (
		connA             *ibctesting.TestConnection
		connB             *ibctesting.TestConnection
		previousChannelID string
		portCap           *capabilitytypes.Capability
		heightDiff        uint64
	)

	testCases := []testCase{
		{"success", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)

			suite.chainB.CreatePortCapability(suite.chainB.NextTestChannel(connB, ibctesting.MockPort).PortID)
			portCap = suite.chainB.GetPortCapability(suite.chainB.NextTestChannel(connB, ibctesting.MockPort).PortID)
		}, true},
		{"success with crossing hello", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			_, channelB, err := suite.coordinator.ChanOpenInitOnBothChains(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			previousChannelID = channelB.ID
			portCap = suite.chainB.GetPortCapability(suite.chainB.NextTestChannel(connB, ibctesting.MockPort).PortID)
		}, true},
		{"previous channel with invalid state", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)

			// make previous channel have wrong ordering
			suite.coordinator.ChanOpenInit(suite.chainB, suite.chainA, connB, connA, ibctesting.MockPort, ibctesting.MockPort, types.UNORDERED)
		}, false},
		{"connection doesn't exist", func() {
			// any non-nil values of connA and connB are acceptable
			suite.Require().NotNil(connA)
			suite.Require().NotNil(connB)

			// pass capability check
			suite.chainB.CreatePortCapability(connB.FirstOrNextTestChannel(ibctesting.MockPort).PortID)
			portCap = suite.chainB.GetPortCapability(connB.FirstOrNextTestChannel(ibctesting.MockPort).PortID)
		}, false},
		{"connection is not OPEN", func() {
			clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			// pass capability check
			suite.chainB.CreatePortCapability(connB.FirstOrNextTestChannel(ibctesting.MockPort).PortID)
			portCap = suite.chainB.GetPortCapability(connB.FirstOrNextTestChannel(ibctesting.MockPort).PortID)

			var err error
			connB, connA, err = suite.coordinator.ConnOpenInit(suite.chainB, suite.chainA, clientB, clientA)
			suite.Require().NoError(err)
		}, false},
		{"consensus state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)

			suite.chainB.CreatePortCapability(suite.chainB.NextTestChannel(connB, ibctesting.MockPort).PortID)
			portCap = suite.chainB.GetPortCapability(suite.chainB.NextTestChannel(connB, ibctesting.MockPort).PortID)

			heightDiff = 3 // consensus state doesn't exist at this height
		}, false},
		{"channel verification failed", func() {
			// not creating a channel on chainA will result in an invalid proof of existence
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			portCap = suite.chainB.GetPortCapability(suite.chainB.NextTestChannel(connB, ibctesting.MockPort).PortID)
		}, false},
		{"port capability not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)

			portCap = capabilitytypes.NewCapability(3)
		}, false},
		{"connection version not negotiated", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)

			// modify connB versions
			conn := suite.chainB.GetConnection(connB)

			version := connectiontypes.NewVersion("2", []string{"ORDER_ORDERED", "ORDER_UNORDERED"})
			conn.Versions = append(conn.Versions, version)

			suite.chainB.App.IBCKeeper.ConnectionKeeper.SetConnection(
				suite.chainB.GetContext(),
				connB.ID, conn,
			)
			suite.chainB.CreatePortCapability(suite.chainB.NextTestChannel(connB, ibctesting.MockPort).PortID)
			portCap = suite.chainB.GetPortCapability(suite.chainB.NextTestChannel(connB, ibctesting.MockPort).PortID)
		}, false},
		{"connection does not support ORDERED channels", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)

			// modify connA versions to only support UNORDERED channels
			conn := suite.chainA.GetConnection(connA)

			version := connectiontypes.NewVersion("1", []string{"ORDER_UNORDERED"})
			conn.Versions = []*connectiontypes.Version{version}

			suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(
				suite.chainA.GetContext(),
				connA.ID, conn,
			)
			suite.chainA.CreatePortCapability(suite.chainA.NextTestChannel(connA, ibctesting.MockPort).PortID)
			portCap = suite.chainA.GetPortCapability(suite.chainA.NextTestChannel(connA, ibctesting.MockPort).PortID)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			heightDiff = 0    // must be explicitly changed in malleate
			previousChannelID = ""

			tc.malleate()
			channelA := connA.FirstOrNextTestChannel(ibctesting.MockPort)
			channelB := connB.FirstOrNextTestChannel(ibctesting.MockPort)
			counterparty := types.NewCounterparty(channelA.PortID, channelA.ID)

			channelKey := host.ChannelKey(counterparty.PortId, counterparty.ChannelId)
			proof, proofHeight := suite.chainA.QueryProof(channelKey)

			channelID, cap, err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanOpenTry(
				suite.chainB.GetContext(), types.ORDERED, []string{connB.ID},
				channelB.PortID, previousChannelID, portCap, counterparty, channelB.Version, connA.FirstOrNextTestChannel(ibctesting.MockPort).Version,
				proof, malleateHeight(proofHeight, heightDiff),
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(cap)

				chanCap, ok := suite.chainB.App.ScopedIBCKeeper.GetCapability(
					suite.chainB.GetContext(),
					host.ChannelCapabilityPath(channelB.PortID, channelID),
				)
				suite.Require().True(ok, "could not retrieve channel capapbility after successful ChanOpenTry")
				suite.Require().Equal(chanCap.String(), cap.String(), "channel capability is not correct")
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestChanOpenAck tests the OpenAck handshake call for channels. It uses message passing
// to enter into the appropriate state and then calls ChanOpenAck directly. The handshake
// call is occurring on chainA.
func (suite *KeeperTestSuite) TestChanOpenAck() {
	var (
		connA                 *ibctesting.TestConnection
		connB                 *ibctesting.TestConnection
		counterpartyChannelID string
		channelCap            *capabilitytypes.Capability
		heightDiff            uint64
	)

	testCases := []testCase{
		{"success", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"success with empty stored counterparty channel ID", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			// set the channel's counterparty channel identifier to empty string
			channel := suite.chainA.GetChannel(channelA)
			channel.Counterparty.ChannelId = ""

			// use a different channel identifier
			counterpartyChannelID = channelB.ID

			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ID, channel)

			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not INIT or TRYOPEN", func() {
			// create fully open channels on both chains
			_, _, connA, connB, _, _ = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelA := connA.Channels[0]
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"connection not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)

			// set the channel's connection hops to wrong connection ID
			channel := suite.chainA.GetChannel(channelA)
			channel.ConnectionHops[0] = "doesnotexist"
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ID, channel)
		}, false},
		{"connection is not OPEN", func() {
			clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)

			var err error
			connA, connB, err = suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// create channel in init
			channelA, _, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"consensus state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)

			heightDiff = 3 // consensus state doesn't exist at this height
		}, false},
		{"invalid counterparty channel identifier", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			counterpartyChannelID = "otheridentifier"

			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel verification failed", func() {
			// chainB is INIT, chainA in TRYOPEN
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelB, channelA, err := suite.coordinator.ChanOpenInit(suite.chainB, suite.chainA, connB, connA, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainA, suite.chainB, channelA, channelB, connA, types.ORDERED)
			suite.Require().NoError(err)

			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel capability not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)

			channelCap = capabilitytypes.NewCapability(6)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()          // reset
			counterpartyChannelID = "" // must be explicitly changed in malleate
			heightDiff = 0             // must be explicitly changed

			tc.malleate()

			channelA := connA.FirstOrNextTestChannel(ibctesting.MockPort)
			channelB := connB.FirstOrNextTestChannel(ibctesting.MockPort)

			if counterpartyChannelID == "" {
				counterpartyChannelID = channelB.ID
			}

			channelKey := host.ChannelKey(channelB.PortID, channelB.ID)
			proof, proofHeight := suite.chainB.QueryProof(channelKey)

			err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanOpenAck(
				suite.chainA.GetContext(), channelA.PortID, channelA.ID, channelCap, channelB.Version, counterpartyChannelID,
				proof, malleateHeight(proofHeight, heightDiff),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestChanOpenConfirm tests the OpenAck handshake call for channels. It uses message passing
// to enter into the appropriate state and then calls ChanOpenConfirm directly. The handshake
// call is occurring on chainB.
func (suite *KeeperTestSuite) TestChanOpenConfirm() {
	var (
		connA      *ibctesting.TestConnection
		connB      *ibctesting.TestConnection
		channelCap *capabilitytypes.Capability
		heightDiff uint64
	)
	testCases := []testCase{
		{"success", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenAck(suite.chainA, suite.chainB, channelA, channelB)
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not TRYOPEN", func() {
			// create fully open channels on both cahins
			_, _, connA, connB, _, _ = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelB := connB.Channels[0]
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"connection not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenAck(suite.chainA, suite.chainB, channelA, channelB)
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)

			// set the channel's connection hops to wrong connection ID
			channel := suite.chainB.GetChannel(channelB)
			channel.ConnectionHops[0] = "doesnotexist"
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(suite.chainB.GetContext(), channelB.PortID, channelB.ID, channel)
		}, false},
		{"connection is not OPEN", func() {
			clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)

			var err error
			connA, connB, err = suite.coordinator.ConnOpenInit(suite.chainB, suite.chainA, clientB, clientA)
			suite.Require().NoError(err)
			channelB := connB.FirstOrNextTestChannel(ibctesting.MockPort)
			suite.chainB.CreateChannelCapability(channelB.PortID, channelB.ID)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"consensus state not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenAck(suite.chainA, suite.chainB, channelA, channelB)
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)

			heightDiff = 3
		}, false},
		{"channel verification failed", func() {
			// chainA is INIT, chainB in TRYOPEN
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"channel capability not found", func() {
			_, _, connA, connB = suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
			channelA, channelB, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenTry(suite.chainB, suite.chainA, channelB, channelA, connB, types.ORDERED)
			suite.Require().NoError(err)

			err = suite.coordinator.ChanOpenAck(suite.chainA, suite.chainB, channelA, channelB)
			suite.Require().NoError(err)

			channelCap = capabilitytypes.NewCapability(6)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			heightDiff = 0    // must be explicitly changed

			tc.malleate()

			channelA := connA.FirstOrNextTestChannel(ibctesting.MockPort)
			channelB := connB.FirstOrNextTestChannel(ibctesting.MockPort)

			channelKey := host.ChannelKey(channelA.PortID, channelA.ID)
			proof, proofHeight := suite.chainA.QueryProof(channelKey)

			err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanOpenConfirm(
				suite.chainB.GetContext(), channelB.PortID, channelB.ID,
				channelCap, proof, malleateHeight(proofHeight, heightDiff),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestChanCloseInit tests the initial closing of a handshake on chainA by calling
// ChanCloseInit. Both chains will use message passing to setup OPEN channels.
func (suite *KeeperTestSuite) TestChanCloseInit() {
	var (
		connA      *ibctesting.TestConnection
		connB      *ibctesting.TestConnection
		channelCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success", func() {
			_, _, connA, connB, _, _ = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelA := connA.Channels[0]
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, true},
		{"channel doesn't exist", func() {
			// any non-nil values work for connections
			suite.Require().NotNil(connA)
			suite.Require().NotNil(connB)
			channelA := connA.FirstOrNextTestChannel(ibctesting.MockPort)

			// ensure channel capability check passes
			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel state is CLOSED", func() {
			_, _, connA, connB, _, _ = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelA := connA.Channels[0]
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)

			// close channel
			err := suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.Require().NoError(err)
		}, false},
		{"connection not found", func() {
			_, _, connA, connB, _, _ = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelA := connA.Channels[0]
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)

			// set the channel's connection hops to wrong connection ID
			channel := suite.chainA.GetChannel(channelA)
			channel.ConnectionHops[0] = "doesnotexist"
			suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ID, channel)
		}, false},
		{"connection is not OPEN", func() {
			clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)

			var err error
			connA, connB, err = suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
			suite.Require().NoError(err)

			// create channel in init
			channelA, _, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)

			// ensure channel capability check passes
			suite.chainA.CreateChannelCapability(channelA.PortID, channelA.ID)
			channelCap = suite.chainA.GetChannelCapability(channelA.PortID, channelA.ID)
		}, false},
		{"channel capability not found", func() {
			_, _, connA, connB, _, _ = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			channelA := connA.FirstOrNextTestChannel(ibctesting.MockPort)

			err := suite.chainA.App.IBCKeeper.ChannelKeeper.ChanCloseInit(
				suite.chainA.GetContext(), channelA.PortID, channelA.ID, channelCap,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestChanCloseConfirm tests the confirming closing channel ends by calling ChanCloseConfirm
// on chainB. Both chains will use message passing to setup OPEN channels. ChanCloseInit is
// bypassed on chainA by setting the channel state in the ChannelKeeper.
func (suite *KeeperTestSuite) TestChanCloseConfirm() {
	var (
		connA      *ibctesting.TestConnection
		connB      *ibctesting.TestConnection
		channelA   ibctesting.TestChannel
		channelB   ibctesting.TestChannel
		channelCap *capabilitytypes.Capability
		heightDiff uint64
	)

	testCases := []testCase{
		{"success", func() {
			_, _, connA, connB, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)

			err := suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.Require().NoError(err)
		}, true},
		{"channel doesn't exist", func() {
			// any non-nil values work for connections
			suite.Require().NotNil(connA)
			suite.Require().NotNil(connB)
			channelB = connB.FirstOrNextTestChannel(ibctesting.MockPort)

			// ensure channel capability check passes
			suite.chainB.CreateChannelCapability(channelB.PortID, channelB.ID)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"channel state is CLOSED", func() {
			_, _, connA, connB, _, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)

			err := suite.coordinator.SetChannelClosed(suite.chainB, suite.chainA, channelB)
			suite.Require().NoError(err)
		}, false},
		{"connection not found", func() {
			_, _, connA, connB, _, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)

			// set the channel's connection hops to wrong connection ID
			channel := suite.chainB.GetChannel(channelB)
			channel.ConnectionHops[0] = "doesnotexist"
			suite.chainB.App.IBCKeeper.ChannelKeeper.SetChannel(suite.chainB.GetContext(), channelB.PortID, channelB.ID, channel)
		}, false},
		{"connection is not OPEN", func() {
			clientA, clientB := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)

			var err error
			connB, connA, err = suite.coordinator.ConnOpenInit(suite.chainB, suite.chainA, clientB, clientA)
			suite.Require().NoError(err)

			// create channel in init
			channelB, _, err := suite.coordinator.ChanOpenInit(suite.chainB, suite.chainA, connB, connA, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
			suite.Require().NoError(err)

			// ensure channel capability check passes
			suite.chainB.CreateChannelCapability(channelB.PortID, channelB.ID)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"consensus state not found", func() {
			_, _, connA, connB, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)

			err := suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.Require().NoError(err)

			heightDiff = 3
		}, false},
		{"channel verification failed", func() {
			// channel not closed
			_, _, connA, connB, _, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)
		}, false},
		{"channel capability not found", func() {
			_, _, connA, connB, channelA, channelB = suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
			channelCap = suite.chainB.GetChannelCapability(channelB.PortID, channelB.ID)

			err := suite.coordinator.SetChannelClosed(suite.chainA, suite.chainB, channelA)
			suite.Require().NoError(err)

			channelCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			heightDiff = 0    // must explicitly be changed

			tc.malleate()

			channelA = connA.FirstOrNextTestChannel(ibctesting.MockPort)
			channelB = connB.FirstOrNextTestChannel(ibctesting.MockPort)

			channelKey := host.ChannelKey(channelA.PortID, channelA.ID)
			proof, proofHeight := suite.chainA.QueryProof(channelKey)

			err := suite.chainB.App.IBCKeeper.ChannelKeeper.ChanCloseConfirm(
				suite.chainB.GetContext(), channelB.PortID, channelB.ID, channelCap,
				proof, malleateHeight(proofHeight, heightDiff),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func malleateHeight(height exported.Height, diff uint64) exported.Height {
	return clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight()+diff)
}
