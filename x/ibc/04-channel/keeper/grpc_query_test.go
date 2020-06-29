package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

func (suite *KeeperTestSuite) TestQueryChannel() {
	var (
		req        *types.QueryChannelRequest
		expChannel types.Channel
		found      bool
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
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
			"sucess",
			func() {
				_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
				channelA := connA.NextTestChannel()
				// init channel
				channelA, _, err := suite.coordinator.ChanOpenInit(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
				suite.Require().NoError(err)

				expChannel, found = suite.chainA.App.IBCKeeper.ChannelKeeper.GetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ID)
				suite.Require().True(found)

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
			"sucess",
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
			"sucess",
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
