package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *KeeperTestSuite) TestQueryChannel() {
	var (
		req        *types.QueryChannelRequest
		expChannel types.Channel
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid port ID",
			func() {
				req = &types.QueryChannelRequest{
					PortId:    "",
					ChannelId: "test-channel-id",
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryChannelRequest{
					PortId:    "test-port-id",
					ChannelId: "",
				}
			},
			false,
		},
		{"channel not found",
			func() {
				req = &types.QueryChannelRequest{
					PortId:    "test-port-id",
					ChannelId: "test-channel-id",
				}
			},
			false,
		},
		{
			"success",
			func() {
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
				// init channel
				channelA, _, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
				suite.Require().NoError(err)

				expChannel = suite.chainA.GetChannel(channelA)

				req = &types.QueryChannelRequest{
					PortId:    channelA.PortID,
					ChannelId: channelA.ID,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.Channel(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(&expChannel, res.Channel)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryChannels() {
	var (
		req         *types.QueryChannelsRequest
		expChannels = []*types.IdentifiedChannel{}
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"empty pagination",
			func() {
				req = &types.QueryChannelsRequest{}
			},
			true,
		},
		{
			"success",
			func() {
				_, _, connA0, connB0, testchannel0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
				// channel0 on first connection on chainA
				counterparty0 := types.Counterparty{
					PortId:    connB0.Channels[0].PortID,
					ChannelId: connB0.Channels[0].ID,
				}

				// channel1 is second channel on first connection on chainA
				testchannel1, _ := suite.coordinator.CreateMockChannels(suite.chainA, suite.chainB, connA0, connB0, types.ORDERED)
				counterparty1 := types.Counterparty{
					PortId:    connB0.Channels[1].PortID,
					ChannelId: connB0.Channels[1].ID,
				}

				channel0 := types.NewChannel(
					types.OPEN, types.UNORDERED,
					counterparty0, []string{connA0.ID}, testchannel0.Version,
				)
				channel1 := types.NewChannel(
					types.OPEN, types.ORDERED,
					counterparty1, []string{connA0.ID}, testchannel1.Version,
				)

				idCh0 := types.NewIdentifiedChannel(testchannel0.PortID, testchannel0.ID, channel0)
				idCh1 := types.NewIdentifiedChannel(testchannel1.PortID, testchannel1.ID, channel1)

				expChannels = []*types.IdentifiedChannel{&idCh0, &idCh1}

				req = &types.QueryChannelsRequest{
					Pagination: &query.PageRequest{
						Key:        nil,
						Limit:      2,
						CountTotal: true,
					},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.Channels(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expChannels, res.Channels)
				suite.Require().Equal(len(expChannels), int(res.Pagination.Total))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryConnectionChannels() {
	var (
		req         *types.QueryConnectionChannelsRequest
		expChannels = []*types.IdentifiedChannel{}
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid connection ID",
			func() {
				req = &types.QueryConnectionChannelsRequest{
					Connection: "",
				}
			},
			false,
		},
		{
			"success",
			func() {
				_, _, connA0, connB0, testchannel0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
				// channel0 on first connection on chainA
				counterparty0 := types.Counterparty{
					PortId:    connB0.Channels[0].PortID,
					ChannelId: connB0.Channels[0].ID,
				}

				// channel1 is second channel on first connection on chainA
				testchannel1, _ := suite.coordinator.CreateMockChannels(suite.chainA, suite.chainB, connA0, connB0, types.ORDERED)
				counterparty1 := types.Counterparty{
					PortId:    connB0.Channels[1].PortID,
					ChannelId: connB0.Channels[1].ID,
				}

				channel0 := types.NewChannel(
					types.OPEN, types.UNORDERED,
					counterparty0, []string{connA0.ID}, testchannel0.Version,
				)
				channel1 := types.NewChannel(
					types.OPEN, types.ORDERED,
					counterparty1, []string{connA0.ID}, testchannel1.Version,
				)

				idCh0 := types.NewIdentifiedChannel(testchannel0.PortID, testchannel0.ID, channel0)
				idCh1 := types.NewIdentifiedChannel(testchannel1.PortID, testchannel1.ID, channel1)

				expChannels = []*types.IdentifiedChannel{&idCh0, &idCh1}

				req = &types.QueryConnectionChannelsRequest{
					Connection: connA0.ID,
					Pagination: &query.PageRequest{
						Key:        nil,
						Limit:      2,
						CountTotal: true,
					},
				}
			},
			true,
		},
		{
			"success, empty response",
			func() {
				suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
				expChannels = []*types.IdentifiedChannel{}
				req = &types.QueryConnectionChannelsRequest{
					Connection: "externalConnID",
					Pagination: &query.PageRequest{
						Key:        nil,
						Limit:      2,
						CountTotal: false,
					},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.ConnectionChannels(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expChannels, res.Channels)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryChannelClientState() {
	var (
		req                      *types.QueryChannelClientStateRequest
		expIdentifiedClientState clienttypes.IdentifiedClientState
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid port ID",
			func() {
				req = &types.QueryChannelClientStateRequest{
					PortId:    "",
					ChannelId: "test-channel-id",
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryChannelClientStateRequest{
					PortId:    "test-port-id",
					ChannelId: "",
				}
			},
			false,
		},
		{
			"channel not found",
			func() {
				req = &types.QueryChannelClientStateRequest{
					PortId:    "test-port-id",
					ChannelId: "test-channel-id",
				}
			},
			false,
		},
		{
			"connection not found",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

				channel := suite.chainA.GetChannel(channelA)
				// update channel to reference a connection that does not exist
				channel.ConnectionHops[0] = "doesnotexist"

				// set connection hops to wrong connection ID
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ID, channel)

				req = &types.QueryChannelClientStateRequest{
					PortId:    channelA.PortID,
					ChannelId: channelA.ID,
				}
			}, false,
		},
		{
			"client state for channel's connection not found",
			func() {
				_, _, connA, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

				// set connection to empty so clientID is empty
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, connectiontypes.ConnectionEnd{})

				req = &types.QueryChannelClientStateRequest{
					PortId:    channelA.PortID,
					ChannelId: channelA.ID,
				}
			}, false,
		},
		{
			"success",
			func() {
				clientA, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
				// init channel
				channelA, _, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
				suite.Require().NoError(err)

				expClientState := suite.chainA.GetClientState(clientA)
				expIdentifiedClientState = clienttypes.NewIdentifiedClientState(clientA, expClientState)

				req = &types.QueryChannelClientStateRequest{
					PortId:    channelA.PortID,
					ChannelId: channelA.ID,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.ChannelClientState(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(&expIdentifiedClientState, res.IdentifiedClientState)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryChannelConsensusState() {
	var (
		req               *types.QueryChannelConsensusStateRequest
		expConsensusState exported.ConsensusState
		expClientID       string
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid port ID",
			func() {
				req = &types.QueryChannelConsensusStateRequest{
					PortId:        "",
					ChannelId:     "test-channel-id",
					VersionNumber: 0,
					VersionHeight: 1,
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryChannelConsensusStateRequest{
					PortId:        "test-port-id",
					ChannelId:     "",
					VersionNumber: 0,
					VersionHeight: 1,
				}
			},
			false,
		},
		{
			"channel not found",
			func() {
				req = &types.QueryChannelConsensusStateRequest{
					PortId:        "test-port-id",
					ChannelId:     "test-channel-id",
					VersionNumber: 0,
					VersionHeight: 1,
				}
			},
			false,
		},
		{
			"connection not found",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

				channel := suite.chainA.GetChannel(channelA)
				// update channel to reference a connection that does not exist
				channel.ConnectionHops[0] = "doesnotexist"

				// set connection hops to wrong connection ID
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ID, channel)

				req = &types.QueryChannelConsensusStateRequest{
					PortId:        channelA.PortID,
					ChannelId:     channelA.ID,
					VersionNumber: 0,
					VersionHeight: 1,
				}
			}, false,
		},
		{
			"consensus state for channel's connection not found",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

				req = &types.QueryChannelConsensusStateRequest{
					PortId:        channelA.PortID,
					ChannelId:     channelA.ID,
					VersionNumber: 0,
					VersionHeight: uint64(suite.chainA.GetContext().BlockHeight()), // use current height
				}
			}, false,
		},
		{
			"success",
			func() {
				clientA, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
				// init channel
				channelA, _, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, ibctesting.MockPort, ibctesting.MockPort, types.ORDERED)
				suite.Require().NoError(err)

				clientState := suite.chainA.GetClientState(clientA)
				expConsensusState, _ = suite.chainA.GetConsensusState(clientA, clientState.GetLatestHeight())
				suite.Require().NotNil(expConsensusState)
				expClientID = clientA

				req = &types.QueryChannelConsensusStateRequest{
					PortId:        channelA.PortID,
					ChannelId:     channelA.ID,
					VersionNumber: clientState.GetLatestHeight().GetVersionNumber(),
					VersionHeight: clientState.GetLatestHeight().GetVersionHeight(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.ChannelConsensusState(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				consensusState, err := clienttypes.UnpackConsensusState(res.ConsensusState)
				suite.Require().NoError(err)
				suite.Require().Equal(expConsensusState, consensusState)
				suite.Require().Equal(expClientID, res.ClientId)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryPacketCommitment() {
	var (
		req           *types.QueryPacketCommitmentRequest
		expCommitment []byte
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid port ID",
			func() {
				req = &types.QueryPacketCommitmentRequest{
					PortId:    "",
					ChannelId: "test-channel-id",
					Sequence:  0,
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryPacketCommitmentRequest{
					PortId:    "test-port-id",
					ChannelId: "",
					Sequence:  0,
				}
			},
			false,
		},
		{"invalid sequence",
			func() {
				req = &types.QueryPacketCommitmentRequest{
					PortId:    "test-port-id",
					ChannelId: "test-channel-id",
					Sequence:  0,
				}
			},
			false,
		},
		{"channel not found",
			func() {
				req = &types.QueryPacketCommitmentRequest{
					PortId:    "test-port-id",
					ChannelId: "test-channel-id",
					Sequence:  1,
				}
			},
			false,
		},
		{
			"success",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
				expCommitment = []byte("hash")
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 1, expCommitment)

				req = &types.QueryPacketCommitmentRequest{
					PortId:    channelA.PortID,
					ChannelId: channelA.ID,
					Sequence:  1,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.PacketCommitment(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expCommitment, res.Commitment)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryPacketCommitments() {
	var (
		req            *types.QueryPacketCommitmentsRequest
		expCommitments = []*types.PacketAckCommitment{}
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid ID",
			func() {
				req = &types.QueryPacketCommitmentsRequest{
					PortId:    "",
					ChannelId: "test-channel-id",
				}
			},
			false,
		},
		{
			"success, empty res",
			func() {
				expCommitments = []*types.PacketAckCommitment{}

				req = &types.QueryPacketCommitmentsRequest{
					PortId:    "test-port-id",
					ChannelId: "test-channel-id",
					Pagination: &query.PageRequest{
						Key:        nil,
						Limit:      2,
						CountTotal: true,
					},
				}
			},
			true,
		},
		{
			"success",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

				expCommitments = make([]*types.PacketAckCommitment, 9)

				for i := uint64(0); i < 9; i++ {
					commitment := types.NewPacketAckCommitment(channelA.PortID, channelA.ID, i, []byte(fmt.Sprintf("hash_%d", i)))
					suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), commitment.PortId, commitment.ChannelId, commitment.Sequence, commitment.Hash)
					expCommitments[i] = &commitment
				}

				req = &types.QueryPacketCommitmentsRequest{
					PortId:    channelA.PortID,
					ChannelId: channelA.ID,
					Pagination: &query.PageRequest{
						Key:        nil,
						Limit:      11,
						CountTotal: true,
					},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.PacketCommitments(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expCommitments, res.Commitments)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryPacketAcknowledgement() {
	var (
		req    *types.QueryPacketAcknowledgementRequest
		expAck []byte
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid port ID",
			func() {
				req = &types.QueryPacketAcknowledgementRequest{
					PortId:    "",
					ChannelId: "test-channel-id",
					Sequence:  0,
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryPacketAcknowledgementRequest{
					PortId:    "test-port-id",
					ChannelId: "",
					Sequence:  0,
				}
			},
			false,
		},
		{"invalid sequence",
			func() {
				req = &types.QueryPacketAcknowledgementRequest{
					PortId:    "test-port-id",
					ChannelId: "test-channel-id",
					Sequence:  0,
				}
			},
			false,
		},
		{"channel not found",
			func() {
				req = &types.QueryPacketAcknowledgementRequest{
					PortId:    "test-port-id",
					ChannelId: "test-channel-id",
					Sequence:  1,
				}
			},
			false,
		},
		{
			"success",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
				expAck = []byte("hash")
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 1, expAck)

				req = &types.QueryPacketAcknowledgementRequest{
					PortId:    channelA.PortID,
					ChannelId: channelA.ID,
					Sequence:  1,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.PacketAcknowledgement(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expAck, res.Acknowledgement)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryUnreceivedPackets() {
	var (
		req    *types.QueryUnreceivedPacketsRequest
		expSeq = []uint64{}
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid port ID",
			func() {
				req = &types.QueryUnreceivedPacketsRequest{
					PortId:    "",
					ChannelId: "test-channel-id",
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryUnreceivedPacketsRequest{
					PortId:    "test-port-id",
					ChannelId: "",
				}
			},
			false,
		},
		{
			"invalid seq",
			func() {
				req = &types.QueryUnreceivedPacketsRequest{
					PortId:                    "test-port-id",
					ChannelId:                 "test-channel-id",
					PacketCommitmentSequences: []uint64{0},
				}
			},
			false,
		},
		{
			"basic success unrelayed packet commitments",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

				// no ack exists

				expSeq = []uint64{1}
				req = &types.QueryUnreceivedPacketsRequest{
					PortId:                    channelA.PortID,
					ChannelId:                 channelA.ID,
					PacketCommitmentSequences: []uint64{1},
				}
			},
			true,
		},
		{
			"basic success unrelayed packet commitments, nothing to relay",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

				// ack exists
				ack := types.NewPacketAckCommitment(channelA.PortID, channelA.ID, 1, []byte("hash"))
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 1, ack.Hash)

				expSeq = []uint64{}
				req = &types.QueryUnreceivedPacketsRequest{
					PortId:                    channelA.PortID,
					ChannelId:                 channelA.ID,
					PacketCommitmentSequences: []uint64{1},
				}
			},
			true,
		},
		{
			"success multiple unrelayed packet commitments",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
				expSeq = []uint64{} // reset
				packetCommitments := []uint64{}

				// set ack for every other sequence
				for seq := uint64(1); seq < 10; seq++ {
					packetCommitments = append(packetCommitments, seq)

					if seq%2 == 0 {
						ack := types.NewPacketAckCommitment(channelA.PortID, channelA.ID, seq, []byte("hash"))
						suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), channelA.PortID, channelA.ID, seq, ack.Hash)
					} else {
						expSeq = append(expSeq, seq)
					}
				}

				req = &types.QueryUnreceivedPacketsRequest{
					PortId:                    channelA.PortID,
					ChannelId:                 channelA.ID,
					PacketCommitmentSequences: packetCommitments,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.UnreceivedPackets(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expSeq, res.Sequences)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryUnrelayedAcks() {
	var (
		req    *types.QueryUnrelayedAcksRequest
		expSeq = []uint64{}
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid port ID",
			func() {
				req = &types.QueryUnrelayedAcksRequest{
					PortId:    "",
					ChannelId: "test-channel-id",
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryUnrelayedAcksRequest{
					PortId:    "test-port-id",
					ChannelId: "",
				}
			},
			false,
		},
		{
			"invalid seq",
			func() {
				req = &types.QueryUnrelayedAcksRequest{
					PortId:                    "test-port-id",
					ChannelId:                 "test-channel-id",
					PacketCommitmentSequences: []uint64{0},
				}
			},
			false,
		},
		{
			"basic success unrelayed packet acks",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

				// ack exists
				ack := types.NewPacketAckCommitment(channelA.PortID, channelA.ID, 1, []byte("hash"))
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 1, ack.Hash)

				expSeq = []uint64{1}
				req = &types.QueryUnrelayedAcksRequest{
					PortId:                    channelA.PortID,
					ChannelId:                 channelA.ID,
					PacketCommitmentSequences: []uint64{1},
				}
			},
			true,
		},
		{
			"basic success unrelayed packet acknowledgements, nothing to relay",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)

				// no ack exists

				expSeq = []uint64{}
				req = &types.QueryUnrelayedAcksRequest{
					PortId:                    channelA.PortID,
					ChannelId:                 channelA.ID,
					PacketCommitmentSequences: []uint64{1},
				}
			},
			true,
		},
		{
			"success multiple unrelayed packet acknowledgements",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
				expSeq = []uint64{} // reset
				packetCommitments := []uint64{}

				// set ack for every other sequence
				for seq := uint64(1); seq < 10; seq++ {
					packetCommitments = append(packetCommitments, seq)

					if seq%2 == 0 {
						ack := types.NewPacketAckCommitment(channelA.PortID, channelA.ID, seq, []byte("hash"))
						suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), channelA.PortID, channelA.ID, seq, ack.Hash)
						expSeq = append(expSeq, seq)
					}
				}

				req = &types.QueryUnrelayedAcksRequest{
					PortId:                    channelA.PortID,
					ChannelId:                 channelA.ID,
					PacketCommitmentSequences: packetCommitments,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.UnrelayedAcks(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expSeq, res.Sequences)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryNextSequenceReceive() {
	var (
		req    *types.QueryNextSequenceReceiveRequest
		expSeq uint64
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{
			"invalid port ID",
			func() {
				req = &types.QueryNextSequenceReceiveRequest{
					PortId:    "",
					ChannelId: "test-channel-id",
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryNextSequenceReceiveRequest{
					PortId:    "test-port-id",
					ChannelId: "",
				}
			},
			false,
		},
		{"channel not found",
			func() {
				req = &types.QueryNextSequenceReceiveRequest{
					PortId:    "test-port-id",
					ChannelId: "test-channel-id",
				}
			},
			false,
		},
		{
			"success",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, types.UNORDERED)
				expSeq = 1
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), channelA.PortID, channelA.ID, expSeq)

				req = &types.QueryNextSequenceReceiveRequest{
					PortId:    channelA.PortID,
					ChannelId: channelA.ID,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.NextSequenceReceive(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expSeq, res.NextSequenceReceive)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
