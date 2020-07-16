package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
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
					PortID:    "",
					ChannelID: "test-channel-id",
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryChannelRequest{
					PortID:    "test-port-id",
					ChannelID: "",
				}
			},
			false,
		},
		{"channel not found",
			func() {
				req = &types.QueryChannelRequest{
					PortID:    "test-port-id",
					ChannelID: "test-channel-id",
				}
			},
			false,
		},
		{
			"success",
			func() {
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
				// init channel
				channelA, _, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
				suite.Require().NoError(err)

				expChannel = suite.chainA.GetChannel(channelA)

				req = &types.QueryChannelRequest{
					PortID:    channelA.PortID,
					ChannelID: channelA.ID,
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
				_, _, connA0, connB0, testchannel0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)
				// channel0 on first connection on chainA
				counterparty0 := types.Counterparty{
					PortID:    connB0.Channels[0].PortID,
					ChannelID: connB0.Channels[0].ID,
				}

				// channel1 is second channel on first connection on chainA
				testchannel1, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA0, connB0, types.ORDERED)
				counterparty1 := types.Counterparty{
					PortID:    connB0.Channels[1].PortID,
					ChannelID: connB0.Channels[1].ID,
				}

				channel0 := types.NewChannel(
					types.OPEN, types.UNORDERED,
					counterparty0, []string{connA0.ID}, ibctesting.ChannelVersion,
				)
				channel1 := types.NewChannel(
					types.OPEN, types.ORDERED,
					counterparty1, []string{connA0.ID}, ibctesting.ChannelVersion,
				)

				idCh0 := types.NewIdentifiedChannel(testchannel0.PortID, testchannel0.ID, channel0)
				idCh1 := types.NewIdentifiedChannel(testchannel1.PortID, testchannel1.ID, channel1)

				expChannels = []*types.IdentifiedChannel{&idCh0, &idCh1}

				req = &types.QueryChannelsRequest{
					Req: &query.PageRequest{
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
				suite.Require().Equal(len(expChannels), int(res.Res.Total))
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
				_, _, connA0, connB0, testchannel0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)
				// channel0 on first connection on chainA
				counterparty0 := types.Counterparty{
					PortID:    connB0.Channels[0].PortID,
					ChannelID: connB0.Channels[0].ID,
				}

				// channel1 is second channel on first connection on chainA
				testchannel1, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA0, connB0, types.ORDERED)
				counterparty1 := types.Counterparty{
					PortID:    connB0.Channels[1].PortID,
					ChannelID: connB0.Channels[1].ID,
				}

				channel0 := types.NewChannel(
					types.OPEN, types.UNORDERED,
					counterparty0, []string{connA0.ID}, ibctesting.ChannelVersion,
				)
				channel1 := types.NewChannel(
					types.OPEN, types.ORDERED,
					counterparty1, []string{connA0.ID}, ibctesting.ChannelVersion,
				)

				idCh0 := types.NewIdentifiedChannel(testchannel0.PortID, testchannel0.ID, channel0)
				idCh1 := types.NewIdentifiedChannel(testchannel1.PortID, testchannel1.ID, channel1)

				expChannels = []*types.IdentifiedChannel{&idCh0, &idCh1}

				req = &types.QueryConnectionChannelsRequest{
					Connection: connB0.ID,
					Req: &query.PageRequest{
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
				suite.coordinator.Setup(suite.chainA, suite.chainB)
				expChannels = []*types.IdentifiedChannel{}
				req = &types.QueryConnectionChannelsRequest{
					Connection: "externalConnID",
					Req: &query.PageRequest{
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
					PortID:    "",
					ChannelID: "test-channel-id",
					Sequence:  0,
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryPacketCommitmentRequest{
					PortID:    "test-port-id",
					ChannelID: "",
					Sequence:  0,
				}
			},
			false,
		},
		{"invalid sequence",
			func() {
				req = &types.QueryPacketCommitmentRequest{
					PortID:    "test-port-id",
					ChannelID: "test-channel-id",
					Sequence:  0,
				}
			},
			false,
		},
		{"channel not found",
			func() {
				req = &types.QueryPacketCommitmentRequest{
					PortID:    "test-port-id",
					ChannelID: "test-channel-id",
					Sequence:  1,
				}
			},
			false,
		},
		{
			"success",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)
				expCommitment = []byte("hash")
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 1, expCommitment)

				req = &types.QueryPacketCommitmentRequest{
					PortID:    channelA.PortID,
					ChannelID: channelA.ID,
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
					PortID:    "",
					ChannelID: "test-channel-id",
				}
			},
			false,
		},
		{
			"success, empty res",
			func() {
				expCommitments = []*types.PacketAckCommitment{}

				req = &types.QueryPacketCommitmentsRequest{
					PortID:    "test-port-id",
					ChannelID: "test-channel-id",
					Req: &query.PageRequest{
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
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				expCommitments = make([]*types.PacketAckCommitment, 9)

				for i := uint64(0); i < 9; i++ {
					commitment := types.NewPacketAckCommitment(channelA.PortID, channelA.ID, i, []byte(fmt.Sprintf("hash_%d", i)))
					suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), commitment.PortID, commitment.ChannelID, commitment.Sequence, commitment.Hash)
					expCommitments[i] = &commitment
				}

				req = &types.QueryPacketCommitmentsRequest{
					PortID:    channelA.PortID,
					ChannelID: channelA.ID,
					Req: &query.PageRequest{
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

func (suite *KeeperTestSuite) TestQueryUnrelayedPackets() {
	var (
		req    *types.QueryUnrelayedPacketsRequest
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
				req = &types.QueryUnrelayedPacketsRequest{
					PortID:    "",
					ChannelID: "test-channel-id",
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryUnrelayedPacketsRequest{
					PortID:    "test-port-id",
					ChannelID: "",
				}
			},
			false,
		},
		{
			"invalid seq",
			func() {
				req = &types.QueryUnrelayedPacketsRequest{
					PortID:                    "test-port-id",
					ChannelID:                 "test-channel-id",
					PacketCommitmentSequences: []uint64{0},
				}
			},
			false,
		},
		{
			"basic success unrelayed packet commitments",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				// no ack exists

				expSeq = []uint64{1}
				req = &types.QueryUnrelayedPacketsRequest{
					PortID:                    channelA.PortID,
					ChannelID:                 channelA.ID,
					PacketCommitmentSequences: []uint64{1},
					Acknowledgements:          false,
				}
			},
			true,
		},
		{
			"basic success unrelayed packet commitments, nothing to relay",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				// ack exists
				ack := types.NewPacketAckCommitment(channelA.PortID, channelA.ID, 1, []byte("hash"))
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 1, ack.Hash)

				expSeq = []uint64{}
				req = &types.QueryUnrelayedPacketsRequest{
					PortID:                    channelA.PortID,
					ChannelID:                 channelA.ID,
					PacketCommitmentSequences: []uint64{1},
					Acknowledgements:          false,
				}
			},
			true,
		},
		{
			"basic success unrelayed acknowledgements",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				// ack exists
				ack := types.NewPacketAckCommitment(channelA.PortID, channelA.ID, 1, []byte("hash"))
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), channelA.PortID, channelA.ID, 1, ack.Hash)

				expSeq = []uint64{1}
				req = &types.QueryUnrelayedPacketsRequest{
					PortID:                    channelA.PortID,
					ChannelID:                 channelA.ID,
					PacketCommitmentSequences: []uint64{1},
					Acknowledgements:          true,
				}
			},
			true,
		},
		{
			"basic success unrelayed acknowledgements, nothing to relay",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				// no ack exists

				expSeq = []uint64{}
				req = &types.QueryUnrelayedPacketsRequest{
					PortID:                    channelA.PortID,
					ChannelID:                 channelA.ID,
					PacketCommitmentSequences: []uint64{1},
					Acknowledgements:          true,
				}
			},
			true,
		},
		{
			"success multiple unrelayed packet commitments",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)
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

				req = &types.QueryUnrelayedPacketsRequest{
					PortID:                    channelA.PortID,
					ChannelID:                 channelA.ID,
					PacketCommitmentSequences: packetCommitments,
					Acknowledgements:          false,
				}
			},
			true,
		},
		{
			"success multiple unrelayed acknowledgements",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)
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

				req = &types.QueryUnrelayedPacketsRequest{
					PortID:                    channelA.PortID,
					ChannelID:                 channelA.ID,
					PacketCommitmentSequences: packetCommitments,
					Acknowledgements:          true,
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

			res, err := suite.chainA.QueryServer.UnrelayedPackets(ctx, req)

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
					PortID:    "",
					ChannelID: "test-channel-id",
				}
			},
			false,
		},
		{
			"invalid channel ID",
			func() {
				req = &types.QueryNextSequenceReceiveRequest{
					PortID:    "test-port-id",
					ChannelID: "",
				}
			},
			false,
		},
		{"channel not found",
			func() {
				req = &types.QueryNextSequenceReceiveRequest{
					PortID:    "test-port-id",
					ChannelID: "test-channel-id",
				}
			},
			false,
		},
		{
			"success",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)
				expSeq = 1
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.chainA.GetContext(), channelA.PortID, channelA.ID, expSeq)

				req = &types.QueryNextSequenceReceiveRequest{
					PortID:    channelA.PortID,
					ChannelID: channelA.ID,
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
