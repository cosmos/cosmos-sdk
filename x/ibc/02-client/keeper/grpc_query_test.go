package keeper_test

import (
	"fmt"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

func (suite *KeeperTestSuite) TestQueryClientState() {
	var (
		req            *types.QueryClientStateRequest
		expClientState *codectypes.Any
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"invalid clientID",
			func() {
				req = &types.QueryClientStateRequest{}
			},
			false,
		},
		{"client not found",
			func() {
				req = &types.QueryClientStateRequest{
					ClientId: testClientID,
				}
			},
			false,
		},
		{
			"success",
			func() {
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false)
				suite.keeper.SetClientState(suite.ctx, testClientID, clientState)

				var err error
				expClientState, err = types.PackClientState(clientState)
				suite.Require().NoError(err)

				req = &types.QueryClientStateRequest{
					ClientId: testClientID,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.ClientState(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				expClientState.ClearCachedValue()
				suite.Require().Equal(expClientState, res.ClientState)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryClientStates() {
	var (
		req             *types.QueryClientStatesRequest
		expClientStates = []*types.IdentifiedClientState(nil)
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty pagination",
			func() {
				req = &types.QueryClientStatesRequest{}
			},
			true,
		},
		{
			"success, no results",
			func() {
				req = &types.QueryClientStatesRequest{
					Pagination: &query.PageRequest{
						Limit:      3,
						CountTotal: true,
					},
				}
			},
			true,
		},
		{
			"success",
			func() {
				clientA1, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				clientA2, _ := suite.coordinator.CreateClient(suite.chainA, suite.chainB, exported.Tendermint)

				clientStateA1 := suite.chainA.GetClientState(clientA1)
				clientStateA2 := suite.chainA.GetClientState(clientA2)

				idcs := types.NewIdentifiedClientState(clientA1, clientStateA1)
				idcs2 := types.NewIdentifiedClientState(clientA2, clientStateA2)

				// order is sorted by client id, localhost is last
				expClientStates = []*types.IdentifiedClientState{&idcs, &idcs2}
				req = &types.QueryClientStatesRequest{
					Pagination: &query.PageRequest{
						Limit:      7,
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
			expClientStates = nil

			tc.malleate()

			// always add localhost which is created by default in init genesis
			localhostClientState := suite.chainA.GetClientState(exported.Localhost.String())
			identifiedLocalhost := types.NewIdentifiedClientState(exported.Localhost.String(), localhostClientState)
			expClientStates = append(expClientStates, &identifiedLocalhost)

			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.ClientStates(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(len(expClientStates), len(res.ClientStates))
				for i := range expClientStates {
					suite.Require().Equal(expClientStates[i].ClientId, res.ClientStates[i].ClientId)
					suite.Require().NotNil(res.ClientStates[i].ClientState)
					suite.Require().Equal(expClientStates[i].ClientState, res.ClientStates[i].ClientState)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryConsensusState() {
	var (
		req               *types.QueryConsensusStateRequest
		expConsensusState *codectypes.Any
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"invalid clientID",
			func() {
				req = &types.QueryConsensusStateRequest{}
			},
			false,
		},
		{
			"invalid height",
			func() {
				req = &types.QueryConsensusStateRequest{
					ClientId:     testClientID,
					EpochNumber:  0,
					EpochHeight:  0,
					LatestHeight: false,
				}
			},
			false,
		},
		{
			"consensus state not found",
			func() {
				req = &types.QueryConsensusStateRequest{
					ClientId:     testClientID,
					LatestHeight: true,
				}
			},
			false,
		},
		{
			"success latest height",
			func() {
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				cs := ibctmtypes.NewConsensusState(
					suite.consensusState.Timestamp, commitmenttypes.NewMerkleRoot([]byte("hash1")), suite.consensusState.Height, nil,
				)
				suite.keeper.SetClientState(suite.ctx, testClientID, clientState)
				suite.keeper.SetClientConsensusState(suite.ctx, testClientID, suite.consensusState.GetHeight(), cs)

				var err error
				expConsensusState, err = types.PackConsensusState(cs)
				suite.Require().NoError(err)

				req = &types.QueryConsensusStateRequest{
					ClientId:     testClientID,
					LatestHeight: true,
				}
			},
			true,
		},
		{
			"success with height",
			func() {
				cs := ibctmtypes.NewConsensusState(
					suite.consensusState.Timestamp, commitmenttypes.NewMerkleRoot([]byte("hash1")), suite.consensusState.Height, nil,
				)
				suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight, cs)

				var err error
				expConsensusState, err = types.PackConsensusState(cs)
				suite.Require().NoError(err)

				req = &types.QueryConsensusStateRequest{
					ClientId:    testClientID,
					EpochNumber: 0,
					EpochHeight: height,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.ConsensusState(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				expConsensusState.ClearCachedValue()
				suite.Require().Equal(expConsensusState, res.ConsensusState)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryConsensusStates() {
	var (
		req                *types.QueryConsensusStatesRequest
		expConsensusStates = []*codectypes.Any(nil)
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"invalid client identifier",
			func() {
				req = &types.QueryConsensusStatesRequest{}
			},
			false,
		},
		{
			"empty pagination",
			func() {
				req = &types.QueryConsensusStatesRequest{
					ClientId: testClientID,
				}
			},
			true,
		},
		{
			"success, no results",
			func() {
				req = &types.QueryConsensusStatesRequest{
					ClientId: testClientID,
					Pagination: &query.PageRequest{
						Limit:      3,
						CountTotal: true,
					},
				}
			},
			true,
		},
		{
			"success",
			func() {
				cs := ibctmtypes.NewConsensusState(
					suite.consensusState.Timestamp, commitmenttypes.NewMerkleRoot([]byte("hash1")), suite.consensusState.Height, nil,
				)
				cs2 := ibctmtypes.NewConsensusState(
					suite.consensusState.Timestamp.Add(time.Second), commitmenttypes.NewMerkleRoot([]byte("hash2")), suite.consensusState.Height, nil,
				)

				suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight, cs)
				suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight.Increment(), cs2)

				any, err := types.PackConsensusState(cs)
				suite.Require().NoError(err)
				any2, err := types.PackConsensusState(cs2)
				suite.Require().NoError(err)

				// order is swapped because the res is sorted by client id
				expConsensusStates = []*codectypes.Any{any, any2}
				req = &types.QueryConsensusStatesRequest{
					ClientId: testClientID,
					Pagination: &query.PageRequest{
						Limit:      3,
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
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.ConsensusStates(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(len(expConsensusStates), len(res.ConsensusStates))
				for i := range expConsensusStates {
					suite.Require().NotNil(res.ConsensusStates[i])
					expConsensusStates[i].ClearCachedValue()
					suite.Require().Equal(expConsensusStates[i], res.ConsensusStates[i])
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
